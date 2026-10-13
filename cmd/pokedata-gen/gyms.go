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
	gender = m.Gender
	if m.Form > 0 {
		forms := r.L.formsBySpecies[sp]
		if m.Form < len(forms) {
			pid = forms[m.Form]
		} else {
			r.fail("%s: forma %d de la especie %d", ctx, m.Form, sp)
		}
	}
	species := r.L.speciesIdent[sp]
	try := func(suffix string) bool {
		if id, ok := r.L.pokemonByIdent[species+"-"+suffix]; ok {
			pid = id
			return true
		}
		// p. ej. darmanitan-galar-standard
		best := 0
		for ident, id := range r.L.pokemonByIdent {
			if strings.HasPrefix(ident, species+"-"+suffix+"-") && (best == 0 || id < best) {
				best = id
			}
		}
		if best != 0 {
			pid = best
		}
		return best != 0
	}
	switch m.FormSuffix {
	case "":
	case "gi":
		gmax = true
	case "g":
		if !try("galar") {
			r.fail("%s: forma de Galar de %s", ctx, species)
		}
	case "a":
		if !try("alola") {
			r.fail("%s: forma de Alola de %s", ctx, species)
		}
	case "h":
		if !try("hisui") {
			r.fail("%s: forma de Hisui de %s", ctx, species)
		}
	case "l":
		if !try("low-key") {
			r.fail("%s: forma low-key de %s", ctx, species)
		}
	case "f":
		try("female")
		if gender == "" {
			gender = "F"
		}
	case "m":
		try("male")
		if gender == "" {
			gender = "M"
		}
	default:
		r.fail("%s: sufijo de forma %q de %s", ctx, m.FormSuffix, species)
	}
	return pid, gender, gmax
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
	// Para cada hueco, el registro "past" con la generación mínima >= gen manda.
	best := map[int]abilitySlot{}
	for _, a := range r.L.abilitiesPast[pokemonID] {
		if a.gen < gen || a.hidden {
			continue
		}
		if b, ok := best[a.slot]; !ok || a.gen < b.gen {
			best[a.slot] = a
		}
	}
	for slot, a := range best {
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

// ---------------------------------------------------------------------

func buildGyms(ds *Dataset, L *Lookups, srcDir string, logf func(string, ...any)) error {
	r := &resolver{L: L}

	tl := ds.T("gym_leaders", "id", "identifier", "name_en", "name_es", "gender")
	leaderID := map[string]int{}
	for i, l := range leaderDefs {
		leaderID[l.id] = i + 1
		tl.Add(i+1, l.id, l.en, l.es, l.gender)
	}

	tg := ds.T("gyms", "id", "identifier", "city_en", "location_id", "region_id")
	gymID := map[string]int{}
	gymBadge := map[string]string{}
	for i, g := range gymDefs {
		gymID[g.id] = i + 1
		gymBadge[g.id] = g.badge
		var loc any
		if g.location != "" {
			id, ok := L.locationByID[g.location]
			if !ok {
				r.fail("gimnasio %s: location %q", g.id, g.location)
			} else {
				loc = id
			}
		}
		reg, ok := L.regionByIdent[g.region]
		if !ok {
			r.fail("gimnasio %s: región %q", g.id, g.region)
		}
		tg.Add(i+1, g.id, g.city, loc, reg)
	}

	battles := append(pretBattles(srcDir), serebiiBattles(srcDir)...)
	// Orden cronológico por juego, estable dentro de cada juego
	rank := map[string]int{}
	for i, g := range gameOrder {
		rank[g] = i
	}
	sort.SliceStable(battles, func(i, j int) bool { return rank[battles[i].VG] < rank[battles[j].VG] })

	tb := ds.T("gym_battles", "id", "gym_id", "gym_leader_id", "version_group_id", "version_id", "gym_order",
		"badge_en", "specialty_type_id", "battle_kind", "rematch_number", "is_double_battle", "notes", "source", "source_ref")
	ti := ds.T("gym_battle_items", "gym_battle_id", "slot", "item_id")
	tp := ds.T("gym_battle_pokemon", "id", "gym_battle_id", "slot", "pokemon_id", "level", "gender", "held_item_id",
		"ability_id", "ability_source", "moves_source", "tera_type_id", "can_dynamax", "is_gigantamax", "difficulty")
	tm := ds.T("gym_battle_pokemon_moves", "gym_battle_pokemon_id", "slot", "move_id")

	covered := map[string]bool{}
	monID := 0
	for bi, b := range battles {
		bid := bi + 1
		ctx := fmt.Sprintf("%s/%s(%s)", b.VG, b.Leader, b.Ref)
		var slot *slotDef
		for i := range gameSlots[b.VG] {
			if gameSlots[b.VG][i].leader == b.Leader {
				slot = &gameSlots[b.VG][i]
			}
		}
		if slot == nil {
			r.fail("%s: líder sin plaza definida en ese juego", ctx)
			continue
		}
		covered[b.VG+"/"+b.Leader] = true
		vg := L.vgByIdent[b.VG]
		gen := L.vgGeneration[vg]
		var version, typ any
		if slot.version != "" {
			version = L.versionByIdent[slot.version]
		}
		if slot.typ != "" {
			typ = L.typeByIdent[slot.typ]
		}
		badge := slot.badge
		if badge == "" {
			badge = gymBadge[slot.gym]
		}
		notes := strings.Trim(strings.Join([]string{slot.notes, b.Notes}, "; "), "; ")
		var rematch any
		if b.Rematch > 0 {
			rematch = b.Rematch
		}
		double := b.Double || forcedDouble[b.VG+"/"+b.Leader]
		tb.Add(bid, gymID[slot.gym], leaderID[b.Leader], vg, version, slot.order, badge, typ, b.Kind, rematch,
			boolInt(double), orNil(notes), b.Source, b.Ref)

		for i, it := range b.Items {
			if id := r.item(it, ctx); id != nil {
				ti.Add(bid, i+1, id)
			}
		}

		for si, m := range b.Mons {
			monID++
			mctx := fmt.Sprintf("%s #%d", ctx, si+1)
			pid, gender, gmax := r.pokemon(m, mctx)
			if pid == 0 {
				continue
			}

			// Movimientos
			var moves []int
			src := "game_data"
			switch {
			case m.Moves != nil:
				for _, mv := range m.Moves {
					moves = append(moves, r.move(mv, mctx))
				}
			default:
				moves = r.defaultMoves(pid, vg, m.Level, mctx)
				src = "default_levelup"
				if len(m.MoveOverride) > 0 {
					src = "mixed"
					for s, mv := range m.MoveOverride {
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
					moves = compact
				}
			}
			if len(moves) > 4 {
				r.fail("%s: más de 4 movimientos", mctx)
				moves = moves[:4]
			}

			// Habilidad
			var ability, abilitySrc any
			switch {
			case gen < 3 || b.VG == "lets-go-pikachu-lets-go-eevee":
				// sin habilidades
			case m.Ability != "":
				id, ok := lookup(L.abilityByNorm, nil, m.Ability)
				if !ok {
					r.fail("%s: habilidad %q", mctx, m.Ability)
				} else {
					ability, abilitySrc = id, "game_data"
				}
			default:
				opts := r.abilitiesIn(pid, gen)
				switch {
				case m.AbilitySlot > 0 && m.AbilitySlot <= len(opts):
					ability, abilitySrc = opts[m.AbilitySlot-1], "game_data"
				case len(opts) == 1:
					ability, abilitySrc = opts[0], "only_option"
				}
			}

			var tera any
			if m.Tera != "" {
				id, ok := L.typeByIdent[m.Tera]
				if !ok {
					r.fail("%s: teratipo %q", mctx, m.Tera)
				} else {
					tera = id
				}
			}
			gmax = gmax || m.Gmax
			tp.Add(monID, bid, si+1, pid, m.Level, orNil(gender), r.item(m.Item, mctx), ability, abilitySrc, src,
				tera, boolInt(m.Dynamax || gmax), boolInt(gmax), m.Difficulty)
			for k, mv := range moves {
				if mv != 0 {
					tm.Add(monID, k+1, mv)
				}
			}
		}
	}

	// Toda plaza definida debe tener al menos un combate
	for vg, slots := range gameSlots {
		for _, s := range slots {
			if !covered[vg+"/"+s.leader] {
				r.fail("%s/%s: no se encontró el equipo en las fuentes", vg, s.leader)
			}
		}
	}

	if len(r.errors) > 0 {
		sort.Strings(r.errors)
		return fmt.Errorf("%d problemas al resolver líderes:\n  %s", len(r.errors), strings.Join(r.errors, "\n  "))
	}
	logf("gimnasios: %d líderes, %d combates, %d Pokémon", len(leaderDefs), len(battles), monID)
	return nil
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
