package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var serebiiGames = []struct{ page, vg string }{
	{"blackwhite", "black-white"},
	{"black2white2", "black-2-white-2"},
	{"xy", "x-y"},
	{"omegarubyalphasapphire", "omega-ruby-alpha-sapphire"},
	{"letsgopikachueevee", "lets-go-pikachu-lets-go-eevee"},
	{"swordshield", "sword-shield"},
	{"brilliantdiamondshiningpearl", "brilliant-diamond-shining-pearl"},
	{"scarletviolet", "scarlet-violet"},
}

// --- utilidades DOM ---

func attr(n *html.Node, k string) string {
	for _, a := range n.Attr {
		if a.Key == k {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, c string) bool {
	for _, x := range strings.Fields(attr(n, "class")) {
		if x == c {
			return true
		}
	}
	return false
}

func findAll(n *html.Node, pred func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if pred(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

func isElem(tag string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == tag }
}

// rows devuelve las filas propias de una tabla (no las de tablas anidadas).
func rows(table *html.Node) []*html.Node {
	var out []*html.Node
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		switch c.Data {
		case "tr":
			out = append(out, c)
		case "tbody", "thead":
			out = append(out, rows(c)...)
		}
	}
	return out
}

func cells(tr *html.Node) []*html.Node {
	var out []*html.Node
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			out = append(out, c)
		}
	}
	return out
}

func text(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(spaceRe.ReplaceAllString(b.String(), " "))
}

func links(n *html.Node, prefix string) []string {
	var out []string
	for _, a := range findAll(n, isElem("a")) {
		if h := attr(a, "href"); strings.Contains(h, prefix) {
			out = append(out, h)
		}
	}
	return out
}

// slug de "/attackdex-bw/workup.shtml" -> "workup"
func slug(href string) string {
	b := filepath.Base(href)
	return strings.TrimSuffix(b, ".shtml")
}

var (
	levelRe    = regexp.MustCompile(`Level (\d+)`)
	formImgRe  = regexp.MustCompile(`/(\d+)-([a-z]+)\.png$`)
	dexNumRe   = regexp.MustCompile(`^\d+$`)
	teraTypeRe = regexp.MustCompile(`Tera Type (\w+)`)
	leaderRe   = regexp.MustCompile(`^(?:Gym )?Leaders? (.+)$`)
)

var serebiiLeaderAlias = map[string]string{
	"ltsurge": leaderLtSurge, "lizatate": "tate-and-liza", "tateliza": "tate-and-liza",
	"wake": leaderCrasherWake, "crasherwake": leaderCrasherWake,
}

func serebiiLeader(name string) string {
	k := norm(name)
	if v, ok := serebiiLeaderAlias[k]; ok {
		return v
	}
	for _, l := range leaderDefs {
		if norm(l.en) == k {
			return l.id
		}
	}
	panic("serebii: líder desconocido " + name)
}

func serebiiBattles(dir string) []Battle {
	var out []Battle
	for _, g := range serebiiGames {
		f, err := os.Open(filepath.Join(dir, "serebii", g.page+".html"))
		if err != nil {
			panic(err)
		}
		doc, err := html.Parse(f) // Serebii sirve latin-1; los nombres útiles son ASCII salvo ♂/♀ (ver gender)
		f.Close()
		if err != nil {
			panic(err)
		}
		tables := findAll(doc, func(n *html.Node) bool { return isElem("table")(n) && hasClass(n, "trainer") })
		for i, t := range tables {
			b := parseSerebiiTable(t)
			b.VG = g.vg
			b.Source = "serebii/" + g.page
			b.Ref = fmt.Sprintf("table #%d", i+1)
			out = append(out, b)
		}
	}
	return out
}

func parseSerebiiTable(t *html.Node) Battle {
	rs := rows(t)
	b := Battle{Kind: "gym"}
	nameIdx := findSerebiiLeaderRow(rs, &b)
	if nameIdx < 1 {
		panic("serebii: tabla sin fila de nombres")
	}
	nameCells := cells(rs[nameIdx])[1:]
	imgCells := cells(rs[nameIdx-1])
	imgCells = imgCells[len(imgCells)-len(nameCells):]
	n := len(nameCells)
	b.Mons = make([]Mon, n)
	for j, c := range nameCells {
		b.Mons[j] = parseSerebiiMon(c, imgCells[j])
	}

	// Filas siguientes: nivel, habilidad/tipos, ataques, objeto, extras
	stage := 0
	for _, r := range rs[nameIdx+1:] {
		cs := cells(r)
		if len(cs) < n {
			continue
		}
		last := cs[len(cs)-n:]
		rowText := text(r)
		switch {
		case stage == 0 && levelRe.MatchString(rowText):
			serebiiLevelRow(&b, cs[0], last)
			stage = 1
		case stage == 1:
			forEachMon(b.Mons, last, serebiiAbility)
			stage = 2
		case strings.HasPrefix(rowText, "Attacks"):
			forEachMon(b.Mons, last, serebiiMoves)
		case strings.HasPrefix(rowText, "Hold Item"):
			forEachMon(b.Mons, last, serebiiItem)
		default:
			forEachMon(b.Mons, last, serebiiExtras)
		}
	}
	return b
}

// findSerebiiLeaderRow devuelve el índice de la fila con el nombre del líder
// (-1 si no la hay) y anota en b el líder y si es el Modo Desafío.
func findSerebiiLeaderRow(rs []*html.Node, b *Battle) int {
	for i, r := range rs {
		cs := cells(r)
		if len(cs) == 0 {
			continue
		}
		if len(cs) == 1 && strings.Contains(text(r), "Challenge Mode") {
			b.Kind, b.Notes = "challenge_mode", "Modo Desafío"
		}
		if m := leaderRe.FindStringSubmatch(text(cs[0])); m != nil && len(cs) > 1 {
			b.Leader = serebiiLeader(m[1])
			return i
		}
	}
	return -1
}

// parseSerebiiMon lee especie y sexo de la celda del nombre y la forma de la imagen.
func parseSerebiiMon(nameCell, imgCell *html.Node) Mon {
	var m Mon
	tx := text(nameCell)
	if strings.Contains(tx, "♂") {
		m.Gender = "M"
	} else if strings.Contains(tx, "♀") {
		m.Gender = "F"
	}
	m.Species = strings.TrimSpace(strings.NewReplacer("♂", "", "♀", "").Replace(tx))
	if ls := links(nameCell, "/pokedex"); len(ls) > 0 {
		s := slug(ls[0])
		if dexNumRe.MatchString(s) {
			m.SpeciesID = atoi(s)
		} else {
			m.Species = s
		}
	}
	for _, img := range findAll(imgCell, isElem("img")) {
		if fm := formImgRe.FindStringSubmatch(attr(img, "src")); fm != nil {
			m.FormSuffix = fm[2]
			if m.SpeciesID == 0 {
				m.SpeciesID = atoi(fm[1])
			}
		}
	}
	return m
}

func forEachMon(mons []Mon, cs []*html.Node, f func(*Mon, *html.Node)) {
	for j, c := range cs {
		f(&mons[j], c)
	}
}

// serebiiLevelRow: la primera celda trae combate doble y objetos del líder; el resto, niveles.
func serebiiLevelRow(b *Battle, first *html.Node, last []*html.Node) {
	b.Double = strings.Contains(text(first), "Double Battle")
	for _, h := range links(first, "/itemdex/") {
		b.Items = append(b.Items, slug(h))
	}
	forEachMon(b.Mons, last, func(m *Mon, c *html.Node) {
		if lm := levelRe.FindStringSubmatch(text(c)); lm != nil {
			m.Level = atoi(lm[1])
		}
	})
}

func serebiiAbility(m *Mon, c *html.Node) {
	if ls := links(c, "/abilitydex/"); len(ls) > 0 {
		m.Ability = slug(ls[0])
	}
}

func serebiiMoves(m *Mon, c *html.Node) {
	m.Moves = []string{}
	for _, h := range links(c, "/attackdex") {
		if s := slug(h); s != "" {
			m.Moves = append(m.Moves, s)
		}
	}
}

func serebiiItem(m *Mon, c *html.Node) {
	if ls := links(c, "/itemdex/"); len(ls) > 0 {
		m.Item = slug(ls[0])
		return
	}
	it := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text(c), "Hold Item"), " :"))
	if it != "" && it != "No Item" && it != "None" {
		m.Item = it
	}
}

func serebiiExtras(m *Mon, c *html.Node) {
	ct := text(c)
	if strings.Contains(ct, "Gigantamax") {
		m.Gmax, m.Dynamax = true, true
	} else if strings.Contains(ct, "Dynamax") {
		m.Dynamax = true
	}
	if tm := teraTypeRe.FindStringSubmatch(ct); tm != nil {
		m.Tera = strings.ToLower(tm[1])
	}
}
