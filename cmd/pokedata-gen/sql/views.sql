-- =====================================================================
--  Índices y vistas de conveniencia
-- =====================================================================

CREATE INDEX idx_pokemon_species            ON pokemon(species_id);
CREATE INDEX idx_pokemon_forms_pokemon      ON pokemon_forms(pokemon_id);
CREATE INDEX idx_pokemon_types_type         ON pokemon_types(type_id);
CREATE INDEX idx_pokemon_abilities_ability  ON pokemon_abilities(ability_id);
CREATE INDEX idx_pokemon_moves_pokemon      ON pokemon_moves(pokemon_id, version_group_id);
CREATE INDEX idx_pokemon_moves_move         ON pokemon_moves(move_id);
CREATE INDEX idx_moves_type                 ON moves(type_id);
CREATE INDEX idx_species_chain              ON pokemon_species(evolution_chain_id);
CREATE INDEX idx_evolution_species          ON pokemon_evolution(evolved_species_id);
CREATE INDEX idx_encounters_pokemon         ON encounters(pokemon_id);
CREATE INDEX idx_encounters_area            ON encounters(location_area_id);
CREATE INDEX idx_machines_move              ON machines(move_id);
CREATE INDEX idx_gym_battles_leader         ON gym_battles(gym_leader_id);
CREATE INDEX idx_gym_battles_vg             ON gym_battles(version_group_id);
CREATE INDEX idx_gym_battle_pokemon_battle  ON gym_battle_pokemon(gym_battle_id);
CREATE INDEX idx_gym_battle_pokemon_pokemon ON gym_battle_pokemon(pokemon_id);

-- Ficha resumida de cada Pokémon (todas las formas)
CREATE VIEW v_pokemon AS
SELECT p.id,
       p.identifier,
       s.id                                   AS national_dex,
       s.name_es, s.name_en,
       s.generation_id,
       t1.name_es                             AS type1_es,
       t2.name_es                             AS type2_es,
       t1.identifier                          AS type1,
       t2.identifier                          AS type2,
       MAX(CASE WHEN st.stat_id = 1 THEN st.base_stat END) AS hp,
       MAX(CASE WHEN st.stat_id = 2 THEN st.base_stat END) AS attack,
       MAX(CASE WHEN st.stat_id = 3 THEN st.base_stat END) AS defense,
       MAX(CASE WHEN st.stat_id = 4 THEN st.base_stat END) AS sp_attack,
       MAX(CASE WHEN st.stat_id = 5 THEN st.base_stat END) AS sp_defense,
       MAX(CASE WHEN st.stat_id = 6 THEN st.base_stat END) AS speed,
       SUM(st.base_stat)                      AS total,
       p.height_dm / 10.0                     AS height_m,
       p.weight_hg / 10.0                     AS weight_kg,
       s.is_legendary, s.is_mythical,
       p.is_default
FROM pokemon p
JOIN pokemon_species s      ON s.id = p.species_id
JOIN pokemon_types pt1      ON pt1.pokemon_id = p.id AND pt1.slot = 1
JOIN types t1               ON t1.id = pt1.type_id
LEFT JOIN pokemon_types pt2 ON pt2.pokemon_id = p.id AND pt2.slot = 2
LEFT JOIN types t2          ON t2.id = pt2.type_id
LEFT JOIN pokemon_stats st  ON st.pokemon_id = p.id
GROUP BY p.id;

-- Habilidades de cada Pokémon con nombre
CREATE VIEW v_pokemon_abilities AS
SELECT p.id AS pokemon_id, p.identifier AS pokemon, pa.slot, pa.is_hidden,
       a.identifier AS ability, a.name_es, a.name_en, a.short_effect_en
FROM pokemon_abilities pa
JOIN pokemon p   ON p.id = pa.pokemon_id
JOIN abilities a ON a.id = pa.ability_id;

-- Tabla de tipos legible
CREATE VIEW v_type_chart AS
SELECT ta.name_es AS attacking_es, td.name_es AS defending_es,
       ta.identifier AS attacking, td.identifier AS defending,
       te.multiplier
FROM type_efficacy te
JOIN types ta ON ta.id = te.attacking_type_id
JOIN types td ON td.id = te.defending_type_id;

-- Movimientos que aprende cada Pokémon en cada juego
CREATE VIEW v_pokemon_learnset AS
SELECT p.identifier AS pokemon, vg.identifier AS version_group,
       mm.identifier AS method, pm.level,
       m.identifier AS move, m.name_es AS move_es, t.identifier AS move_type,
       m.power, m.accuracy, m.pp
FROM pokemon_moves pm
JOIN pokemon p               ON p.id = pm.pokemon_id
JOIN version_groups vg       ON vg.id = pm.version_group_id
JOIN pokemon_move_methods mm ON mm.id = pm.pokemon_move_method_id
JOIN moves m                 ON m.id = pm.move_id
JOIN types t                 ON t.id = m.type_id;

-- Un registro por combate de gimnasio
CREATE VIEW v_gym_battles AS
SELECT b.id              AS gym_battle_id,
       g.id              AS generation,
       vg.identifier     AS version_group,
       v.identifier      AS version,
       r.identifier      AS region,
       gy.city_en,
       loc.name_es       AS city_es,
       b.gym_order,
       b.badge_en,
       l.name_en         AS leader_en,
       l.name_es         AS leader_es,
       t.name_es         AS specialty_type_es,
       t.identifier      AS specialty_type,
       b.battle_kind,
       b.rematch_number,
       b.is_double_battle,
       (SELECT COUNT(*) FROM gym_battle_pokemon x WHERE x.gym_battle_id = b.id) AS team_size,
       (SELECT MAX(level) FROM gym_battle_pokemon x WHERE x.gym_battle_id = b.id) AS max_level,
       b.notes,
       b.source
FROM gym_battles b
JOIN gyms gy           ON gy.id = b.gym_id
JOIN regions r         ON r.id = gy.region_id
LEFT JOIN locations loc ON loc.id = gy.location_id
JOIN gym_leaders l     ON l.id = b.gym_leader_id
JOIN version_groups vg ON vg.id = b.version_group_id
JOIN generations g     ON g.id = vg.generation_id
LEFT JOIN versions v   ON v.id = b.version_id
LEFT JOIN types t      ON t.id = b.specialty_type_id;

-- Equipos completos de los líderes: un registro por Pokémon
CREATE VIEW v_gym_teams AS
SELECT b.gym_battle_id, b.generation, b.version_group, b.version, b.region,
       b.city_en, b.gym_order, b.leader_en, b.leader_es, b.specialty_type,
       b.battle_kind, b.rematch_number,
       bp.slot,
       p.identifier      AS pokemon,
       s.name_es         AS pokemon_es,
       s.name_en         AS pokemon_en,
       bp.level,
       bp.gender,
       a.name_es         AS ability_es,
       a.name_en         AS ability_en,
       bp.ability_source,
       i.name_es         AS held_item_es,
       i.name_en         AS held_item_en,
       tt.identifier     AS tera_type,
       bp.can_dynamax,
       bp.is_gigantamax,
       (SELECT GROUP_CONCAT(name_es, ', ') FROM (
            SELECT m.name_es FROM gym_battle_pokemon_moves bm
            JOIN moves m ON m.id = bm.move_id
            WHERE bm.gym_battle_pokemon_id = bp.id ORDER BY bm.slot)) AS moves_es,
       (SELECT GROUP_CONCAT(name_en, ', ') FROM (
            SELECT m.name_en FROM gym_battle_pokemon_moves bm
            JOIN moves m ON m.id = bm.move_id
            WHERE bm.gym_battle_pokemon_id = bp.id ORDER BY bm.slot)) AS moves_en,
       bp.moves_source
FROM v_gym_battles b
JOIN gym_battle_pokemon bp ON bp.gym_battle_id = b.gym_battle_id
JOIN pokemon p             ON p.id = bp.pokemon_id
JOIN pokemon_species s     ON s.id = p.species_id
LEFT JOIN abilities a      ON a.id = bp.ability_id
LEFT JOIN items i          ON i.id = bp.held_item_id
LEFT JOIN types tt         ON tt.id = bp.tera_type_id;
