package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Mon es un Pokémon de un equipo tal como viene de la fuente (sin resolver a IDs).
type Mon struct {
	Species      string // nombre/constante de especie
	SpeciesID    int    // si la fuente da el nº nacional
	Form         int    // índice de forma (pret gen 4)
	FormSuffix   string // sufijo de imagen de Serebii (g, gi, f, l...)
	Level        int
	Gender       string
	Item         string
	Ability      string
	AbilitySlot  int            // 2 = la fuente fuerza la segunda habilidad (HGSS)
	Moves        []string       // nil = movimientos por defecto
	MoveOverride map[int]string // gen 1: hueco (0-based) -> movimiento especial
	Tera         string
	Dynamax      bool
	Gmax         bool
	Difficulty   any
}

// Battle es un combate contra un líder tal como viene de la fuente.
type Battle struct {
	VG      string
	Leader  string
	Kind    string // gym, rematch, challenge_mode, special
	Rematch int
	Double  bool
	Items   []string
	Mons    []Mon
	Source  string
	Ref     string
	Notes   string
}

func readText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func stripComment(l string) string {
	if i := strings.IndexByte(l, ';'); i >= 0 {
		l = l[:i]
	}
	return strings.TrimSpace(l)
}

func splitArgs(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ---------------------------------------------------------------------
// Gen 1: pokered / pokeyellow
// ---------------------------------------------------------------------

// parseGen1Parties: etiqueta "BrockData" -> lista de equipos (cada uno una línea db)
func parseGen1Parties(path string) map[string][][]Mon {
	out := map[string][][]Mon{}
	cur := ""
	for _, raw := range strings.Split(readText(path), "\n") {
		l := stripComment(raw)
		if strings.HasSuffix(l, ":") && !strings.HasPrefix(l, "db") {
			cur = strings.TrimSuffix(l, ":")
			continue
		}
		if cur == "" || !strings.HasPrefix(l, "db ") {
			continue
		}
		a := splitArgs(strings.TrimPrefix(l, "db "))
		var party []Mon
		if a[0] == "$FF" {
			for i := 1; i+1 < len(a); i += 2 {
				party = append(party, Mon{Level: atoi(a[i]), Species: a[i+1]})
			}
		} else {
			lv := atoi(a[0])
			for _, sp := range a[1:] {
				if sp == "0" {
					break
				}
				party = append(party, Mon{Level: lv, Species: sp})
			}
		}
		out[cur] = append(out[cur], party)
	}
	return out
}

var gen1Leaders = []struct {
	leader, label, class string
	party                int
}{
	{"brock", "BrockData", "BROCK", 1},
	{"misty", "MistyData", "MISTY", 1},
	{"lt-surge", "LtSurgeData", "LT_SURGE", 1},
	{"erika", "ErikaData", "ERIKA", 1},
	{"koga", "KogaData", "KOGA", 1},
	{"sabrina", "SabrinaData", "SABRINA", 1},
	{"blaine", "BlaineData", "BLAINE", 1},
	{"giovanni", "GiovanniData", "GIOVANNI", 3},
}

// Rojo/Azul: "LoneMoves" da un movimiento especial al Pokémon N del líder N (hueco 3).
func parseRBLoneMoves(path string) map[int][2]string {
	out := map[int][2]string{}
	in := false
	n := 0
	for _, raw := range strings.Split(readText(path), "\n") {
		l := stripComment(raw)
		switch {
		case l == "LoneMoves:":
			in = true
		case strings.HasSuffix(l, ":"):
			in = false
		case in && strings.HasPrefix(l, "db "):
			a := splitArgs(strings.TrimPrefix(l, "db "))
			n++
			out[n] = [2]string{a[0], a[1]}
		}
	}
	return out
}

// Amarillo: "SpecialTrainerMoves" -> (clase, nº equipo) -> lista (pokémon 1-based, hueco 1-based, movimiento)
func parseYellowSpecialMoves(path string) map[string][][3]string {
	out := map[string][][3]string{}
	in := false
	key := ""
	for _, raw := range strings.Split(readText(path), "\n") {
		l := stripComment(raw)
		if l == "SpecialTrainerMoves:" {
			in = true
			continue
		}
		if !in || !strings.HasPrefix(l, "db ") {
			continue
		}
		a := splitArgs(strings.TrimPrefix(l, "db "))
		switch {
		case len(a) == 1 && a[0] == "-1":
			return out
		case len(a) == 1 && a[0] == "0":
			key = ""
		case len(a) == 2:
			key = a[0] + "/" + a[1]
		case len(a) == 3 && key != "":
			out[key] = append(out[key], [3]string{a[0], a[1], a[2]})
		}
	}
	return out
}

func gen1Battles(dir, repo, vg string) []Battle {
	parties := parseGen1Parties(filepath.Join(dir, "pret", repo, "data/trainers/parties.asm"))
	special := filepath.Join(dir, "pret", repo, "data/trainers/special_moves.asm")
	var lone map[int][2]string
	var yellow map[string][][3]string
	if repo == "pokered" {
		lone = parseRBLoneMoves(special)
	} else {
		yellow = parseYellowSpecialMoves(special)
	}
	var out []Battle
	for i, g := range gen1Leaders {
		ps := parties[g.label]
		if len(ps) < g.party {
			panic(fmt.Sprintf("%s: no hay equipo %d en %s", repo, g.party, g.label))
		}
		mons := append([]Mon(nil), ps[g.party-1]...)
		for j := range mons {
			mons[j].MoveOverride = map[int]string{}
		}
		if lone != nil {
			lm := lone[i+1]
			idx := atoi(lm[0])
			mons[idx].MoveOverride[2] = lm[1]
		}
		for _, sm := range yellow[fmt.Sprintf("%s/%d", g.class, g.party)] {
			mons[atoi(sm[0])-1].MoveOverride[atoi(sm[1])-1] = sm[2]
		}
		out = append(out, Battle{VG: vg, Leader: g.leader, Kind: "gym", Mons: mons,
			Source: "pret/" + repo, Ref: fmt.Sprintf("%s #%d", g.label, g.party)})
	}
	return out
}

// ---------------------------------------------------------------------
// Gen 2: pokegold / pokecrystal
// ---------------------------------------------------------------------

var gen2Groups = map[string]string{
	"FalknerGroup": "falkner", "BugsyGroup": "bugsy", "WhitneyGroup": "whitney", "MortyGroup": "morty",
	"ChuckGroup": "chuck", "JasmineGroup": "jasmine", "PryceGroup": "pryce", "ClairGroup": "clair",
	"BrockGroup": "brock", "MistyGroup": "misty", "LtSurgeGroup": "lt-surge", "ErikaGroup": "erika",
	"JanineGroup": "janine", "SabrinaGroup": "sabrina", "BlaineGroup": "blaine", "BlueGroup": "blue",
}

func gen2Battles(dir, repo, vg string) []Battle {
	var out []Battle
	group, ttype := "", ""
	partyN := 0
	var cur *Battle
	for _, raw := range strings.Split(readText(filepath.Join(dir, "pret", repo, "data/trainers/parties.asm")), "\n") {
		l := stripComment(raw)
		if strings.HasSuffix(l, "Group:") {
			group, partyN, cur = strings.TrimSuffix(l, ":"), 0, nil
			continue
		}
		leader, ok := gen2Groups[group]
		if !ok || !strings.HasPrefix(l, "db ") {
			continue
		}
		body := strings.TrimPrefix(l, "db ")
		if strings.HasPrefix(body, `"`) { // nuevo equipo: db "NAME@", TRAINERTYPE_X
			partyN++
			a := splitArgs(body)
			ttype = a[len(a)-1]
			cur = nil
			if partyN == 1 {
				out = append(out, Battle{VG: vg, Leader: leader, Kind: "gym", Source: "pret/" + repo,
					Ref: fmt.Sprintf("%s #%d", group, partyN)})
				cur = &out[len(out)-1]
			}
			continue
		}
		if cur == nil || body == "-1" {
			continue
		}
		a := splitArgs(body)
		m := Mon{Level: atoi(a[0]), Species: a[1]}
		rest := a[2:]
		if ttype == "TRAINERTYPE_ITEM" || ttype == "TRAINERTYPE_ITEM_MOVES" {
			if rest[0] != "NO_ITEM" {
				m.Item = rest[0]
			}
			rest = rest[1:]
		}
		if ttype == "TRAINERTYPE_MOVES" || ttype == "TRAINERTYPE_ITEM_MOVES" {
			m.Moves = []string{}
			for _, mv := range rest {
				if mv != "NO_MOVE" {
					m.Moves = append(m.Moves, mv)
				}
			}
		}
		cur.Mons = append(cur.Mons, m)
	}
	return out
}

// ---------------------------------------------------------------------
// Gen 3: pokeruby / pokeemerald / pokefirered (C)
// ---------------------------------------------------------------------

var (
	gen3TrainerRe = regexp.MustCompile(`(?s)\[(TRAINER_\w+)\]\s*=\s*\{(.*?)\n\s*\},`)
	gen3ClassRe   = regexp.MustCompile(`\.trainerClass\s*=\s*(\w+)`)
	gen3NameRe    = regexp.MustCompile(`\.trainerName\s*=\s*_\("([^"]*)"\)`)
	gen3ItemsRe   = regexp.MustCompile(`\.items\s*=\s*\{([^}]*)\}`)
	gen3DoubleRe  = regexp.MustCompile(`\.doubleBattle\s*=\s*(\w+)`)
	gen3PartyRe   = regexp.MustCompile(`\.party\s*=.*?\b((?:sParty|gTrainerParty)_\w+)`)
	gen3ArrRe     = regexp.MustCompile(`(?s)((?:sParty|gTrainerParty)_\w+)\[\]\s*=\s*\{(.*?)\n\};`)
	gen3FieldRe   = regexp.MustCompile(`\.(iv|lvl|level|species|heldItem|moves)\s*=\s*([^\n]*)`)
	gen3SuffixRe  = regexp.MustCompile(`_(\d)$`)
)

var gen3LeaderNames = map[string]string{
	"roxanne": "roxanne", "brawly": "brawly", "wattson": "wattson", "flannery": "flannery",
	"norman": "norman", "winona": "winona", "tateliza": "tate-and-liza", "wallace": "wallace", "juan": "juan",
	"brock": "brock", "misty": "misty", "ltsurge": "lt-surge", "erika": "erika", "koga": "koga",
	"sabrina": "sabrina", "blaine": "blaine", "giovanni": "giovanni",
}

func gen3Battles(dir, repo, trainersFile, vg string) []Battle {
	parties := map[string][]Mon{}
	for _, m := range gen3ArrRe.FindAllStringSubmatch(readText(filepath.Join(dir, "pret", repo, "src/data/trainer_parties.h")), -1) {
		var mons []Mon
		for _, f := range gen3FieldRe.FindAllStringSubmatch(m[2], -1) {
			v := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(f[2]), ","))
			if f[1] == "iv" {
				mons = append(mons, Mon{Difficulty: int64(atoi(v))})
				continue
			}
			cur := &mons[len(mons)-1]
			switch f[1] {
			case "lvl", "level":
				cur.Level = atoi(v)
			case "species":
				cur.Species = strings.TrimPrefix(v, "SPECIES_")
			case "heldItem":
				if v != "ITEM_NONE" {
					cur.Item = strings.TrimPrefix(v, "ITEM_")
				}
			case "moves":
				cur.Moves = []string{}
				for _, mv := range splitArgs(strings.Trim(v, "{}")) {
					if mv != "MOVE_NONE" {
						cur.Moves = append(cur.Moves, strings.TrimPrefix(mv, "MOVE_"))
					}
				}
			}
		}
		parties[m[1]] = mons
	}

	var out []Battle
	for _, m := range gen3TrainerRe.FindAllStringSubmatch(readText(filepath.Join(dir, "pret", repo, "src/data", trainersFile)), -1) {
		key, body := m[1], m[2]
		cls := gen3ClassRe.FindStringSubmatch(body)
		if cls == nil || cls[1] != "TRAINER_CLASS_LEADER" {
			continue
		}
		name := gen3NameRe.FindStringSubmatch(body)[1]
		leader, ok := gen3LeaderNames[norm(name)]
		if !ok {
			panic(fmt.Sprintf("%s: líder desconocido %q (%s)", repo, name, key))
		}
		b := Battle{VG: vg, Leader: leader, Kind: "gym", Source: "pret/" + repo, Ref: key}
		if s := gen3SuffixRe.FindStringSubmatch(key); s != nil && s[1] != "1" {
			b.Kind, b.Rematch = "rematch", atoi(s[1])-1
		}
		if d := gen3DoubleRe.FindStringSubmatch(body); d != nil && d[1] == "TRUE" {
			b.Double = true
		}
		if it := gen3ItemsRe.FindStringSubmatch(body); it != nil {
			for _, x := range splitArgs(it[1]) {
				if x != "ITEM_NONE" {
					b.Items = append(b.Items, strings.TrimPrefix(x, "ITEM_"))
				}
			}
		}
		pm := gen3PartyRe.FindStringSubmatch(body)
		if pm == nil || parties[pm[1]] == nil {
			panic(fmt.Sprintf("%s: equipo no encontrado para %s", repo, key))
		}
		b.Mons = parties[pm[1]]
		out = append(out, b)
	}
	return out
}

// ---------------------------------------------------------------------
// Gen 4: pokediamond / pokeplatinum / pokeheartgold (JSON)
// ---------------------------------------------------------------------

type g4Mon struct {
	Species         string   `json:"species"`
	Form            int      `json:"form"`
	Level           int      `json:"level"`
	Item            *string  `json:"item"`
	Moves           []string `json:"moves"`
	Difficulty      *int     `json:"difficulty"`
	IVScale         *int     `json:"iv_scale"`
	AbilityOverride string   `json:"abilityOverride"`
	GenderOverride  string   `json:"genderOverride"`
}

type g4Trainer struct {
	Class        string   `json:"class"`
	Items        []string `json:"items"`
	Double       any      `json:"double"`
	DoubleBattle any      `json:"doubleBattle"`
	DoubleB      any      `json:"double_battle"`
	Party        []g4Mon  `json:"party"`
}

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	}
	return false
}

func g4Convert(t g4Trainer) ([]string, []Mon, bool) {
	var items []string
	for _, it := range t.Items {
		if it != "ITEM_NONE" {
			items = append(items, strings.TrimPrefix(it, "ITEM_"))
		}
	}
	var mons []Mon
	for _, p := range t.Party {
		m := Mon{Species: strings.TrimPrefix(p.Species, "SPECIES_"), Form: p.Form, Level: p.Level}
		if p.Item != nil && *p.Item != "ITEM_NONE" {
			m.Item = strings.TrimPrefix(*p.Item, "ITEM_")
		}
		if p.Moves != nil {
			m.Moves = []string{}
			for _, mv := range p.Moves {
				if mv != "MOVE_NONE" {
					m.Moves = append(m.Moves, strings.TrimPrefix(mv, "MOVE_"))
				}
			}
		}
		switch {
		case p.Difficulty != nil:
			m.Difficulty = int64(*p.Difficulty)
		case p.IVScale != nil:
			m.Difficulty = int64(*p.IVScale)
		}
		if p.AbilityOverride == "TRPOKE_ABILITY_OVERRIDE_SECOND" {
			m.AbilitySlot = 2
		} else if p.AbilityOverride == "TRPOKE_ABILITY_OVERRIDE_FIRST" {
			m.AbilitySlot = 1
		}
		switch p.GenderOverride {
		case "TRPOKE_GENDER_OVERRIDE_MALE":
			m.Gender = "M"
		case "TRPOKE_GENDER_OVERRIDE_FEMALE":
			m.Gender = "F"
		}
		mons = append(mons, m)
	}
	return items, mons, truthy(t.Double) || truthy(t.DoubleBattle) || truthy(t.DoubleB)
}

var gen4ClassLeader = map[string]string{
	"ROARK": "roark", "GARDENIA": "gardenia", "MAYLENE": "maylene", "WAKE": "crasher-wake",
	"FANTINA": "fantina", "BYRON": "byron", "CANDICE": "candice", "VOLKNER": "volkner",
	"FALKNER": "falkner", "BUGSY": "bugsy", "WHITNEY": "whitney", "MORTY": "morty", "CHUCK": "chuck",
	"JASMINE": "jasmine", "PRYCE": "pryce", "CLAIR": "clair", "BROCK": "brock", "MISTY": "misty",
	"LT_SURGE": "lt-surge", "ERIKA": "erika", "JANINE": "janine", "SABRINA": "sabrina",
	"BLAINE": "blaine", "BLUE": "blue",
}

func gen4LeaderFromClass(cls string) (string, bool) {
	for _, p := range []string{"TRAINER_CLASS_LEADER_", "TRAINERCLASS_LEADER_"} {
		if strings.HasPrefix(cls, p) {
			l, ok := gen4ClassLeader[strings.TrimPrefix(cls, p)]
			return l, ok
		}
	}
	return "", false
}

// Diamante/Perla y HGSS: un único JSON con todos los entrenadores.
// Por líder: 1ª aparición = gimnasio, 2ª = revancha; el resto se ignora.
func gen4ListBattles(path, repo, key, vg string) []Battle {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(readText(path)), &raw); err != nil {
		panic(err)
	}
	var list []g4Trainer
	if err := json.Unmarshal(raw[key], &list); err != nil {
		panic(err)
	}
	seen := map[string]int{}
	var out []Battle
	for i, t := range list {
		leader, ok := gen4LeaderFromClass(t.Class)
		if !ok || len(t.Party) == 0 {
			continue
		}
		seen[leader]++
		b := Battle{VG: vg, Leader: leader, Kind: "gym", Source: "pret/" + repo, Ref: fmt.Sprintf("trainer #%d", i)}
		switch seen[leader] {
		case 1:
		case 2:
			b.Kind, b.Rematch = "rematch", 1
		default:
			continue
		}
		b.Items, b.Mons, b.Double = g4Convert(t)
		out = append(out, b)
	}
	return out
}

// Platino: un JSON por entrenador.
func platinumBattles(dir, vg string) []Battle {
	files, _ := filepath.Glob(filepath.Join(dir, "pret", "pokeplatinum", "res/trainers/data", "leader_*.json"))
	sort.Strings(files)
	var out []Battle
	for _, f := range files {
		var t g4Trainer
		if err := json.Unmarshal([]byte(readText(f)), &t); err != nil {
			panic(err)
		}
		leader, ok := gen4LeaderFromClass(t.Class)
		if !ok {
			panic("platino: clase desconocida " + t.Class)
		}
		base := strings.TrimSuffix(filepath.Base(f), ".json")
		b := Battle{VG: vg, Leader: leader, Kind: "gym", Source: "pret/pokeplatinum", Ref: base}
		switch {
		case strings.HasSuffix(base, "_rematch"):
			b.Kind, b.Rematch = "rematch", 1
		case strings.HasSuffix(base, "_fight_area"):
			b.Kind, b.Notes = "special", "Combate en la Zona Combate (Fight Area)"
		}
		b.Items, b.Mons, b.Double = g4Convert(t)
		out = append(out, b)
	}
	return out
}

func pretBattles(dir string) []Battle {
	var out []Battle
	out = append(out, gen1Battles(dir, "pokered", "red-blue")...)
	out = append(out, gen1Battles(dir, "pokeyellow", "yellow")...)
	out = append(out, gen2Battles(dir, "pokegold", "gold-silver")...)
	out = append(out, gen2Battles(dir, "pokecrystal", "crystal")...)
	out = append(out, gen3Battles(dir, "pokeruby", "trainers_en.h", "ruby-sapphire")...)
	out = append(out, gen3Battles(dir, "pokeemerald", "trainers.h", "emerald")...)
	out = append(out, gen3Battles(dir, "pokefirered", "trainers.h", "firered-leafgreen")...)
	out = append(out, gen4ListBattles(filepath.Join(dir, "pret", "pokediamond", "files/poketool/trainer/trdata.json"),
		"pokediamond", "trdata", "diamond-pearl")...)
	out = append(out, platinumBattles(dir, "platinum")...)
	out = append(out, gen4ListBattles(filepath.Join(dir, "pret", "pokeheartgold", "files/poketool/trainer/trainers.json"),
		"pokeheartgold", "trainers", "heartgold-soulsilver")...)
	return out
}
