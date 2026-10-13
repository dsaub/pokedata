package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	pokeapiCSVBase = "https://raw.githubusercontent.com/PokeAPI/pokeapi/master/data/v2/csv/"
	pretBase       = "https://raw.githubusercontent.com/pret/"
	serebiiBase    = "https://www.serebii.net/"
	userAgent      = "Mozilla/5.0 (X11; Linux x86_64) pokedata-gen/1.0"

	// Ficheros de equipos en las descompilaciones de pret.
	pretGen12Parties = "data/trainers/parties.asm"
	pretGen3Parties  = "src/data/trainer_parties.h"
)

// CSV de PokeAPI que usa el generador.
var pokeapiCSVs = []string{
	"regions", "region_names", "generations", "generation_names", "version_groups",
	"version_group_regions", "versions", "version_names",
	"move_damage_classes", "move_damage_class_prose", "types", "type_names", "type_efficacy",
	"type_efficacy_past", "stats", "stat_names", "contest_types", "contest_type_names",
	"natures", "nature_names", "characteristics", "characteristic_text",
	"abilities", "ability_names", "ability_prose", "ability_flavor_text",
	"item_pockets", "item_pocket_names", "item_categories", "item_category_prose",
	"item_fling_effects", "item_fling_effect_prose", "items", "item_names", "item_prose",
	"item_flavor_text", "item_flags", "item_flag_prose", "item_flag_map",
	"berry_firmness", "berry_firmness_names", "berries", "berry_flavors",
	"move_targets", "move_target_prose", "move_meta_categories", "move_meta_category_prose",
	"move_meta_ailments", "move_meta_ailment_names", "moves", "move_names", "move_effect_prose",
	"move_flavor_text", "move_meta", "move_meta_stat_changes", "move_flags", "move_flag_prose",
	"move_flag_map", "move_changelog", "machines",
	"growth_rates", "growth_rate_prose", "experience", "egg_groups", "egg_group_prose",
	"pokemon_colors", "pokemon_color_names", "pokemon_shapes", "pokemon_shape_prose",
	"pokemon_habitats", "pokemon_habitat_names", "evolution_chains",
	"pokemon_species", "pokemon_species_names", "pokemon_species_flavor_text", "pokemon_egg_groups",
	"pokemon", "pokemon_forms", "pokemon_form_names", "pokemon_types", "pokemon_types_past",
	"pokemon_abilities", "pokemon_abilities_past", "pokemon_stats", "pokemon_stats_past",
	"pokemon_items", "pokemon_move_methods", "pokemon_move_method_prose", "pokemon_moves",
	"evolution_triggers", "evolution_trigger_prose", "genders",
	"locations", "location_names", "location_areas", "location_area_prose", "pokemon_evolution",
	"pokedexes", "pokedex_prose", "pokemon_dex_numbers",
	"encounter_methods", "encounter_method_prose", "encounter_slots", "encounters",
	"encounter_conditions", "encounter_condition_prose", "encounter_condition_values",
	"encounter_condition_value_prose", "encounter_condition_value_map",
}

// Ficheros de las descompilaciones de pret (datos reales de los juegos, gen 1-4).
var pretFiles = [][2]string{
	{"pokered", pretGen12Parties},
	{"pokered", "data/trainers/special_moves.asm"},
	{"pokeyellow", pretGen12Parties},
	{"pokeyellow", "data/trainers/special_moves.asm"},
	{"pokegold", pretGen12Parties},
	{"pokecrystal", pretGen12Parties},
	{"pokeruby", pretGen3Parties},
	{"pokeruby", "src/data/trainers_en.h"},
	{"pokeemerald", pretGen3Parties},
	{"pokeemerald", "src/data/trainers.h"},
	{"pokefirered", pretGen3Parties},
	{"pokefirered", "src/data/trainers.h"},
	{"pokediamond", "files/poketool/trainer/trdata.json"},
	{"pokeheartgold", "files/poketool/trainer/trainers.json"},
}

var platinumLeaderFiles = []string{
	"leader_roark", "leader_roark_rematch", "leader_gardenia", "leader_gardenia_rematch",
	"leader_fantina", "leader_fantina_rematch", "leader_maylene", "leader_maylene_rematch",
	"leader_wake", "leader_wake_rematch", "leader_byron", "leader_byron_rematch",
	"leader_candice", "leader_candice_rematch", "leader_volkner", "leader_volkner_rematch",
	"leader_volkner_fight_area",
}

// Páginas de Serebii (gen 5-9, sin descompilación disponible).
var serebiiPages = []string{
	"blackwhite/gyms", "black2white2/gyms", "xy/gyms", "omegarubyalphasapphire/gyms",
	"letsgopikachueevee/gyms", "swordshield/gyms", "brilliantdiamondshiningpearl/gyms",
	"scarletviolet/gyms",
}

type download struct{ url, dest string }

func sourceList(dir string) []download {
	var l []download
	for _, n := range pokeapiCSVs {
		l = append(l, download{pokeapiCSVBase + n + ".csv", filepath.Join(dir, "pokeapi", n+".csv")})
	}
	for _, f := range pretFiles {
		l = append(l, download{pretBase + f[0] + "/HEAD/" + f[1], filepath.Join(dir, "pret", f[0], f[1])})
	}
	for _, n := range platinumLeaderFiles {
		p := "res/trainers/data/" + n + ".json"
		l = append(l, download{pretBase + "pokeplatinum/HEAD/" + p, filepath.Join(dir, "pret", "pokeplatinum", p)})
	}
	for _, p := range serebiiPages {
		l = append(l, download{serebiiBase + p + ".shtml", filepath.Join(dir, "serebii", filepath.Dir(p)+".html")})
	}
	return l
}

// ensureSources descarga las fuentes que falten (o todas si refresh).
func ensureSources(ctx context.Context, dir string, refresh, offline bool, logf func(string, ...any)) error {
	var todo []download
	for _, d := range sourceList(dir) {
		if _, err := os.Stat(d.dest); err == nil && !refresh {
			continue
		}
		todo = append(todo, d)
	}
	if len(todo) == 0 {
		logf("fuentes: todas en caché (%s)", dir)
		return nil
	}
	if offline {
		return fmt.Errorf("modo offline y faltan %d ficheros, p. ej. %s", len(todo), todo[0].dest)
	}
	logf("fuentes: descargando %d ficheros en %s", len(todo), dir)

	client := &http.Client{Timeout: 5 * time.Minute}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, 6)
	)
	for _, d := range todo {
		wg.Add(1)
		go func(d download) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var err error
			for attempt := 1; attempt <= 3; attempt++ {
				if err = fetchOne(ctx, client, d); err == nil {
					return
				}
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}(d)
	}
	wg.Wait()
	if len(errs) > 0 {
		return fmt.Errorf("%d descargas fallidas, p. ej.: %w", len(errs), errs[0])
	}
	return nil
}

func fetchOne(ctx context.Context, client *http.Client, d download) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", d.url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP %d", d.url, resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(d.dest), 0o755); err != nil {
		return err
	}
	tmp := d.dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("%s: %w", d.url, err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, d.dest)
}
