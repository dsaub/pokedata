-- =====================================================================
--  pokedata — esquema SQLite
--  Parte 1: datos de la saga (fuente: PokeAPI, CSV oficiales)
--  Parte 2: líderes de gimnasio (fuentes: pret gen 1-4, Serebii gen 5-9)
--
--  Convenciones:
--    * `identifier` = slug estable en inglés (p. ej. 'mr-mime').
--    * `name_en` / `name_es` = nombre localizado (NULL si no existe en esa lengua).
--    * Alturas en decímetros, pesos en hectogramos (como en los juegos).
-- =====================================================================


-- ---------------------------------------------------------------------
--  Generaciones, regiones y juegos
-- ---------------------------------------------------------------------
CREATE TABLE regions (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE generations (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    main_region_id  INTEGER REFERENCES regions(id),
    name_en         TEXT,
    name_es         TEXT
);

CREATE TABLE version_groups (
    id             INTEGER PRIMARY KEY,
    identifier     TEXT NOT NULL UNIQUE,
    generation_id  INTEGER NOT NULL REFERENCES generations(id),
    sort_order     INTEGER
);

CREATE TABLE version_group_regions (
    version_group_id  INTEGER NOT NULL REFERENCES version_groups(id),
    region_id         INTEGER NOT NULL REFERENCES regions(id),
    PRIMARY KEY (version_group_id, region_id)
);

CREATE TABLE versions (
    id                INTEGER PRIMARY KEY,
    identifier        TEXT NOT NULL UNIQUE,
    version_group_id  INTEGER NOT NULL REFERENCES version_groups(id),
    name_en           TEXT,
    name_es           TEXT
);

-- ---------------------------------------------------------------------
--  Tipos y estadísticas
-- ---------------------------------------------------------------------
CREATE TABLE move_damage_classes (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,       -- physical / special / status
    name_en         TEXT,
    name_es         TEXT,
    description_en  TEXT
);

CREATE TABLE types (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,
    generation_id    INTEGER NOT NULL REFERENCES generations(id),
    damage_class_id  INTEGER REFERENCES move_damage_classes(id),  -- clase de daño del tipo antes de gen 4
    name_en          TEXT,
    name_es          TEXT
);

-- Tabla de efectividades actual. multiplier: 0, 0.5, 1, 2
CREATE TABLE type_efficacy (
    attacking_type_id  INTEGER NOT NULL REFERENCES types(id),
    defending_type_id  INTEGER NOT NULL REFERENCES types(id),
    multiplier         REAL NOT NULL,
    PRIMARY KEY (attacking_type_id, defending_type_id)
);

-- Efectividades que eran distintas en generaciones pasadas (válidas HASTA generation_id inclusive)
CREATE TABLE type_efficacy_past (
    attacking_type_id  INTEGER NOT NULL REFERENCES types(id),
    defending_type_id  INTEGER NOT NULL REFERENCES types(id),
    multiplier         REAL NOT NULL,
    generation_id      INTEGER NOT NULL REFERENCES generations(id),
    PRIMARY KEY (attacking_type_id, defending_type_id, generation_id)
);

CREATE TABLE stats (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,      -- hp, attack, defense, special-attack, ...
    damage_class_id  INTEGER REFERENCES move_damage_classes(id),
    is_battle_only   INTEGER NOT NULL,
    name_en          TEXT,
    name_es          TEXT
);

-- Sabores / tipos de concurso (los usan naturalezas y bayas)
CREATE TABLE contest_types (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT,
    flavor_en   TEXT,
    flavor_es   TEXT
);

CREATE TABLE natures (
    id                 INTEGER PRIMARY KEY,
    identifier         TEXT NOT NULL UNIQUE,
    name_en            TEXT,
    name_es            TEXT,
    increased_stat_id  INTEGER NOT NULL REFERENCES stats(id),
    decreased_stat_id  INTEGER NOT NULL REFERENCES stats(id),
    likes_flavor_id    INTEGER NOT NULL REFERENCES contest_types(id),
    hates_flavor_id    INTEGER NOT NULL REFERENCES contest_types(id),
    game_index         INTEGER
);

CREATE TABLE characteristics (
    id          INTEGER PRIMARY KEY,
    stat_id     INTEGER NOT NULL REFERENCES stats(id),
    gene_mod_5  INTEGER NOT NULL,
    message_en  TEXT,
    message_es  TEXT
);

-- ---------------------------------------------------------------------
--  Habilidades
-- ---------------------------------------------------------------------
CREATE TABLE abilities (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,
    generation_id    INTEGER NOT NULL REFERENCES generations(id),
    is_main_series   INTEGER NOT NULL,
    name_en          TEXT,
    name_es          TEXT,
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,   -- texto del juego más reciente
    flavor_text_es   TEXT
);

-- ---------------------------------------------------------------------
--  Objetos
-- ---------------------------------------------------------------------
CREATE TABLE item_pockets (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE item_categories (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    pocket_id   INTEGER NOT NULL REFERENCES item_pockets(id),
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE item_fling_effects (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    effect_en   TEXT
);

CREATE TABLE items (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,
    category_id      INTEGER NOT NULL REFERENCES item_categories(id),
    cost             INTEGER,
    fling_power      INTEGER,
    fling_effect_id  INTEGER REFERENCES item_fling_effects(id),
    name_en          TEXT,
    name_es          TEXT,
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,
    flavor_text_es   TEXT
);

CREATE TABLE item_flags (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    name_en         TEXT,
    description_en  TEXT
);

CREATE TABLE item_flag_map (
    item_id       INTEGER NOT NULL REFERENCES items(id),
    item_flag_id  INTEGER NOT NULL REFERENCES item_flags(id),
    PRIMARY KEY (item_id, item_flag_id)
);

CREATE TABLE item_flavor_text (
    item_id           INTEGER NOT NULL REFERENCES items(id),
    version_group_id  INTEGER NOT NULL REFERENCES version_groups(id),
    language          TEXT NOT NULL CHECK (language IN ('en','es')),
    flavor_text       TEXT NOT NULL,
    PRIMARY KEY (item_id, version_group_id, language)
);

CREATE TABLE berry_firmness (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT
);

CREATE TABLE berries (
    id                    INTEGER PRIMARY KEY,
    item_id               INTEGER NOT NULL UNIQUE REFERENCES items(id),
    firmness_id           INTEGER REFERENCES berry_firmness(id),
    natural_gift_power    INTEGER,
    natural_gift_type_id  INTEGER REFERENCES types(id),
    size_mm               INTEGER,
    max_harvest           INTEGER,
    growth_time           INTEGER,
    soil_dryness          INTEGER,
    smoothness            INTEGER
);

CREATE TABLE berry_flavors (
    berry_id         INTEGER NOT NULL REFERENCES berries(id),
    contest_type_id  INTEGER NOT NULL REFERENCES contest_types(id),
    flavor           INTEGER NOT NULL,
    PRIMARY KEY (berry_id, contest_type_id)
);

-- ---------------------------------------------------------------------
--  Movimientos
-- ---------------------------------------------------------------------
CREATE TABLE move_targets (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    name_en         TEXT,
    description_en  TEXT
);

CREATE TABLE move_meta_categories (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    description_en  TEXT
);

CREATE TABLE move_meta_ailments (
    id          INTEGER PRIMARY KEY,   -- -1 = desconocido, 0 = ninguno
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE moves (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,
    generation_id    INTEGER NOT NULL REFERENCES generations(id),
    type_id          INTEGER NOT NULL REFERENCES types(id),
    damage_class_id  INTEGER REFERENCES move_damage_classes(id),
    power            INTEGER,
    pp               INTEGER,
    accuracy         INTEGER,
    priority         INTEGER NOT NULL,
    target_id        INTEGER REFERENCES move_targets(id),
    effect_chance    INTEGER,
    contest_type_id  INTEGER REFERENCES contest_types(id),
    name_en          TEXT,
    name_es          TEXT,
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,
    flavor_text_es   TEXT
);

-- Metadatos de batalla del movimiento (1:1 con moves)
CREATE TABLE move_meta (
    move_id           INTEGER PRIMARY KEY REFERENCES moves(id),
    meta_category_id  INTEGER REFERENCES move_meta_categories(id),
    meta_ailment_id   INTEGER REFERENCES move_meta_ailments(id),
    min_hits          INTEGER,
    max_hits          INTEGER,
    min_turns         INTEGER,
    max_turns         INTEGER,
    drain             INTEGER,   -- % de daño drenado (negativo = retroceso)
    healing           INTEGER,   -- % de PS curados
    crit_rate         INTEGER,   -- etapas extra de crítico
    ailment_chance    INTEGER,
    flinch_chance     INTEGER,
    stat_chance       INTEGER
);

CREATE TABLE move_meta_stat_changes (
    move_id  INTEGER NOT NULL REFERENCES moves(id),
    stat_id  INTEGER NOT NULL REFERENCES stats(id),
    change   INTEGER NOT NULL,
    PRIMARY KEY (move_id, stat_id)
);

CREATE TABLE move_flags (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,   -- contact, sound, punch, ...
    name_en         TEXT,
    description_en  TEXT
);

CREATE TABLE move_flag_map (
    move_id       INTEGER NOT NULL REFERENCES moves(id),
    move_flag_id  INTEGER NOT NULL REFERENCES move_flags(id),
    PRIMARY KEY (move_id, move_flag_id)
);

-- Valores antiguos de un movimiento (válidos ANTES del version_group_id indicado)
CREATE TABLE move_changelog (
    move_id                     INTEGER NOT NULL REFERENCES moves(id),
    changed_in_version_group_id INTEGER NOT NULL REFERENCES version_groups(id),
    type_id                     INTEGER REFERENCES types(id),
    power                       INTEGER,
    pp                          INTEGER,
    accuracy                    INTEGER,
    priority                    INTEGER,
    target_id                   INTEGER REFERENCES move_targets(id),
    effect_chance               INTEGER,
    PRIMARY KEY (move_id, changed_in_version_group_id)
);

CREATE TABLE move_flavor_text (
    move_id           INTEGER NOT NULL REFERENCES moves(id),
    version_group_id  INTEGER NOT NULL REFERENCES version_groups(id),
    language          TEXT NOT NULL CHECK (language IN ('en','es')),
    flavor_text       TEXT NOT NULL,
    PRIMARY KEY (move_id, version_group_id, language)
);

-- MT/MO/DT por juego
CREATE TABLE machines (
    machine_number    INTEGER NOT NULL,
    version_group_id  INTEGER NOT NULL REFERENCES version_groups(id),
    item_id           INTEGER NOT NULL REFERENCES items(id),
    move_id           INTEGER NOT NULL REFERENCES moves(id),
    PRIMARY KEY (version_group_id, item_id)
);

-- ---------------------------------------------------------------------
--  Especies y Pokémon
-- ---------------------------------------------------------------------
CREATE TABLE growth_rates (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    formula     TEXT,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE experience (
    growth_rate_id  INTEGER NOT NULL REFERENCES growth_rates(id),
    level           INTEGER NOT NULL,
    experience      INTEGER NOT NULL,
    PRIMARY KEY (growth_rate_id, level)
);

CREATE TABLE egg_groups (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE pokemon_colors (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE pokemon_shapes (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE pokemon_habitats (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE evolution_chains (
    id                     INTEGER PRIMARY KEY,
    baby_trigger_item_id   INTEGER REFERENCES items(id)
);

-- Especie: lo común a todas las formas (Pikachu, Charizard, ...)
CREATE TABLE pokemon_species (
    id                       INTEGER PRIMARY KEY,         -- = nº de la Pokédex Nacional
    identifier               TEXT NOT NULL UNIQUE,
    name_en                  TEXT,
    name_es                  TEXT,
    genus_en                 TEXT,
    genus_es                 TEXT,
    generation_id            INTEGER NOT NULL REFERENCES generations(id),
    evolves_from_species_id  INTEGER REFERENCES pokemon_species(id),
    evolution_chain_id       INTEGER REFERENCES evolution_chains(id),
    color_id                 INTEGER REFERENCES pokemon_colors(id),
    shape_id                 INTEGER REFERENCES pokemon_shapes(id),
    habitat_id               INTEGER REFERENCES pokemon_habitats(id),
    gender_rate              INTEGER NOT NULL,  -- -1 sin género; si no, octavos de probabilidad de hembra
    capture_rate             INTEGER NOT NULL,
    base_happiness           INTEGER,
    is_baby                  INTEGER NOT NULL,
    hatch_counter            INTEGER,
    has_gender_differences   INTEGER NOT NULL,
    growth_rate_id           INTEGER NOT NULL REFERENCES growth_rates(id),
    forms_switchable         INTEGER NOT NULL,
    is_legendary             INTEGER NOT NULL,
    is_mythical              INTEGER NOT NULL,
    sort_order               INTEGER,
    flavor_text_en           TEXT,   -- entrada de Pokédex más reciente
    flavor_text_es           TEXT
);

CREATE TABLE pokemon_species_flavor_text (
    species_id   INTEGER NOT NULL REFERENCES pokemon_species(id),
    version_id   INTEGER NOT NULL REFERENCES versions(id),
    language     TEXT NOT NULL CHECK (language IN ('en','es')),
    flavor_text  TEXT NOT NULL,
    PRIMARY KEY (species_id, version_id, language)
);

CREATE TABLE pokemon_egg_groups (
    species_id    INTEGER NOT NULL REFERENCES pokemon_species(id),
    egg_group_id  INTEGER NOT NULL REFERENCES egg_groups(id),
    PRIMARY KEY (species_id, egg_group_id)
);

-- Pokémon "jugable": cada forma con estadísticas/tipos propios (charizard, charizard-mega-x, ...)
CREATE TABLE pokemon (
    id               INTEGER PRIMARY KEY,
    identifier       TEXT NOT NULL UNIQUE,
    species_id       INTEGER NOT NULL REFERENCES pokemon_species(id),
    height_dm        INTEGER,
    weight_hg        INTEGER,
    base_experience  INTEGER,
    sort_order       INTEGER,
    is_default       INTEGER NOT NULL      -- 1 = forma base de la especie
);

CREATE TABLE pokemon_forms (
    id                              INTEGER PRIMARY KEY,
    identifier                      TEXT NOT NULL UNIQUE,
    form_identifier                 TEXT,
    pokemon_id                      INTEGER NOT NULL REFERENCES pokemon(id),
    introduced_in_version_group_id  INTEGER REFERENCES version_groups(id),
    is_default                      INTEGER NOT NULL,
    is_battle_only                  INTEGER NOT NULL,
    is_mega                         INTEGER NOT NULL,
    form_order                      INTEGER,
    sort_order                      INTEGER,
    form_name_en                    TEXT,
    form_name_es                    TEXT,
    full_name_en                    TEXT,
    full_name_es                    TEXT
);

CREATE TABLE pokemon_types (
    pokemon_id  INTEGER NOT NULL REFERENCES pokemon(id),
    type_id     INTEGER NOT NULL REFERENCES types(id),
    slot        INTEGER NOT NULL,
    PRIMARY KEY (pokemon_id, slot)
);

-- Tipos anteriores (válidos HASTA generation_id inclusive), p. ej. Clefairy era Normal hasta gen 5
CREATE TABLE pokemon_types_past (
    pokemon_id     INTEGER NOT NULL REFERENCES pokemon(id),
    generation_id  INTEGER NOT NULL REFERENCES generations(id),
    type_id        INTEGER NOT NULL REFERENCES types(id),
    slot           INTEGER NOT NULL,
    PRIMARY KEY (pokemon_id, generation_id, slot)
);

CREATE TABLE pokemon_abilities (
    pokemon_id  INTEGER NOT NULL REFERENCES pokemon(id),
    ability_id  INTEGER NOT NULL REFERENCES abilities(id),
    is_hidden   INTEGER NOT NULL,
    slot        INTEGER NOT NULL,
    PRIMARY KEY (pokemon_id, slot)
);

-- Habilidades anteriores (válidas HASTA generation_id inclusive). ability_id NULL = ese hueco no existía
CREATE TABLE pokemon_abilities_past (
    pokemon_id     INTEGER NOT NULL REFERENCES pokemon(id),
    generation_id  INTEGER NOT NULL REFERENCES generations(id),
    ability_id     INTEGER REFERENCES abilities(id),
    is_hidden      INTEGER NOT NULL,
    slot           INTEGER NOT NULL,
    PRIMARY KEY (pokemon_id, generation_id, slot)
);

CREATE TABLE pokemon_stats (
    pokemon_id  INTEGER NOT NULL REFERENCES pokemon(id),
    stat_id     INTEGER NOT NULL REFERENCES stats(id),
    base_stat   INTEGER NOT NULL,
    effort      INTEGER NOT NULL,     -- EV que otorga
    PRIMARY KEY (pokemon_id, stat_id)
);

CREATE TABLE pokemon_stats_past (
    pokemon_id     INTEGER NOT NULL REFERENCES pokemon(id),
    generation_id  INTEGER NOT NULL REFERENCES generations(id),
    stat_id        INTEGER NOT NULL REFERENCES stats(id),
    base_stat      INTEGER,
    effort         INTEGER,
    PRIMARY KEY (pokemon_id, generation_id, stat_id)
);

-- Objetos que pueden llevar los Pokémon salvajes
CREATE TABLE pokemon_items (
    pokemon_id  INTEGER NOT NULL REFERENCES pokemon(id),
    version_id  INTEGER NOT NULL REFERENCES versions(id),
    item_id     INTEGER NOT NULL REFERENCES items(id),
    rarity      INTEGER NOT NULL,
    PRIMARY KEY (pokemon_id, version_id, item_id)
);

CREATE TABLE pokemon_move_methods (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,   -- level-up, machine, egg, tutor, ...
    name_en         TEXT,
    name_es         TEXT,
    description_en  TEXT
);

-- Aprendizaje de movimientos por juego
CREATE TABLE pokemon_moves (
    pokemon_id               INTEGER NOT NULL REFERENCES pokemon(id),
    version_group_id         INTEGER NOT NULL REFERENCES version_groups(id),
    move_id                  INTEGER NOT NULL REFERENCES moves(id),
    pokemon_move_method_id   INTEGER NOT NULL REFERENCES pokemon_move_methods(id),
    level                    INTEGER NOT NULL,   -- 0 si no se aprende por nivel
    sort_order               INTEGER,
    mastery                  INTEGER             -- nivel de dominio (Leyendas: Arceus)
);

-- ---------------------------------------------------------------------
--  Evolución
-- ---------------------------------------------------------------------
CREATE TABLE evolution_triggers (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE genders (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE
);

-- Ubicaciones (también las usa la evolución por lugar)
CREATE TABLE locations (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    region_id   INTEGER REFERENCES regions(id),
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE location_areas (
    id           INTEGER PRIMARY KEY,
    location_id  INTEGER NOT NULL REFERENCES locations(id),
    game_index   INTEGER,
    identifier   TEXT,
    name_en      TEXT
);

-- Cada fila = una forma de evolucionar a evolved_species_id
CREATE TABLE pokemon_evolution (
    id                       INTEGER PRIMARY KEY,
    evolved_species_id       INTEGER NOT NULL REFERENCES pokemon_species(id),
    evolution_trigger_id     INTEGER NOT NULL REFERENCES evolution_triggers(id),
    version_group_id         INTEGER REFERENCES version_groups(id),
    is_default               INTEGER,
    trigger_item_id          INTEGER REFERENCES items(id),
    minimum_level            INTEGER,
    gender_id                INTEGER REFERENCES genders(id),
    location_id              INTEGER REFERENCES locations(id),
    held_item_id             INTEGER REFERENCES items(id),
    time_of_day              TEXT,
    known_move_id            INTEGER REFERENCES moves(id),
    known_move_type_id       INTEGER REFERENCES types(id),
    minimum_happiness        INTEGER,
    minimum_beauty           INTEGER,
    minimum_affection        INTEGER,
    relative_physical_stats  INTEGER,   -- 1: Atq>Def, 0: igual, -1: Atq<Def
    party_species_id         INTEGER REFERENCES pokemon_species(id),
    party_type_id            INTEGER REFERENCES types(id),
    trade_species_id         INTEGER REFERENCES pokemon_species(id),
    needs_overworld_rain     INTEGER,
    turn_upside_down         INTEGER,
    needs_multiplayer        INTEGER,
    near_special_rock        INTEGER,
    region_id                INTEGER REFERENCES regions(id),
    required_pokemon_form_id INTEGER REFERENCES pokemon_forms(id),
    evolved_pokemon_form_id  INTEGER REFERENCES pokemon_forms(id),
    used_move_id             INTEGER REFERENCES moves(id),
    minimum_move_count       INTEGER,
    minimum_steps            INTEGER,
    minimum_damage_taken     INTEGER,
    nature_bitmask           INTEGER,
    condition_expression     TEXT,
    percentage_chance        INTEGER
);

-- ---------------------------------------------------------------------
--  Pokédex regionales
-- ---------------------------------------------------------------------
CREATE TABLE pokedexes (
    id              INTEGER PRIMARY KEY,
    identifier      TEXT NOT NULL UNIQUE,
    region_id       INTEGER REFERENCES regions(id),
    is_main_series  INTEGER NOT NULL,
    name_en         TEXT,
    name_es         TEXT,
    description_en  TEXT
);

CREATE TABLE pokemon_dex_numbers (
    species_id      INTEGER NOT NULL REFERENCES pokemon_species(id),
    pokedex_id      INTEGER NOT NULL REFERENCES pokedexes(id),
    pokedex_number  INTEGER NOT NULL,
    PRIMARY KEY (species_id, pokedex_id)
);

-- ---------------------------------------------------------------------
--  Encuentros salvajes
-- ---------------------------------------------------------------------
CREATE TABLE encounter_methods (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,   -- walk, surf, old-rod, ...
    sort_order  INTEGER,
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE encounter_slots (
    id                   INTEGER PRIMARY KEY,
    version_group_id     INTEGER NOT NULL REFERENCES version_groups(id),
    encounter_method_id  INTEGER NOT NULL REFERENCES encounter_methods(id),
    slot                 INTEGER,
    rarity               INTEGER     -- % de aparición del hueco
);

CREATE TABLE encounters (
    id                 INTEGER PRIMARY KEY,
    version_id         INTEGER NOT NULL REFERENCES versions(id),
    location_area_id   INTEGER NOT NULL REFERENCES location_areas(id),
    encounter_slot_id  INTEGER NOT NULL REFERENCES encounter_slots(id),
    pokemon_id         INTEGER NOT NULL REFERENCES pokemon(id),
    min_level          INTEGER NOT NULL,
    max_level          INTEGER NOT NULL
);

CREATE TABLE encounter_conditions (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,   -- time, season, swarm, ...
    name_en     TEXT,
    name_es     TEXT
);

CREATE TABLE encounter_condition_values (
    id                      INTEGER PRIMARY KEY,
    encounter_condition_id  INTEGER NOT NULL REFERENCES encounter_conditions(id),
    identifier              TEXT NOT NULL UNIQUE,
    is_default              INTEGER NOT NULL,
    name_en                 TEXT,
    name_es                 TEXT
);

CREATE TABLE encounter_condition_value_map (
    encounter_id                  INTEGER NOT NULL REFERENCES encounters(id),
    encounter_condition_value_id  INTEGER NOT NULL REFERENCES encounter_condition_values(id),
    PRIMARY KEY (encounter_id, encounter_condition_value_id)
);

-- =====================================================================
--  PARTE 2: LÍDERES DE GIMNASIO
-- =====================================================================

-- La persona (Brock aparece en RBY, GSC, FRLG, HGSS, LGPE...)
CREATE TABLE gym_leaders (
    id          INTEGER PRIMARY KEY,
    identifier  TEXT NOT NULL UNIQUE,
    name_en     TEXT NOT NULL,
    name_es     TEXT NOT NULL,
    gender      TEXT CHECK (gender IN ('M','F','MF'))   -- 'MF' = pareja (Vito y Leti)
);

-- El gimnasio físico (Gimnasio de Plateada, ...)
CREATE TABLE gyms (
    id           INTEGER PRIMARY KEY,
    identifier   TEXT NOT NULL UNIQUE,
    city_en      TEXT NOT NULL,
    location_id  INTEGER REFERENCES locations(id),   -- ciudad en `locations` (nombre en español allí)
    region_id    INTEGER NOT NULL REFERENCES regions(id)
);

-- Un combate concreto contra un líder en un juego concreto
CREATE TABLE gym_battles (
    id                 INTEGER PRIMARY KEY,
    gym_id             INTEGER NOT NULL REFERENCES gyms(id),
    gym_leader_id      INTEGER NOT NULL REFERENCES gym_leaders(id),
    version_group_id   INTEGER NOT NULL REFERENCES version_groups(id),
    version_id         INTEGER REFERENCES versions(id),   -- NULL = en todas las versiones del grupo
    gym_order          INTEGER NOT NULL,                  -- nº de medalla / orden recomendado en ese juego
    badge_en           TEXT NOT NULL,                     -- medalla que se obtiene (en Galar depende de la versión)
    specialty_type_id  INTEGER REFERENCES types(id),      -- NULL = sin tipo (Azul en GSC/HGSS)
    battle_kind        TEXT NOT NULL CHECK (battle_kind IN ('gym','rematch','challenge_mode','special')),
    rematch_number     INTEGER,                           -- 1, 2... para battle_kind = 'rematch'
    is_double_battle   INTEGER NOT NULL DEFAULT 0,
    notes              TEXT,
    source             TEXT NOT NULL,                     -- p. ej. 'pret/pokeemerald', 'serebii/swordshield'
    source_ref         TEXT                               -- identificador del entrenador en la fuente
);

-- Objetos de la mochila que el líder usa en combate (Pociones, Hiperpociones...)
CREATE TABLE gym_battle_items (
    gym_battle_id  INTEGER NOT NULL REFERENCES gym_battles(id),
    slot           INTEGER NOT NULL,
    item_id        INTEGER NOT NULL REFERENCES items(id),
    PRIMARY KEY (gym_battle_id, slot)
);

CREATE TABLE gym_battle_pokemon (
    id                 INTEGER PRIMARY KEY,
    gym_battle_id      INTEGER NOT NULL REFERENCES gym_battles(id),
    slot               INTEGER NOT NULL,                  -- orden en el equipo (1 = sale primero)
    pokemon_id         INTEGER NOT NULL REFERENCES pokemon(id),
    level              INTEGER NOT NULL,
    gender             TEXT CHECK (gender IN ('M','F')),
    held_item_id       INTEGER REFERENCES items(id),
    ability_id         INTEGER REFERENCES abilities(id),
    -- Origen de la habilidad:
    --   'game_data' = indicada por la fuente
    --   'only_option' = la especie solo tenía una habilidad posible en esa generación
    --   NULL = no aplicable (gen 1-2, Let's Go) o indeterminada
    ability_source     TEXT CHECK (ability_source IN ('game_data','only_option')),
    -- Origen de los movimientos:
    --   'game_data' = moveset explícito del juego
    --   'default_levelup' = el juego no define moveset: últimos 4 movimientos por nivel (como hace el juego)
    --   'mixed' = por defecto + movimientos especiales sobreescritos (gen 1)
    moves_source       TEXT NOT NULL CHECK (moves_source IN ('game_data','default_levelup','mixed')),
    tera_type_id       INTEGER REFERENCES types(id),     -- gen 9
    can_dynamax        INTEGER NOT NULL DEFAULT 0,        -- gen 8: es el Pokémon que dinamaximiza
    is_gigantamax      INTEGER NOT NULL DEFAULT 0,        -- gen 8: gigamaximiza
    difficulty         INTEGER,                           -- IV / "difficulty" interno (gen 3-4, escala 0-255)
    UNIQUE (gym_battle_id, slot)
);

CREATE TABLE gym_battle_pokemon_moves (
    gym_battle_pokemon_id  INTEGER NOT NULL REFERENCES gym_battle_pokemon(id),
    slot                   INTEGER NOT NULL CHECK (slot BETWEEN 1 AND 4),
    move_id                INTEGER NOT NULL REFERENCES moves(id),
    PRIMARY KEY (gym_battle_pokemon_id, slot)
);
