-- =====================================================================
--  pokedata — esquema MySQL / MariaDB
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
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE generations (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    main_region_id  INT REFERENCES regions(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE version_groups (
    id             INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    generation_id  INT NOT NULL REFERENCES generations(id),
    sort_order     INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE version_group_regions (
    version_group_id  INT NOT NULL REFERENCES version_groups(id),
    region_id         INT NOT NULL REFERENCES regions(id),
    PRIMARY KEY (version_group_id, region_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE versions (
    id                INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    version_group_id  INT NOT NULL REFERENCES version_groups(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Tipos y estadísticas
-- ---------------------------------------------------------------------
CREATE TABLE move_damage_classes (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,       -- physical / special / status
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE types (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    generation_id    INT NOT NULL REFERENCES generations(id),
    damage_class_id  INT REFERENCES move_damage_classes(id),  -- clase de daño del tipo antes de gen 4
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabla de efectividades actual. multiplier: 0, 0.5, 1, 2
CREATE TABLE type_efficacy (
    attacking_type_id  INT NOT NULL REFERENCES types(id),
    defending_type_id  INT NOT NULL REFERENCES types(id),
    multiplier         DOUBLE NOT NULL,
    PRIMARY KEY (attacking_type_id, defending_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Efectividades que eran distintas en generaciones pasadas (válidas HASTA generation_id inclusive)
CREATE TABLE type_efficacy_past (
    attacking_type_id  INT NOT NULL REFERENCES types(id),
    defending_type_id  INT NOT NULL REFERENCES types(id),
    multiplier         DOUBLE NOT NULL,
    generation_id      INT NOT NULL REFERENCES generations(id),
    PRIMARY KEY (attacking_type_id, defending_type_id, generation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE stats (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,      -- hp, attack, defense, special-attack, ...
    damage_class_id  INT REFERENCES move_damage_classes(id),
    is_battle_only   INT NOT NULL,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sabores / tipos de concurso (los usan naturalezas y bayas)
CREATE TABLE contest_types (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    flavor_en VARCHAR(255),
    flavor_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE natures (
    id                 INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    increased_stat_id  INT NOT NULL REFERENCES stats(id),
    decreased_stat_id  INT NOT NULL REFERENCES stats(id),
    likes_flavor_id    INT NOT NULL REFERENCES contest_types(id),
    hates_flavor_id    INT NOT NULL REFERENCES contest_types(id),
    game_index         INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE characteristics (
    id          INT PRIMARY KEY,
    stat_id     INT NOT NULL REFERENCES stats(id),
    gene_mod_5  INT NOT NULL,
    message_en  TEXT,
    message_es  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Habilidades
-- ---------------------------------------------------------------------
CREATE TABLE abilities (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    generation_id    INT NOT NULL REFERENCES generations(id),
    is_main_series   INT NOT NULL,
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,   -- texto del juego más reciente
    flavor_text_es   TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Objetos
-- ---------------------------------------------------------------------
CREATE TABLE item_pockets (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE item_categories (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    pocket_id   INT NOT NULL REFERENCES item_pockets(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE item_fling_effects (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    effect_en   TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE items (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    category_id      INT NOT NULL REFERENCES item_categories(id),
    cost             INT,
    fling_power      INT,
    fling_effect_id  INT REFERENCES item_fling_effects(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,
    flavor_text_es   TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE item_flags (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE item_flag_map (
    item_id       INT NOT NULL REFERENCES items(id),
    item_flag_id  INT NOT NULL REFERENCES item_flags(id),
    PRIMARY KEY (item_id, item_flag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE item_flavor_text (
    item_id           INT NOT NULL REFERENCES items(id),
    version_group_id  INT NOT NULL REFERENCES version_groups(id),
    language VARCHAR(10) NOT NULL CHECK (language IN ('en','es')),
    flavor_text       TEXT NOT NULL,
    PRIMARY KEY (item_id, version_group_id, language)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE berry_firmness (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE berries (
    id                    INT PRIMARY KEY,
    item_id               INT NOT NULL UNIQUE REFERENCES items(id),
    firmness_id           INT REFERENCES berry_firmness(id),
    natural_gift_power    INT,
    natural_gift_type_id  INT REFERENCES types(id),
    size_mm               INT,
    max_harvest           INT,
    growth_time           INT,
    soil_dryness          INT,
    smoothness            INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE berry_flavors (
    berry_id         INT NOT NULL REFERENCES berries(id),
    contest_type_id  INT NOT NULL REFERENCES contest_types(id),
    flavor           INT NOT NULL,
    PRIMARY KEY (berry_id, contest_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Movimientos
-- ---------------------------------------------------------------------
CREATE TABLE move_targets (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_meta_categories (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_meta_ailments (
    id          INT PRIMARY KEY,   -- -1 = desconocido, 0 = ninguno
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE moves (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    generation_id    INT NOT NULL REFERENCES generations(id),
    type_id          INT NOT NULL REFERENCES types(id),
    damage_class_id  INT REFERENCES move_damage_classes(id),
    power            INT,
    pp               INT,
    accuracy         INT,
    priority         INT NOT NULL,
    target_id        INT REFERENCES move_targets(id),
    effect_chance    INT,
    contest_type_id  INT REFERENCES contest_types(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    short_effect_en  TEXT,
    effect_en        TEXT,
    flavor_text_en   TEXT,
    flavor_text_es   TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Metadatos de batalla del movimiento (1:1 con moves)
CREATE TABLE move_meta (
    move_id           INT PRIMARY KEY REFERENCES moves(id),
    meta_category_id  INT REFERENCES move_meta_categories(id),
    meta_ailment_id   INT REFERENCES move_meta_ailments(id),
    min_hits          INT,
    max_hits          INT,
    min_turns         INT,
    max_turns         INT,
    drain             INT,   -- % de daño drenado (negativo = retroceso)
    healing           INT,   -- % de PS curados
    crit_rate         INT,   -- etapas extra de crítico
    ailment_chance    INT,
    flinch_chance     INT,
    stat_chance       INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_meta_stat_changes (
    move_id  INT NOT NULL REFERENCES moves(id),
    stat_id  INT NOT NULL REFERENCES stats(id),
    `change` INT NOT NULL,
    PRIMARY KEY (move_id, stat_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_flags (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,   -- contact, sound, punch, ...
    name_en VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_flag_map (
    move_id       INT NOT NULL REFERENCES moves(id),
    move_flag_id  INT NOT NULL REFERENCES move_flags(id),
    PRIMARY KEY (move_id, move_flag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Valores antiguos de un movimiento (válidos ANTES del version_group_id indicado)
CREATE TABLE move_changelog (
    move_id                     INT NOT NULL REFERENCES moves(id),
    changed_in_version_group_id INT NOT NULL REFERENCES version_groups(id),
    type_id                     INT REFERENCES types(id),
    power                       INT,
    pp                          INT,
    accuracy                    INT,
    priority                    INT,
    target_id                   INT REFERENCES move_targets(id),
    effect_chance               INT,
    PRIMARY KEY (move_id, changed_in_version_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE move_flavor_text (
    move_id           INT NOT NULL REFERENCES moves(id),
    version_group_id  INT NOT NULL REFERENCES version_groups(id),
    language VARCHAR(10) NOT NULL CHECK (language IN ('en','es')),
    flavor_text       TEXT NOT NULL,
    PRIMARY KEY (move_id, version_group_id, language)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- MT/MO/DT por juego
CREATE TABLE machines (
    machine_number    INT NOT NULL,
    version_group_id  INT NOT NULL REFERENCES version_groups(id),
    item_id           INT NOT NULL REFERENCES items(id),
    move_id           INT NOT NULL REFERENCES moves(id),
    PRIMARY KEY (version_group_id, item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Especies y Pokémon
-- ---------------------------------------------------------------------
CREATE TABLE growth_rates (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    formula     TEXT,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE experience (
    growth_rate_id  INT NOT NULL REFERENCES growth_rates(id),
    level           INT NOT NULL,
    experience      INT NOT NULL,
    PRIMARY KEY (growth_rate_id, level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE egg_groups (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_colors (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_shapes (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_habitats (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE evolution_chains (
    id                     INT PRIMARY KEY,
    baby_trigger_item_id   INT REFERENCES items(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Especie: lo común a todas las formas (Pikachu, Charizard, ...)
CREATE TABLE pokemon_species (
    id                       INT PRIMARY KEY,         -- = nº de la Pokédex Nacional
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    genus_en VARCHAR(255),
    genus_es VARCHAR(255),
    generation_id            INT NOT NULL REFERENCES generations(id),
    evolves_from_species_id  INT REFERENCES pokemon_species(id),
    evolution_chain_id       INT REFERENCES evolution_chains(id),
    color_id                 INT REFERENCES pokemon_colors(id),
    shape_id                 INT REFERENCES pokemon_shapes(id),
    habitat_id               INT REFERENCES pokemon_habitats(id),
    gender_rate              INT NOT NULL,  -- -1 sin género; si no, octavos de probabilidad de hembra
    capture_rate             INT NOT NULL,
    base_happiness           INT,
    is_baby                  INT NOT NULL,
    hatch_counter            INT,
    has_gender_differences   INT NOT NULL,
    growth_rate_id           INT NOT NULL REFERENCES growth_rates(id),
    forms_switchable         INT NOT NULL,
    is_legendary             INT NOT NULL,
    is_mythical              INT NOT NULL,
    sort_order               INT,
    flavor_text_en           TEXT,   -- entrada de Pokédex más reciente
    flavor_text_es           TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_species_flavor_text (
    species_id   INT NOT NULL REFERENCES pokemon_species(id),
    version_id   INT NOT NULL REFERENCES versions(id),
    language VARCHAR(10) NOT NULL CHECK (language IN ('en','es')),
    flavor_text  TEXT NOT NULL,
    PRIMARY KEY (species_id, version_id, language)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_egg_groups (
    species_id    INT NOT NULL REFERENCES pokemon_species(id),
    egg_group_id  INT NOT NULL REFERENCES egg_groups(id),
    PRIMARY KEY (species_id, egg_group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pokémon "jugable": cada forma con estadísticas/tipos propios (charizard, charizard-mega-x, ...)
CREATE TABLE pokemon (
    id               INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    species_id       INT NOT NULL REFERENCES pokemon_species(id),
    height_dm        INT,
    weight_hg        INT,
    base_experience  INT,
    sort_order       INT,
    is_default       INT NOT NULL      -- 1 = forma base de la especie
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_forms (
    id                              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    form_identifier VARCHAR(100),
    pokemon_id                      INT NOT NULL REFERENCES pokemon(id),
    introduced_in_version_group_id  INT REFERENCES version_groups(id),
    is_default                      INT NOT NULL,
    is_battle_only                  INT NOT NULL,
    is_mega                         INT NOT NULL,
    form_order                      INT,
    sort_order                      INT,
    form_name_en VARCHAR(255),
    form_name_es VARCHAR(255),
    full_name_en VARCHAR(255),
    full_name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_types (
    pokemon_id  INT NOT NULL REFERENCES pokemon(id),
    type_id     INT NOT NULL REFERENCES types(id),
    slot        INT NOT NULL,
    PRIMARY KEY (pokemon_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tipos anteriores (válidos HASTA generation_id inclusive), p. ej. Clefairy era Normal hasta gen 5
CREATE TABLE pokemon_types_past (
    pokemon_id     INT NOT NULL REFERENCES pokemon(id),
    generation_id  INT NOT NULL REFERENCES generations(id),
    type_id        INT NOT NULL REFERENCES types(id),
    slot           INT NOT NULL,
    PRIMARY KEY (pokemon_id, generation_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_abilities (
    pokemon_id  INT NOT NULL REFERENCES pokemon(id),
    ability_id  INT NOT NULL REFERENCES abilities(id),
    is_hidden   INT NOT NULL,
    slot        INT NOT NULL,
    PRIMARY KEY (pokemon_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Habilidades anteriores (válidas HASTA generation_id inclusive). ability_id NULL = ese hueco no existía
CREATE TABLE pokemon_abilities_past (
    pokemon_id     INT NOT NULL REFERENCES pokemon(id),
    generation_id  INT NOT NULL REFERENCES generations(id),
    ability_id     INT REFERENCES abilities(id),
    is_hidden      INT NOT NULL,
    slot           INT NOT NULL,
    PRIMARY KEY (pokemon_id, generation_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_stats (
    pokemon_id  INT NOT NULL REFERENCES pokemon(id),
    stat_id     INT NOT NULL REFERENCES stats(id),
    base_stat   INT NOT NULL,
    effort      INT NOT NULL,     -- EV que otorga
    PRIMARY KEY (pokemon_id, stat_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_stats_past (
    pokemon_id     INT NOT NULL REFERENCES pokemon(id),
    generation_id  INT NOT NULL REFERENCES generations(id),
    stat_id        INT NOT NULL REFERENCES stats(id),
    base_stat      INT,
    effort         INT,
    PRIMARY KEY (pokemon_id, generation_id, stat_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Objetos que pueden llevar los Pokémon salvajes
CREATE TABLE pokemon_items (
    pokemon_id  INT NOT NULL REFERENCES pokemon(id),
    version_id  INT NOT NULL REFERENCES versions(id),
    item_id     INT NOT NULL REFERENCES items(id),
    rarity      INT NOT NULL,
    PRIMARY KEY (pokemon_id, version_id, item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_move_methods (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,   -- level-up, machine, egg, tutor, ...
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Aprendizaje de movimientos por juego
CREATE TABLE pokemon_moves (
    pokemon_id               INT NOT NULL REFERENCES pokemon(id),
    version_group_id         INT NOT NULL REFERENCES version_groups(id),
    move_id                  INT NOT NULL REFERENCES moves(id),
    pokemon_move_method_id   INT NOT NULL REFERENCES pokemon_move_methods(id),
    level                    INT NOT NULL,   -- 0 si no se aprende por nivel
    sort_order               INT,
    mastery                  INT             -- nivel de dominio (Leyendas: Arceus)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Evolución
-- ---------------------------------------------------------------------
CREATE TABLE evolution_triggers (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE genders (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Ubicaciones (también las usa la evolución por lugar)
CREATE TABLE locations (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    region_id   INT REFERENCES regions(id),
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE location_areas (
    id           INT PRIMARY KEY,
    location_id  INT NOT NULL REFERENCES locations(id),
    game_index   INT,
    identifier VARCHAR(100),
    name_en VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Cada fila = una forma de evolucionar a evolved_species_id
CREATE TABLE pokemon_evolution (
    id                       INT PRIMARY KEY,
    evolved_species_id       INT NOT NULL REFERENCES pokemon_species(id),
    evolution_trigger_id     INT NOT NULL REFERENCES evolution_triggers(id),
    version_group_id         INT REFERENCES version_groups(id),
    is_default               INT,
    trigger_item_id          INT REFERENCES items(id),
    minimum_level            INT,
    gender_id                INT REFERENCES genders(id),
    location_id              INT REFERENCES locations(id),
    held_item_id             INT REFERENCES items(id),
    time_of_day VARCHAR(50),
    known_move_id            INT REFERENCES moves(id),
    known_move_type_id       INT REFERENCES types(id),
    minimum_happiness        INT,
    minimum_beauty           INT,
    minimum_affection        INT,
    relative_physical_stats  INT,   -- 1: Atq>Def, 0: igual, -1: Atq<Def
    party_species_id         INT REFERENCES pokemon_species(id),
    party_type_id            INT REFERENCES types(id),
    trade_species_id         INT REFERENCES pokemon_species(id),
    needs_overworld_rain     INT,
    turn_upside_down         INT,
    needs_multiplayer        INT,
    near_special_rock        INT,
    region_id                INT REFERENCES regions(id),
    required_pokemon_form_id INT REFERENCES pokemon_forms(id),
    evolved_pokemon_form_id  INT REFERENCES pokemon_forms(id),
    used_move_id             INT REFERENCES moves(id),
    minimum_move_count       INT,
    minimum_steps            INT,
    minimum_damage_taken     INT,
    nature_bitmask           INT,
    condition_expression     TEXT,
    percentage_chance        INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Pokédex regionales
-- ---------------------------------------------------------------------
CREATE TABLE pokedexes (
    id              INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    region_id       INT REFERENCES regions(id),
    is_main_series  INT NOT NULL,
    name_en VARCHAR(255),
    name_es VARCHAR(255),
    description_en  TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE pokemon_dex_numbers (
    species_id      INT NOT NULL REFERENCES pokemon_species(id),
    pokedex_id      INT NOT NULL REFERENCES pokedexes(id),
    pokedex_number  INT NOT NULL,
    PRIMARY KEY (species_id, pokedex_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------
--  Encuentros salvajes
-- ---------------------------------------------------------------------
CREATE TABLE encounter_methods (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,   -- walk, surf, old-rod, ...
    sort_order  INT,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE encounter_slots (
    id                   INT PRIMARY KEY,
    version_group_id     INT NOT NULL REFERENCES version_groups(id),
    encounter_method_id  INT NOT NULL REFERENCES encounter_methods(id),
    slot                 INT,
    rarity               INT     -- % de aparición del hueco
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE encounters (
    id                 INT PRIMARY KEY,
    version_id         INT NOT NULL REFERENCES versions(id),
    location_area_id   INT NOT NULL REFERENCES location_areas(id),
    encounter_slot_id  INT NOT NULL REFERENCES encounter_slots(id),
    pokemon_id         INT NOT NULL REFERENCES pokemon(id),
    min_level          INT NOT NULL,
    max_level          INT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE encounter_conditions (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,   -- time, season, swarm, ...
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE encounter_condition_values (
    id                      INT PRIMARY KEY,
    encounter_condition_id  INT NOT NULL REFERENCES encounter_conditions(id),
    identifier VARCHAR(100) NOT NULL UNIQUE,
    is_default              INT NOT NULL,
    name_en VARCHAR(255),
    name_es VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE encounter_condition_value_map (
    encounter_id                  INT NOT NULL REFERENCES encounters(id),
    encounter_condition_value_id  INT NOT NULL REFERENCES encounter_condition_values(id),
    PRIMARY KEY (encounter_id, encounter_condition_value_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================================
--  PARTE 2: LÍDERES DE GIMNASIO
-- =====================================================================

-- La persona (Brock aparece en RBY, GSC, FRLG, HGSS, LGPE...)
CREATE TABLE gym_leaders (
    id          INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    name_en VARCHAR(255) NOT NULL,
    name_es VARCHAR(255) NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('M','F','MF'))   -- 'MF' = pareja (Vito y Leti)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- El gimnasio físico (Gimnasio de Plateada, ...)
CREATE TABLE gyms (
    id           INT PRIMARY KEY,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    city_en VARCHAR(100) NOT NULL,
    location_id  INT REFERENCES locations(id),   -- ciudad en `locations` (nombre en español allí)
    region_id    INT NOT NULL REFERENCES regions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Un combate concreto contra un líder en un juego concreto
CREATE TABLE gym_battles (
    id                 INT PRIMARY KEY,
    gym_id             INT NOT NULL REFERENCES gyms(id),
    gym_leader_id      INT NOT NULL REFERENCES gym_leaders(id),
    version_group_id   INT NOT NULL REFERENCES version_groups(id),
    version_id         INT REFERENCES versions(id),   -- NULL = en todas las versiones del grupo
    gym_order          INT NOT NULL,                  -- nº de medalla / orden recomendado en ese juego
    badge_en VARCHAR(100) NOT NULL,                     -- medalla que se obtiene (en Galar depende de la versión)
    specialty_type_id  INT REFERENCES types(id),      -- NULL = sin tipo (Azul en GSC/HGSS)
    battle_kind VARCHAR(50) NOT NULL CHECK (battle_kind IN ('gym','rematch','challenge_mode','special')),
    rematch_number     INT,                           -- 1, 2... para battle_kind = 'rematch'
    is_double_battle   INT NOT NULL DEFAULT 0,
    notes              TEXT,
    source VARCHAR(100) NOT NULL,                     -- p. ej. 'pret/pokeemerald', 'serebii/swordshield'
    source_ref VARCHAR(100)                               -- identificador del entrenador en la fuente
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Objetos de la mochila que el líder usa en combate (Pociones, Hiperpociones...)
CREATE TABLE gym_battle_items (
    gym_battle_id  INT NOT NULL REFERENCES gym_battles(id),
    slot           INT NOT NULL,
    item_id        INT NOT NULL REFERENCES items(id),
    PRIMARY KEY (gym_battle_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gym_battle_pokemon (
    id                 INT PRIMARY KEY,
    gym_battle_id      INT NOT NULL REFERENCES gym_battles(id),
    slot               INT NOT NULL,                  -- orden en el equipo (1 = sale primero)
    pokemon_id         INT NOT NULL REFERENCES pokemon(id),
    level              INT NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('M','F')),
    held_item_id       INT REFERENCES items(id),
    ability_id         INT REFERENCES abilities(id),
    -- Origen de la habilidad:
    --   'game_data' = indicada por la fuente
    --   'only_option' = la especie solo tenía una habilidad posible en esa generación
    --   NULL = no aplicable (gen 1-2, Let's Go) o indeterminada
    ability_source VARCHAR(50) CHECK (ability_source IN ('game_data','only_option')),
    -- Origen de los movimientos:
    --   'game_data' = moveset explícito del juego
    --   'default_levelup' = el juego no define moveset: últimos 4 movimientos por nivel (como hace el juego)
    --   'mixed' = por defecto + movimientos especiales sobreescritos (gen 1)
    moves_source VARCHAR(50) NOT NULL CHECK (moves_source IN ('game_data','default_levelup','mixed')),
    tera_type_id       INT REFERENCES types(id),     -- gen 9
    can_dynamax        INT NOT NULL DEFAULT 0,        -- gen 8: es el Pokémon que dinamaximiza
    is_gigantamax      INT NOT NULL DEFAULT 0,        -- gen 8: gigamaximiza
    difficulty         INT,                           -- IV / "difficulty" interno (gen 3-4, escala 0-255)
    UNIQUE (gym_battle_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE gym_battle_pokemon_moves (
    gym_battle_pokemon_id  INT NOT NULL REFERENCES gym_battle_pokemon(id),
    slot                   INT NOT NULL CHECK (slot BETWEEN 1 AND 4),
    move_id                INT NOT NULL REFERENCES moves(id),
    PRIMARY KEY (gym_battle_pokemon_id, slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
