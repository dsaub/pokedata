package main

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	langES = 7
	langEN = 9
)

var langCode = map[int]string{langES: "es", langEN: "en"}

// Src da acceso cacheado a los CSV de PokeAPI.
type Src struct {
	dir   string
	cache map[string]*CSV
}

func (s *Src) csv(name string) *CSV {
	if c, ok := s.cache[name]; ok {
		return c
	}
	c := readCSV(filepath.Join(s.dir, "pokeapi", name+".csv"))
	s.cache[name] = c
	return c
}

// nameMap: id -> idioma -> texto
type nameMap map[int]map[int]string

func (m nameMap) get(id, lang int) any {
	if v, ok := m[id][lang]; ok && v != "" {
		return v
	}
	return nil
}

// names lee una tabla *_names / *_prose y devuelve la columna col en es/en.
func (s *Src) names(file, idCol, col string) nameMap {
	c := s.csv(file)
	langCol := "local_language_id"
	if _, ok := c.idx[langCol]; !ok {
		langCol = "language_id"
	}
	m := nameMap{}
	c.Each(func(r Rec) {
		l := r.Int(langCol)
		if l != langES && l != langEN {
			return
		}
		id := r.Int(idCol)
		if m[id] == nil {
			m[id] = map[int]string{}
		}
		m[id][l] = r.Str(col)
	})
	return m
}

// Texto de los juegos: saltos de línea y guiones blandos fuera.
var spaceRe = regexp.MustCompile(`\s+`)

func cleanFlavor(s string) string {
	s = strings.ReplaceAll(s, "­\n", "")
	s = strings.ReplaceAll(s, "­", "")
	s = strings.ReplaceAll(s, "\f", " ")
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}

// Marcado de PokeAPI: [texto]{mechanic:algo} o []{move:tackle}
var markupRe = regexp.MustCompile(`\[([^\]]*)\]\{\w+:([^}]*)\}`)

func cleanMarkup(s string) string {
	return markupRe.ReplaceAllStringFunc(s, func(m string) string {
		p := markupRe.FindStringSubmatch(m)
		if p[1] != "" {
			return p[1]
		}
		return strings.ReplaceAll(p[2], "-", " ")
	})
}

func orNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// latestFlavor elige el texto más reciente (según rank) por entidad e idioma.
func latestFlavor(c *CSV, idCol, verCol string, rank func(int) int) nameMap {
	best := map[[2]int]int{}
	m := nameMap{}
	c.Each(func(r Rec) {
		l := r.Int("language_id")
		if l != langES && l != langEN {
			return
		}
		id := r.Int(idCol)
		k := [2]int{id, l}
		rk := rank(r.Int(verCol))
		if b, ok := best[k]; ok && b >= rk {
			return
		}
		best[k] = rk
		if m[id] == nil {
			m[id] = map[int]string{}
		}
		m[id][l] = cleanFlavor(r.Str("flavor_text"))
	})
	return m
}

// ---------------------------------------------------------------------
// Índices que necesita la parte de gimnasios
// ---------------------------------------------------------------------

type lvMove struct{ level, order, move int }

type abilitySlot struct {
	ability, slot, gen int // gen = 0 para las actuales
	hidden             bool
}

type Lookups struct {
	speciesByNorm  map[string]int
	speciesIdent   map[int]string
	pokemonByIdent map[string]int
	defaultPokemon map[int]int // especie -> pokemon por defecto
	moveByNorm     map[string]int
	itemByNorm     map[string]int
	abilityByNorm  map[string]int
	typeByIdent    map[string]int
	vgByIdent      map[string]int
	versionByIdent map[string]int
	vgGeneration   map[int]int
	regionByIdent  map[string]int
	locationByID   map[string]int
	locationNameEn map[int]string
	abilityGen     map[int]int
	levelup        map[[2]int][]lvMove // (pokemon, version_group)
	abilitiesNow   map[int][]abilitySlot
	abilitiesPast  map[int][]abilitySlot
	formsBySpecies map[int][]int // especie -> pokemon_id por form_order (formas no por defecto)
}

var normRe = regexp.MustCompile(`[^a-z0-9]`)

// norm reduce un nombre a [a-z0-9] para casar fuentes distintas.
func norm(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("é", "e", "♀", "f", "♂", "m", "’", "", "'", "").Replace(s)
	return normRe.ReplaceAllString(s, "")
}

func addNorm(m map[string]int, key string, id int) {
	k := norm(key)
	if k == "" {
		return
	}
	if _, ok := m[k]; !ok {
		m[k] = id
	}
}

// ---------------------------------------------------------------------

func loadPokeAPI(ds *Dataset, s *Src) *Lookups {
	L := &Lookups{
		speciesByNorm: map[string]int{}, speciesIdent: map[int]string{}, pokemonByIdent: map[string]int{}, defaultPokemon: map[int]int{},
		moveByNorm: map[string]int{}, itemByNorm: map[string]int{}, abilityByNorm: map[string]int{},
		typeByIdent: map[string]int{}, vgByIdent: map[string]int{}, versionByIdent: map[string]int{},
		vgGeneration: map[int]int{}, regionByIdent: map[string]int{}, locationByID: map[string]int{},
		locationNameEn: map[int]string{}, abilityGen: map[int]int{}, levelup: map[[2]int][]lvMove{},
		abilitiesNow: map[int][]abilitySlot{}, abilitiesPast: map[int][]abilitySlot{},
		formsBySpecies: map[int][]int{},
	}

	vgRank, versionRank := loadGames(ds, s, L)
	loadTypesAndStats(ds, s, L)
	loadAbilities(ds, s, L, vgRank)
	loadItems(ds, s, L, vgRank)
	loadMoves(ds, s, L, vgRank)
	loadBreeding(ds, s)
	loadSpecies(ds, s, L, versionRank)
	loadPokemon(ds, s, L)
	loadEvolutionAndLocations(ds, s, L)
	loadEncounters(ds, s)
	loadFormsBySpecies(s, L)
	return L
}

// loadGames: Regiones, generaciones, juegos.
func loadGames(ds *Dataset, s *Src, L *Lookups) (vgRank, versionRank func(int) int) {
	n := s.names("region_names", "region_id", "name")
	t := ds.T("regions", "id", "identifier", "name_en", "name_es")
	s.csv("regions").Each(func(r Rec) {
		id := r.Int("id")
		L.regionByIdent[r.Str("identifier")] = id
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})

	n = s.names("generation_names", "generation_id", "name")
	t = ds.T("generations", "id", "identifier", "main_region_id", "name_en", "name_es")
	s.csv("generations").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.I("main_region_id"), n.get(id, langEN), n.get(id, langES))
	})

	vgOrder := map[int]int{}
	t = ds.T("version_groups", "id", "identifier", "generation_id", "sort_order")
	s.csv("version_groups").Each(func(r Rec) {
		id := r.Int("id")
		vgOrder[id] = r.Int("order")
		L.vgByIdent[r.Str("identifier")] = id
		L.vgGeneration[id] = r.Int("generation_id")
		t.Add(id, r.Str("identifier"), r.Int("generation_id"), r.I("order"))
	})

	t = ds.T("version_group_regions", "version_group_id", "region_id")
	s.csv("version_group_regions").Each(func(r Rec) { t.Add(r.Int("version_group_id"), r.Int("region_id")) })

	versionVG := map[int]int{}
	n = s.names("version_names", "version_id", "name")
	t = ds.T("versions", "id", "identifier", "version_group_id", "name_en", "name_es")
	s.csv("versions").Each(func(r Rec) {
		id := r.Int("id")
		versionVG[id] = r.Int("version_group_id")
		L.versionByIdent[r.Str("identifier")] = id
		t.Add(id, r.Str("identifier"), r.Int("version_group_id"), n.get(id, langEN), n.get(id, langES))
	})
	vgRank = func(vg int) int { return vgOrder[vg] }
	versionRank = func(v int) int { return vgOrder[versionVG[v]]*100 + v }
	return vgRank, versionRank
}

// loadTypesAndStats: Tipos, estadísticas, naturalezas.
func loadTypesAndStats(ds *Dataset, s *Src, L *Lookups) {
	n := s.names("move_damage_class_prose", "move_damage_class_id", "name")
	d := s.names("move_damage_class_prose", "move_damage_class_id", "description")
	t := ds.T("move_damage_classes", "id", "identifier", "name_en", "name_es", "description_en")
	s.csv("move_damage_classes").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES), d.get(id, langEN))
	})

	n = s.names("type_names", "type_id", "name")
	t = ds.T("types", "id", "identifier", "generation_id", "damage_class_id", "name_en", "name_es")
	s.csv("types").Each(func(r Rec) {
		id := r.Int("id")
		L.typeByIdent[r.Str("identifier")] = id
		t.Add(id, r.Str("identifier"), r.Int("generation_id"), r.I("damage_class_id"), n.get(id, langEN), n.get(id, langES))
	})

	t = ds.T("type_efficacy", "attacking_type_id", "defending_type_id", "multiplier")
	s.csv("type_efficacy").Each(func(r Rec) {
		t.Add(r.Int("damage_type_id"), r.Int("target_type_id"), float64(r.Int("damage_factor"))/100)
	})
	t = ds.T("type_efficacy_past", "attacking_type_id", "defending_type_id", "multiplier", "generation_id")
	s.csv("type_efficacy_past").Each(func(r Rec) {
		t.Add(r.Int("damage_type_id"), r.Int("target_type_id"), float64(r.Int("damage_factor"))/100, r.Int("generation_id"))
	})

	n = s.names("stat_names", "stat_id", "name")
	t = ds.T("stats", "id", "identifier", "damage_class_id", "is_battle_only", "name_en", "name_es")
	s.csv("stats").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.I("damage_class_id"), r.Int("is_battle_only"), n.get(id, langEN), n.get(id, langES))
	})

	n = s.names("contest_type_names", "contest_type_id", "name")
	f := s.names("contest_type_names", "contest_type_id", "flavor")
	t = ds.T("contest_types", "id", "identifier", "name_en", "name_es", "flavor_en", "flavor_es")
	s.csv("contest_types").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES), f.get(id, langEN), f.get(id, langES))
	})

	n = s.names("nature_names", "nature_id", "name")
	t = ds.T("natures", "id", "identifier", "name_en", "name_es", "increased_stat_id", "decreased_stat_id",
		"likes_flavor_id", "hates_flavor_id", "game_index")
	s.csv("natures").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES), r.Int("increased_stat_id"),
			r.Int("decreased_stat_id"), r.Int("likes_flavor_id"), r.Int("hates_flavor_id"), r.I("game_index"))
	})

	n = s.names("characteristic_text", "characteristic_id", "message")
	t = ds.T("characteristics", "id", "stat_id", "gene_mod_5", "message_en", "message_es")
	s.csv("characteristics").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Int("stat_id"), r.Int("gene_mod_5"), n.get(id, langEN), n.get(id, langES))
	})
}

// loadAbilities: Habilidades.
func loadAbilities(ds *Dataset, s *Src, L *Lookups, vgRank func(int) int) {
	n := s.names("ability_names", "ability_id", "name")
	se := s.names("ability_prose", "ability_id", "short_effect")
	ef := s.names("ability_prose", "ability_id", "effect")
	fl := latestFlavor(s.csv("ability_flavor_text"), "ability_id", "version_group_id", vgRank)
	t := ds.T("abilities", "id", "identifier", "generation_id", "is_main_series", "name_en", "name_es",
		"short_effect_en", "effect_en", "flavor_text_en", "flavor_text_es")
	s.csv("abilities").Each(func(r Rec) {
		id := r.Int("id")
		L.abilityGen[id] = r.Int("generation_id")
		addNorm(L.abilityByNorm, r.Str("identifier"), id)
		if v, ok := n[id][langEN]; ok {
			addNorm(L.abilityByNorm, v, id)
		}
		t.Add(id, r.Str("identifier"), r.Int("generation_id"), r.B("is_main_series"), n.get(id, langEN), n.get(id, langES),
			proseClean(se.get(id, langEN)), proseClean(ef.get(id, langEN)), fl.get(id, langEN), fl.get(id, langES))
	})
}

// loadItems: Objetos.
func loadItems(ds *Dataset, s *Src, L *Lookups, vgRank func(int) int) {
	n := s.names("item_pocket_names", "item_pocket_id", "name")
	t := ds.T("item_pockets", "id", "identifier", "name_en", "name_es")
	s.csv("item_pockets").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})

	n = s.names("item_category_prose", "item_category_id", "name")
	t = ds.T("item_categories", "id", "identifier", "pocket_id", "name_en", "name_es")
	s.csv("item_categories").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.Int("pocket_id"), n.get(id, langEN), n.get(id, langES))
	})

	n = s.names("item_fling_effect_prose", "item_fling_effect_id", "effect")
	t = ds.T("item_fling_effects", "id", "identifier", "effect_en")
	s.csv("item_fling_effects").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), proseClean(n.get(id, langEN)))
	})

	n = s.names("item_names", "item_id", "name")
	se := s.names("item_prose", "item_id", "short_effect")
	ef := s.names("item_prose", "item_id", "effect")
	ifl := s.csv("item_flavor_text")
	fl := latestFlavor(ifl, "item_id", "version_group_id", vgRank)
	t = ds.T("items", "id", "identifier", "category_id", "cost", "fling_power", "fling_effect_id", "name_en", "name_es",
		"short_effect_en", "effect_en", "flavor_text_en", "flavor_text_es")
	itemIdents := map[string]bool{}
	s.csv("items").Each(func(r Rec) {
		id := r.Int("id")
		ident := r.Str("identifier")
		if itemIdents[ident] { // PokeAPI repite alguno (roseli-berry)
			ident += "-" + strconv.Itoa(id)
		}
		itemIdents[ident] = true
		addNorm(L.itemByNorm, ident, id)
		if v, ok := n[id][langEN]; ok {
			addNorm(L.itemByNorm, v, id)
		}
		t.Add(id, ident, r.Int("category_id"), r.I("cost"), r.I("fling_power"), r.I("fling_effect_id"),
			n.get(id, langEN), n.get(id, langES), proseClean(se.get(id, langEN)), proseClean(ef.get(id, langEN)),
			fl.get(id, langEN), fl.get(id, langES))
	})

	n = s.names("item_flag_prose", "item_flag_id", "name")
	d := s.names("item_flag_prose", "item_flag_id", "description")
	t = ds.T("item_flags", "id", "identifier", "name_en", "description_en")
	s.csv("item_flags").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), d.get(id, langEN))
	})
	t = ds.T("item_flag_map", "item_id", "item_flag_id")
	s.csv("item_flag_map").Each(func(r Rec) { t.Add(r.Int("item_id"), r.Int("item_flag_id")) })

	t = ds.T("item_flavor_text", "item_id", "version_group_id", "language", "flavor_text")
	addFlavorRows(t, ifl, "item_id", "version_group_id")

	n = s.names("berry_firmness_names", "berry_firmness_id", "name")
	t = ds.T("berry_firmness", "id", "identifier", "name_en")
	s.csv("berry_firmness").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN))
	})
	t = ds.T("berries", "id", "item_id", "firmness_id", "natural_gift_power", "natural_gift_type_id", "size_mm",
		"max_harvest", "growth_time", "soil_dryness", "smoothness")
	s.csv("berries").Each(func(r Rec) {
		t.Add(r.Int("id"), r.Int("item_id"), r.I("firmness_id"), r.I("natural_gift_power"), r.I("natural_gift_type_id"),
			r.I("size"), r.I("max_harvest"), r.I("growth_time"), r.I("soil_dryness"), r.I("smoothness"))
	})
	t = ds.T("berry_flavors", "berry_id", "contest_type_id", "flavor")
	s.csv("berry_flavors").Each(func(r Rec) { t.Add(r.Int("berry_id"), r.Int("contest_type_id"), r.Int("flavor")) })
}

// loadMoves: Movimientos.
func loadMoves(ds *Dataset, s *Src, L *Lookups, vgRank func(int) int) {
	n := s.names("move_target_prose", "move_target_id", "name")
	d := s.names("move_target_prose", "move_target_id", "description")
	t := ds.T("move_targets", "id", "identifier", "name_en", "description_en")
	s.csv("move_targets").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), d.get(id, langEN))
	})

	d = s.names("move_meta_category_prose", "move_meta_category_id", "description")
	t = ds.T("move_meta_categories", "id", "identifier", "description_en")
	s.csv("move_meta_categories").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), d.get(id, langEN))
	})

	n = s.names("move_meta_ailment_names", "move_meta_ailment_id", "name")
	t = ds.T("move_meta_ailments", "id", "identifier", "name_en", "name_es")
	s.csv("move_meta_ailments").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})

	n = s.names("move_names", "move_id", "name")
	se := s.names("move_effect_prose", "move_effect_id", "short_effect")
	ef := s.names("move_effect_prose", "move_effect_id", "effect")
	mfl := s.csv("move_flavor_text")
	fl := latestFlavor(mfl, "move_id", "version_group_id", vgRank)
	t = ds.T("moves", "id", "identifier", "generation_id", "type_id", "damage_class_id", "power", "pp", "accuracy",
		"priority", "target_id", "effect_chance", "contest_type_id", "name_en", "name_es", "short_effect_en",
		"effect_en", "flavor_text_en", "flavor_text_es")
	s.csv("moves").Each(func(r Rec) {
		id := r.Int("id")
		addNorm(L.moveByNorm, r.Str("identifier"), id)
		if v, ok := n[id][langEN]; ok {
			addNorm(L.moveByNorm, v, id)
		}
		eff := r.Int("effect_id")
		chance := r.Str("effect_chance")
		sub := func(v any) any {
			if v == nil {
				return nil
			}
			return proseClean(strings.ReplaceAll(v.(string), "$effect_chance", chance))
		}
		t.Add(id, r.Str("identifier"), r.Int("generation_id"), r.Int("type_id"), r.I("damage_class_id"), r.I("power"),
			r.I("pp"), r.I("accuracy"), r.Int("priority"), r.I("target_id"), r.I("effect_chance"), r.I("contest_type_id"),
			n.get(id, langEN), n.get(id, langES), sub(se.get(eff, langEN)), sub(ef.get(eff, langEN)),
			fl.get(id, langEN), fl.get(id, langES))
	})

	t = ds.T("move_meta", "move_id", "meta_category_id", "meta_ailment_id", "min_hits", "max_hits", "min_turns",
		"max_turns", "drain", "healing", "crit_rate", "ailment_chance", "flinch_chance", "stat_chance")
	s.csv("move_meta").Each(func(r Rec) {
		t.Add(r.Int("move_id"), r.I("meta_category_id"), r.I("meta_ailment_id"), r.I("min_hits"), r.I("max_hits"),
			r.I("min_turns"), r.I("max_turns"), r.I("drain"), r.I("healing"), r.I("crit_rate"), r.I("ailment_chance"),
			r.I("flinch_chance"), r.I("stat_chance"))
	})
	t = ds.T("move_meta_stat_changes", "move_id", "stat_id", "change")
	s.csv("move_meta_stat_changes").Each(func(r Rec) { t.Add(r.Int("move_id"), r.Int("stat_id"), r.Int("change")) })

	n = s.names("move_flag_prose", "move_flag_id", "name")
	d = s.names("move_flag_prose", "move_flag_id", "description")
	t = ds.T("move_flags", "id", "identifier", "name_en", "description_en")
	s.csv("move_flags").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), proseClean(d.get(id, langEN)))
	})
	t = ds.T("move_flag_map", "move_id", "move_flag_id")
	s.csv("move_flag_map").Each(func(r Rec) { t.Add(r.Int("move_id"), r.Int("move_flag_id")) })

	t = ds.T("move_changelog", "move_id", "changed_in_version_group_id", "type_id", "power", "pp", "accuracy",
		"priority", "target_id", "effect_chance")
	s.csv("move_changelog").Each(func(r Rec) {
		t.Add(r.Int("move_id"), r.Int("changed_in_version_group_id"), r.I("type_id"), r.I("power"), r.I("pp"),
			r.I("accuracy"), r.I("priority"), r.I("target_id"), r.I("effect_chance"))
	})

	t = ds.T("move_flavor_text", "move_id", "version_group_id", "language", "flavor_text")
	addFlavorRows(t, mfl, "move_id", "version_group_id")

	t = ds.T("machines", "machine_number", "version_group_id", "item_id", "move_id")
	s.csv("machines").Each(func(r Rec) {
		t.Add(r.Int("machine_number"), r.Int("version_group_id"), r.Int("item_id"), r.Int("move_id"))
	})
}

// loadBreeding: Crianza, crecimiento, aspecto.
func loadBreeding(ds *Dataset, s *Src) {
	n := s.names("growth_rate_prose", "growth_rate_id", "name")
	t := ds.T("growth_rates", "id", "identifier", "formula", "name_en", "name_es")
	s.csv("growth_rates").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.S("formula"), n.get(id, langEN), n.get(id, langES))
	})
	t = ds.T("experience", "growth_rate_id", "level", "experience")
	s.csv("experience").Each(func(r Rec) { t.Add(r.Int("growth_rate_id"), r.Int("level"), r.Int("experience")) })

	simpleNamed(ds, s, "egg_groups", "egg_group_prose", "egg_group_id")
	simpleNamed(ds, s, "pokemon_colors", "pokemon_color_names", "pokemon_color_id")
	simpleNamed(ds, s, "pokemon_shapes", "pokemon_shape_prose", "pokemon_shape_id")
	simpleNamed(ds, s, "pokemon_habitats", "pokemon_habitat_names", "pokemon_habitat_id")

	t = ds.T("evolution_chains", "id", "baby_trigger_item_id")
	s.csv("evolution_chains").Each(func(r Rec) { t.Add(r.Int("id"), r.I("baby_trigger_item_id")) })
}

// loadSpecies: Especies.
func loadSpecies(ds *Dataset, s *Src, L *Lookups, versionRank func(int) int) {
	n := s.names("pokemon_species_names", "pokemon_species_id", "name")
	g := s.names("pokemon_species_names", "pokemon_species_id", "genus")
	sfl := s.csv("pokemon_species_flavor_text")
	fl := latestFlavor(sfl, "species_id", "version_id", versionRank)
	t := ds.T("pokemon_species", "id", "identifier", "name_en", "name_es", "genus_en", "genus_es", "generation_id",
		"evolves_from_species_id", "evolution_chain_id", "color_id", "shape_id", "habitat_id", "gender_rate",
		"capture_rate", "base_happiness", "is_baby", "hatch_counter", "has_gender_differences", "growth_rate_id",
		"forms_switchable", "is_legendary", "is_mythical", "sort_order", "flavor_text_en", "flavor_text_es")
	s.csv("pokemon_species").Each(func(r Rec) {
		id := r.Int("id")
		L.speciesIdent[id] = r.Str("identifier")
		addNorm(L.speciesByNorm, r.Str("identifier"), id)
		if v, ok := n[id][langEN]; ok {
			addNorm(L.speciesByNorm, v, id)
		}
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES), g.get(id, langEN), g.get(id, langES),
			r.Int("generation_id"), r.I("evolves_from_species_id"), r.I("evolution_chain_id"), r.I("color_id"),
			r.I("shape_id"), r.I("habitat_id"), r.Int("gender_rate"), r.Int("capture_rate"), r.I("base_happiness"),
			r.Int("is_baby"), r.I("hatch_counter"), r.Int("has_gender_differences"), r.Int("growth_rate_id"),
			r.Int("forms_switchable"), r.Int("is_legendary"), r.Int("is_mythical"), r.I("order"),
			fl.get(id, langEN), fl.get(id, langES))
	})
	t = ds.T("pokemon_species_flavor_text", "species_id", "version_id", "language", "flavor_text")
	addFlavorRows(t, sfl, "species_id", "version_id")

	t = ds.T("pokemon_egg_groups", "species_id", "egg_group_id")
	s.csv("pokemon_egg_groups").Each(func(r Rec) { t.Add(r.Int("species_id"), r.Int("egg_group_id")) })
}

// loadPokemon: Pokémon y formas.
func loadPokemon(ds *Dataset, s *Src, L *Lookups) {
	t := ds.T("pokemon", "id", "identifier", "species_id", "height_dm", "weight_hg", "base_experience", "sort_order", "is_default")
	s.csv("pokemon").Each(func(r Rec) {
		id := r.Int("id")
		L.pokemonByIdent[r.Str("identifier")] = id
		if r.Int("is_default") == 1 {
			L.defaultPokemon[r.Int("species_id")] = id
		}
		t.Add(id, r.Str("identifier"), r.Int("species_id"), r.I("height"), r.I("weight"), r.I("base_experience"),
			r.I("order"), r.Int("is_default"))
	})

	fn := s.names("pokemon_form_names", "pokemon_form_id", "form_name")
	pn := s.names("pokemon_form_names", "pokemon_form_id", "pokemon_name")
	t = ds.T("pokemon_forms", "id", "identifier", "form_identifier", "pokemon_id", "introduced_in_version_group_id",
		"is_default", "is_battle_only", "is_mega", "form_order", "sort_order", "form_name_en", "form_name_es",
		"full_name_en", "full_name_es")
	s.csv("pokemon_forms").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.S("form_identifier"), r.Int("pokemon_id"), r.I("introduced_in_version_group_id"),
			r.Int("is_default"), r.Int("is_battle_only"), r.Int("is_mega"), r.I("form_order"), r.I("order"),
			fn.get(id, langEN), fn.get(id, langES), pn.get(id, langEN), pn.get(id, langES))
	})

	t = ds.T("pokemon_types", "pokemon_id", "type_id", "slot")
	s.csv("pokemon_types").Each(func(r Rec) { t.Add(r.Int("pokemon_id"), r.Int("type_id"), r.Int("slot")) })
	t = ds.T("pokemon_types_past", "pokemon_id", "generation_id", "type_id", "slot")
	s.csv("pokemon_types_past").Each(func(r Rec) {
		t.Add(r.Int("pokemon_id"), r.Int("generation_id"), r.Int("type_id"), r.Int("slot"))
	})

	t = ds.T("pokemon_abilities", "pokemon_id", "ability_id", "is_hidden", "slot")
	s.csv("pokemon_abilities").Each(func(r Rec) {
		p := r.Int("pokemon_id")
		L.abilitiesNow[p] = append(L.abilitiesNow[p], abilitySlot{ability: r.Int("ability_id"), slot: r.Int("slot"), hidden: r.Int("is_hidden") == 1})
		t.Add(p, r.Int("ability_id"), r.Int("is_hidden"), r.Int("slot"))
	})
	t = ds.T("pokemon_abilities_past", "pokemon_id", "generation_id", "ability_id", "is_hidden", "slot")
	s.csv("pokemon_abilities_past").Each(func(r Rec) {
		p := r.Int("pokemon_id")
		L.abilitiesPast[p] = append(L.abilitiesPast[p], abilitySlot{ability: r.Int("ability_id"), slot: r.Int("slot"),
			gen: r.Int("generation_id"), hidden: r.B("is_hidden") == int64(1)})
		t.Add(p, r.Int("generation_id"), r.I("ability_id"), r.B("is_hidden"), r.Int("slot"))
	})

	t = ds.T("pokemon_stats", "pokemon_id", "stat_id", "base_stat", "effort")
	s.csv("pokemon_stats").Each(func(r Rec) {
		t.Add(r.Int("pokemon_id"), r.Int("stat_id"), r.Int("base_stat"), r.Int("effort"))
	})
	t = ds.T("pokemon_stats_past", "pokemon_id", "generation_id", "stat_id", "base_stat", "effort")
	s.csv("pokemon_stats_past").Each(func(r Rec) {
		t.Add(r.Int("pokemon_id"), r.Int("generation_id"), r.Int("stat_id"), r.I("base_stat"), r.I("effort"))
	})

	t = ds.T("pokemon_items", "pokemon_id", "version_id", "item_id", "rarity")
	s.csv("pokemon_items").Each(func(r Rec) {
		t.Add(r.Int("pokemon_id"), r.Int("version_id"), r.Int("item_id"), r.Int("rarity"))
	})

	n := s.names("pokemon_move_method_prose", "pokemon_move_method_id", "name")
	d := s.names("pokemon_move_method_prose", "pokemon_move_method_id", "description")
	t = ds.T("pokemon_move_methods", "id", "identifier", "name_en", "name_es", "description_en")
	s.csv("pokemon_move_methods").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES), d.get(id, langEN))
	})

	t = ds.T("pokemon_moves", "pokemon_id", "version_group_id", "move_id", "pokemon_move_method_id", "level",
		"sort_order", "mastery")
	s.csv("pokemon_moves").Each(func(r Rec) {
		p, vg, m, meth := r.Int("pokemon_id"), r.Int("version_group_id"), r.Int("move_id"), r.Int("pokemon_move_method_id")
		if meth == 1 { // level-up
			k := [2]int{p, vg}
			L.levelup[k] = append(L.levelup[k], lvMove{r.Int("level"), r.Int("order"), m})
		}
		t.Add(p, vg, m, meth, r.Int("level"), r.I("order"), r.I("mastery"))
	})
}

// loadEvolutionAndLocations: Evolución y ubicaciones.
func loadEvolutionAndLocations(ds *Dataset, s *Src, L *Lookups) {
	n := s.names("evolution_trigger_prose", "evolution_trigger_id", "name")
	t := ds.T("evolution_triggers", "id", "identifier", "name_en", "name_es")
	s.csv("evolution_triggers").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})
	t = ds.T("genders", "id", "identifier")
	s.csv("genders").Each(func(r Rec) { t.Add(r.Int("id"), r.Str("identifier")) })

	n = s.names("location_names", "location_id", "name")
	t = ds.T("locations", "id", "identifier", "region_id", "name_en", "name_es")
	s.csv("locations").Each(func(r Rec) {
		id := r.Int("id")
		L.locationByID[r.Str("identifier")] = id
		if v, ok := n[id][langEN]; ok {
			L.locationNameEn[id] = v
		}
		t.Add(id, r.Str("identifier"), r.I("region_id"), n.get(id, langEN), n.get(id, langES))
	})
	n = s.names("location_area_prose", "location_area_id", "name")
	t = ds.T("location_areas", "id", "location_id", "game_index", "identifier", "name_en")
	s.csv("location_areas").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Int("location_id"), r.I("game_index"), r.S("identifier"), n.get(id, langEN))
	})

	evoCols := []string{"id", "evolved_species_id", "evolution_trigger_id", "version_group_id", "is_default",
		"trigger_item_id", "minimum_level", "gender_id", "location_id", "held_item_id", "time_of_day", "known_move_id",
		"known_move_type_id", "minimum_happiness", "minimum_beauty", "minimum_affection", "relative_physical_stats",
		"party_species_id", "party_type_id", "trade_species_id", "needs_overworld_rain", "turn_upside_down",
		"needs_multiplayer", "near_special_rock", "region_id", "required_pokemon_form_id", "evolved_pokemon_form_id",
		"used_move_id", "minimum_move_count", "minimum_steps", "minimum_damage_taken", "nature_bitmask",
		"condition_expression", "percentage_chance"}
	t = ds.T("pokemon_evolution", evoCols...)
	s.csv("pokemon_evolution").Each(func(r Rec) {
		row := make([]any, len(evoCols))
		for i, c := range evoCols {
			switch c {
			case "time_of_day", "condition_expression":
				row[i] = r.S(c)
			case "needs_overworld_rain", "turn_upside_down", "needs_multiplayer", "near_special_rock", "is_default":
				row[i] = r.B(c)
			default:
				row[i] = r.I(c)
			}
		}
		t.Add(row...)
	})

	n = s.names("pokedex_prose", "pokedex_id", "name")
	d := s.names("pokedex_prose", "pokedex_id", "description")
	t = ds.T("pokedexes", "id", "identifier", "region_id", "is_main_series", "name_en", "name_es", "description_en")
	s.csv("pokedexes").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.I("region_id"), r.Int("is_main_series"), n.get(id, langEN), n.get(id, langES), d.get(id, langEN))
	})
	t = ds.T("pokemon_dex_numbers", "species_id", "pokedex_id", "pokedex_number")
	s.csv("pokemon_dex_numbers").Each(func(r Rec) {
		t.Add(r.Int("species_id"), r.Int("pokedex_id"), r.Int("pokedex_number"))
	})
}

// loadEncounters: Encuentros.
func loadEncounters(ds *Dataset, s *Src) {
	n := s.names("encounter_method_prose", "encounter_method_id", "name")
	t := ds.T("encounter_methods", "id", "identifier", "sort_order", "name_en", "name_es")
	s.csv("encounter_methods").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), r.I("order"), n.get(id, langEN), n.get(id, langES))
	})
	t = ds.T("encounter_slots", "id", "version_group_id", "encounter_method_id", "slot", "rarity")
	s.csv("encounter_slots").Each(func(r Rec) {
		t.Add(r.Int("id"), r.Int("version_group_id"), r.Int("encounter_method_id"), r.I("slot"), r.I("rarity"))
	})
	t = ds.T("encounters", "id", "version_id", "location_area_id", "encounter_slot_id", "pokemon_id", "min_level", "max_level")
	s.csv("encounters").Each(func(r Rec) {
		t.Add(r.Int("id"), r.Int("version_id"), r.Int("location_area_id"), r.Int("encounter_slot_id"),
			r.Int("pokemon_id"), r.Int("min_level"), r.Int("max_level"))
	})
	n = s.names("encounter_condition_prose", "encounter_condition_id", "name")
	t = ds.T("encounter_conditions", "id", "identifier", "name_en", "name_es")
	s.csv("encounter_conditions").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})
	n = s.names("encounter_condition_value_prose", "encounter_condition_value_id", "name")
	t = ds.T("encounter_condition_values", "id", "encounter_condition_id", "identifier", "is_default", "name_en", "name_es")
	s.csv("encounter_condition_values").Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Int("encounter_condition_id"), r.Str("identifier"), r.Int("is_default"), n.get(id, langEN), n.get(id, langES))
	})
	t = ds.T("encounter_condition_value_map", "encounter_id", "encounter_condition_value_id")
	s.csv("encounter_condition_value_map").Each(func(r Rec) {
		t.Add(r.Int("encounter_id"), r.Int("encounter_condition_value_id"))
	})
}

// loadFormsBySpecies: Formas alternativas por especie (para pret gen 4 "form": N).
func loadFormsBySpecies(s *Src, L *Lookups) {
	type fo struct{ order, pokemon int }
	tmp := map[int][]fo{}
	s.csv("pokemon").Each(func(r Rec) {
		tmp[r.Int("species_id")] = append(tmp[r.Int("species_id")], fo{r.Int("order")*10000 + r.Int("id"), r.Int("id")})
	})
	for sp, l := range tmp {
		sort.Slice(l, func(i, j int) bool { return l[i].order < l[j].order })
		for _, x := range l {
			L.formsBySpecies[sp] = append(L.formsBySpecies[sp], x.pokemon)
		}
	}
}

func proseClean(v any) any {
	if v == nil {
		return nil
	}
	return orNil(strings.TrimSpace(cleanMarkup(v.(string))))
}

func addFlavorRows(t *Table, c *CSV, idCol, verCol string) {
	seen := map[[3]int]bool{}
	c.Each(func(r Rec) {
		l := r.Int("language_id")
		code, ok := langCode[l]
		if !ok {
			return
		}
		k := [3]int{r.Int(idCol), r.Int(verCol), l}
		if seen[k] {
			return
		}
		seen[k] = true
		t.Add(k[0], k[1], code, cleanFlavor(r.Str("flavor_text")))
	})
}

// simpleNamed carga tablas (id, identifier, name_en, name_es).
func simpleNamed(ds *Dataset, s *Src, table, namesFile, idCol string) {
	n := s.names(namesFile, idCol, "name")
	t := ds.T(table, "id", "identifier", "name_en", "name_es")
	s.csv(table).Each(func(r Rec) {
		id := r.Int("id")
		t.Add(id, r.Str("identifier"), n.get(id, langEN), n.get(id, langES))
	})
}

func atoi(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		panic("entero inválido: " + s)
	}
	return n
}
