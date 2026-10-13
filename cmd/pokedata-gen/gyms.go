package main

import (
	"fmt"
	"sort"
	"strings"
)

// Nombres antiguos (gen 1-3) o variantes de Serebii -> forma normalizada de PokeAPI.
var moveAlias = map[string]string{
	"psychicm":     "psychic",
	"hijumpkick":   "highjumpkick",
	"faintattack":  "feintattack",
	"smellingsalt": "smellingsalts",
	"vicegrip":     "visegrip",
}

var itemAlias = map[string]string{}

var speciesAlias = map[string]string{}

// resolver acumula los nombres que no se pudieron casar para informar de todos a la vez.
type resolver struct {
	L      *Lookups
	errors []string
}

func (r *resolver) fail(format string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, a...))
}

func lookup(m map[string]int, alias map[string]string, name string) (int, bool) {
	k := norm(name)
	if a, ok := alias[k]; ok {
		k = a
	}
	id, ok := m[k]
	return id, ok
}

func (r *resolver) move(name, ctx string) int {
	id, ok := lookup(r.L.moveByNorm, moveAlias, name)
	if !ok {
		r.fail("%s: movimiento %q", ctx, name)
	}
	return id
}

func (r *resolver) item(name, ctx string) any {
	if name == "" {
		return nil
	}
	id, ok := lookup(r.L.itemByNorm, itemAlias, name)
	if !ok {
		r.fail("%s: objeto %q", ctx, name)
		return nil
	}
	return id
}

func (r *resolver) pokemon(m Mon, ctx string) (pokemonID int, gender string, gmax bool) {
	sp := m.SpeciesID
	if sp == 0 {
		id, ok := lookup(r.L.speciesByNorm, speciesAlias, m.Species)
		if !ok {
			r.fail("%s: especie %q", ctx, m.Species)
			return 0, "", false
		}
		sp = id
	}
	pid := r.L.defaultPokemon[sp]
	if m.Form > 0 {
		forms := r.L.formsBySpecies[sp]
		if m.Form < len(forms) {
			pid = forms[m.Form]
		} else {
			r.fail("%s: forma %d de la especie %d", ctx, m.Form, sp)
		}
	}
	return r.applyFormSuffix(pid, r.L.speciesIdent[sp], m.FormSuffix, m.Gender, ctx)
}

// Sufijos de Serebii de formas regionales: sufijo -> (forma en PokeAPI, nombre en el error).
var regionalSuffix = map[string][2]string{
	"g": {"galar", "de Galar"},
	"a": {"alola", "de Alola"},
	"h": {"hisui", "de Hisui"},
	"l": {"low-key", "low-key"},
}

// applyFormSuffix ajusta el Pokémon y el sexo según el sufijo de forma de Serebii.
func (r *resolver) applyFormSuffix(pid int, species, suffix, gender, ctx string) (int, string, bool) {
	if rs, ok := regionalSuffix[suffix]; ok {
		if id, ok := r.formByIdent(species, rs[0]); ok {
			pid = id
		} else {
			r.fail("%s: forma %s de %s", ctx, rs[1], species)
		}
		return pid, gender, false
	}
	switch suffix {
	case "":
	case "gi":
		return pid, gender, true
	case "f", "m":
		form, g := "female", "F"
		if suffix == "m" {
			form, g = "male", "M"
		}
		if id, ok := r.formByIdent(species, form); ok {
			pid = id
		}
		if gender == "" {
			gender = g
		}
	default:
		r.fail("%s: sufijo de forma %q de %s", ctx, suffix, species)
	}
	return pid, gender, false
}

// formByIdent busca el Pokémon "<especie>-<forma>" o, si no existe, el de menor
// id que empiece por "<especie>-<forma>-" (p. ej. darmanitan-galar-standard).
func (r *resolver) formByIdent(species, form string) (int, bool) {
	if id, ok := r.L.pokemonByIdent[species+"-"+form]; ok {
		return id, true
	}
	best := 0
	for ident, id := range r.L.pokemonByIdent {
		if strings.HasPrefix(ident, species+"-"+form+"-") && (best == 0 || id < best) {
			best = id
		}
	}
	return best, best != 0
}

// defaultMoves reproduce lo que hacen los juegos cuando el entrenador no
// define moveset: los 4 últimos movimientos aprendidos por nivel.
func (r *resolver) defaultMoves(pokemonID, vg, level int, ctx string) []int {
	lst := r.L.levelup[[2]int{pokemonID, vg}]
	if lst == nil {
		r.fail("%s: sin movimientos por nivel para pokemon %d en vg %d", ctx, pokemonID, vg)
		return nil
	}
	lst = append([]lvMove(nil), lst...)
	sort.SliceStable(lst, func(i, j int) bool {
		if lst[i].level != lst[j].level {
			return lst[i].level < lst[j].level
		}
		return lst[i].order < lst[j].order
	})
	var cur []int
	for _, m := range lst {
		if m.level > level {
			break
		}
		dup := false
		for _, c := range cur {
			if c == m.move {
				dup = true
			}
		}
		if dup {
			continue
		}
		cur = append(cur, m.move)
		if len(cur) > 4 {
			cur = cur[1:]
		}
	}
	return cur
}

// abilitiesIn devuelve las habilidades no ocultas posibles del Pokémon en una generación.
func (r *resolver) abilitiesIn(pokemonID, gen int) []int {
	bySlot := map[int]int{}
	for _, a := range r.L.abilitiesNow[pokemonID] {
		if !a.hidden {
			bySlot[a.slot] = a.ability
		}
	}
	for slot, a := range r.pastAbilities(pokemonID, gen) {
		if a.ability == 0 {
			delete(bySlot, slot)
		} else {
			bySlot[slot] = a.ability
		}
	}
	var slots []int
	for s := range bySlot {
		slots = append(slots, s)
	}
	sort.Ints(slots)
	var out []int
	seen := map[int]bool{}
	for _, s := range slots {
		a := bySlot[s]
		if seen[a] || r.L.abilityGen[a] > gen {
			continue
		}
		seen[a] = true
		out = append(out, a)
	}
	return out
}

// pastAbilities: para cada hueco, el registro "past" con la generación mínima >= gen manda.
func (r *resolver) pastAbilities(pokemonID, gen int) map[int]abilitySlot {
	best := map[int]abilitySlot{}
	for _, a := range r.L.abilitiesPast[pokemonID] {
		if a.gen < gen || a.hidden {
			continue
		}
		if b, ok := best[a.slot]; !ok || a.gen < b.gen {
			best[a.slot] = a
		}
	}
	return best
}

// ---------------------------------------------------------------------

func buildGyms(ds *Dataset, L *Lookups, srcDir string, logf func(string, ...any)) error {
	gb := &gymBuilder{r: &resolver{L: L}, L: L, covered: map[string]bool{}}
	gb.addLeaders(ds)
	gb.addGyms(ds)

	battles := append(pretBattles(srcDir), serebiiBattles(srcDir)...)
	// Orden cronológico por juego, estable dentro de cada juego
	rank := map[string]int{}
	for i, g := range gameOrder {
		rank[g] = i
	}
	sort.SliceStable(battles, func(i, j int) bool { return rank[battles[i].VG] < rank[battles[j].VG] })

	gb.tb = ds.T("gym_battles", "id", "gym_id", "gym_leader_id", "version_group_id", "version_id", "gym_order",
		"badge_en", "specialty_type_id", "battle_kind", "rematch_number", "is_double_battle", "notes", "source", "source_ref")
	gb.ti = ds.T("gym_battle_items", "gym_battle_id", "slot", "item_id")
	gb.tp = ds.T("gym_battle_pokemon", "id", "gym_battle_id", "slot", "pokemon_id", "level", "gender", "held_item_id",
		"ability_id", "ability_source", "moves_source", "tera_type_id", "can_dynamax", "is_gigantamax", "difficulty")
	gb.tm = ds.T("gym_battle_pokemon_moves", "gym_battle_pokemon_id", "slot", "move_id")

	for bi, b := range battles {
		gb.addBattle(bi+1, b)
	}
	gb.checkCoverage()

	r := gb.r
	if len(r.errors) > 0 {
		sort.Strings(r.errors)
		return fmt.Errorf("%d problemas al resolver líderes:\n  %s", len(r.errors), strings.Join(r.errors, "\n  "))
	}
	logf("gimnasios: %d líderes, %d combates, %d Pokémon", len(leaderDefs), len(battles), gb.monID)
	return nil
}

// gymBuilder rellena las tablas de gimnasios, líderes y combates.
type gymBuilder struct {
	r              *resolver
	L              *Lookups
	tb, ti, tp, tm *Table
	leaderID       map[string]int
	gymID          map[string]int
	gymBadge       map[string]string
	covered        map[string]bool
	monID          int
}

func (gb *gymBuilder) addLeaders(ds *Dataset) {
	tl := ds.T("gym_leaders", "id", "identifier", "name_en", "name_es", "gender")
	gb.leaderID = map[string]int{}
	for i, l := range leaderDefs {
		gb.leaderID[l.id] = i + 1
		tl.Add(i+1, l.id, l.en, l.es, l.gender)
	}
}

func (gb *gymBuilder) addGyms(ds *Dataset) {
	tg := ds.T("gyms", "id", "identifier", "city_en", "location_id", "region_id")
	gb.gymID = map[string]int{}
	gb.gymBadge = map[string]string{}
	for i, g := range gymDefs {
		gb.gymID[g.id] = i + 1
		gb.gymBadge[g.id] = g.badge
		var loc any
		if g.location != "" {
			id, ok := gb.L.locationByID[g.location]
			if !ok {
				gb.r.fail("gimnasio %s: location %q", g.id, g.location)
			} else {
				loc = id
			}
		}
		reg, ok := gb.L.regionByIdent[g.region]
		if !ok {
			gb.r.fail("gimnasio %s: región %q", g.id, g.region)
		}
		tg.Add(i+1, g.id, g.city, loc, reg)
	}
}

// findSlot devuelve la plaza del líder en el juego (la última si hay varias).
func findSlot(vg, leader string) *slotDef {
	var slot *slotDef
	for i := range gameSlots[vg] {
		if gameSlots[vg][i].leader == leader {
			slot = &gameSlots[vg][i]
		}
	}
	return slot
}

func (gb *gymBuilder) addBattle(bid int, b Battle) {
	L := gb.L
	ctx := fmt.Sprintf("%s/%s(%s)", b.VG, b.Leader, b.Ref)
	slot := findSlot(b.VG, b.Leader)
	if slot == nil {
		gb.r.fail("%s: líder sin plaza definida en ese juego", ctx)
		return
	}
	gb.covered[b.VG+"/"+b.Leader] = true
	vg := L.vgByIdent[b.VG]
	var version, typ any
	if slot.version != "" {
		version = L.versionByIdent[slot.version]
	}
	if slot.typ != "" {
		typ = L.typeByIdent[slot.typ]
	}
	badge := slot.badge
	if badge == "" {
		badge = gb.gymBadge[slot.gym]
	}
	notes := strings.Trim(strings.Join([]string{slot.notes, b.Notes}, "; "), "; ")
	var rematch any
	if b.Rematch > 0 {
		rematch = b.Rematch
	}
	double := b.Double || forcedDouble[b.VG+"/"+b.Leader]
	gb.tb.Add(bid, gb.gymID[slot.gym], gb.leaderID[b.Leader], vg, version, slot.order, badge, typ, b.Kind, rematch,
		boolInt(double), orNil(notes), b.Source, b.Ref)

	for i, it := range b.Items {
		if id := gb.r.item(it, ctx); id != nil {
			gb.ti.Add(bid, i+1, id)
		}
	}
	for si, m := range b.Mons {
		gb.monID++
		gb.addMon(bid, si+1, m, b.VG, vg, fmt.Sprintf("%s #%d", ctx, si+1))
	}
}

func (gb *gymBuilder) addMon(bid, slot int, m Mon, vgIdent string, vg int, mctx string) {
	r := gb.r
	pid, gender, gmax := r.pokemon(m, mctx)
	if pid == 0 {
		return
	}
	gen := gb.L.vgGeneration[vg]
	moves, src := r.monMoves(m, pid, vg, mctx)
	ability, abilitySrc := r.monAbility(m, pid, gen, vgIdent, mctx)
	tera := r.teraType(m.Tera, mctx)
	gmax = gmax || m.Gmax
	gb.tp.Add(gb.monID, bid, slot, pid, m.Level, orNil(gender), r.item(m.Item, mctx), ability, abilitySrc, src,
		tera, boolInt(m.Dynamax || gmax), boolInt(gmax), m.Difficulty)
	for k, mv := range moves {
		if mv != 0 {
			gb.tm.Add(gb.monID, k+1, mv)
		}
	}
}

// monMoves devuelve los movimientos del Pokémon y de dónde salen.
func (r *resolver) monMoves(m Mon, pid, vg int, mctx string) ([]int, string) {
	var moves []int
	src := "game_data"
	if m.Moves != nil {
		for _, mv := range m.Moves {
			moves = append(moves, r.move(mv, mctx))
		}
	} else {
		moves = r.defaultMoves(pid, vg, m.Level, mctx)
		src = "default_levelup"
		if len(m.MoveOverride) > 0 {
			src = "mixed"
			moves = r.overrideMoves(moves, m.MoveOverride, mctx)
		}
	}
	if len(moves) > 4 {
		r.fail("%s: más de 4 movimientos", mctx)
		moves = moves[:4]
	}
	return moves, src
}

// overrideMoves sustituye huecos concretos de los movimientos por defecto.
func (r *resolver) overrideMoves(moves []int, override map[int]string, mctx string) []int {
	for s, mv := range override {
		id := r.move(mv, mctx)
		for len(moves) <= s {
			moves = append(moves, 0)
		}
		moves[s] = id
	}
	// huecos intermedios vacíos (si el Pokémon sabía menos movimientos)
	var compact []int
	for _, x := range moves {
		if x != 0 {
			compact = append(compact, x)
		}
	}
	return compact
}

// monAbility devuelve la habilidad del Pokémon y de dónde sale (nil si no se sabe).
func (r *resolver) monAbility(m Mon, pid, gen int, vgIdent, mctx string) (ability, src any) {
	switch {
	case gen < 3 || vgIdent == "lets-go-pikachu-lets-go-eevee":
		// sin habilidades
	case m.Ability != "":
		id, ok := lookup(r.L.abilityByNorm, nil, m.Ability)
		if !ok {
			r.fail("%s: habilidad %q", mctx, m.Ability)
		} else {
			return id, "game_data"
		}
	default:
		opts := r.abilitiesIn(pid, gen)
		switch {
		case m.AbilitySlot > 0 && m.AbilitySlot <= len(opts):
			return opts[m.AbilitySlot-1], "game_data"
		case len(opts) == 1:
			return opts[0], "only_option"
		}
	}
	return nil, nil
}

func (r *resolver) teraType(ident, mctx string) any {
	if ident == "" {
		return nil
	}
	id, ok := r.L.typeByIdent[ident]
	if !ok {
		r.fail("%s: teratipo %q", mctx, ident)
		return nil
	}
	return id
}

// checkCoverage comprueba que toda plaza definida tiene al menos un combate.
func (gb *gymBuilder) checkCoverage() {
	for vg, slots := range gameSlots {
		for _, s := range slots {
			if !gb.covered[vg+"/"+s.leader] {
				gb.r.fail("%s/%s: no se encontró el equipo en las fuentes", vg, s.leader)
			}
		}
	}
}

// Combates dobles que la fuente no marca como tales.
var forcedDouble = map[string]bool{
	"scarlet-violet/ryme": true,
	"sword-shield/raihan": true,
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
