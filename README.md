# pokedata

`pokedata-gen` es un binario Go estático (sin cgo) que descarga las fuentes y genera
una base de datos SQLite con todos los datos de Pokémon y los equipos de los
líderes de gimnasio de todas las generaciones. Si se le pide, vuelca también un
script `.sql` equivalente.

```sh
go build -o pokedata-gen ./cmd/pokedata-gen

./pokedata-gen                         # descarga las fuentes y crea pokedata.db
./pokedata-gen -v                      # muestra el progreso de la descarga cada 5s
./pokedata-gen -sql pokedata.sql       # además, escribe el script SQL
./pokedata-gen -offline                # solo con la caché de ./sources
./pokedata-gen -refresh                # vuelve a descargar todas las fuentes
./pokedata-gen -db /data/poke.db -sources /tmp/src -q
```

Con el script: `sqlite3 pokedata.db < pokedata.sql`.

La generación falla, sin dejar un `.db` a medias, si algún nombre de las fuentes
no se puede resolver, si hay claves foráneas rotas o si alguna vista no es válida.

## En una imagen Docker

```dockerfile
FROM golang:1.27 AS pokedata
WORKDIR /src
COPY pokedata/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /pokedata-gen ./cmd/pokedata-gen
# Opcional: generar la BD en la build en vez de en tiempo de ejecución
RUN /pokedata-gen -db /pokedata.db -sources /tmp/sources

FROM gcr.io/distroless/static
COPY --from=pokedata /pokedata-gen /usr/local/bin/pokedata-gen
COPY --from=pokedata /pokedata.db /data/pokedata.db
```

## Fuentes

| Datos | Fuente |
|---|---|
| Pokémon, especies, formas, tipos, movimientos, habilidades, objetos, naturalezas, evoluciones, Pokédex, encuentros… | CSV oficiales de [PokeAPI](https://github.com/PokeAPI/pokeapi) |
| Líderes gen 1-4 (RA/Am, OP/Cr, RZ/E/RFVH, DP/Pt/HGSS) | Descompilaciones de [pret](https://github.com/pret): datos reales de los juegos |
| Líderes gen 5-9 (NB, N2B2, XY, ROZA, Let's Go, EpEs, DBPR, EP) | Páginas de gimnasios de [Serebii](https://www.serebii.net) |

Sol/Luna y Ultrasol/Ultraluna no tienen gimnasios: la generación 7 está
cubierta por Let's Go Pikachu/Eevee.

## Esquema

Está en `cmd/pokedata-gen/sql/schema.sql` (comentado) y las vistas en `views.sql`.
Los nombres van en inglés y español (`name_en`, `name_es`).

Líderes de gimnasio:

```
gym_leaders ─┐
gyms ────────┼─< gym_battles >── version_groups / versions / types
             │        │
             │        ├─< gym_battle_items >── items
             │        └─< gym_battle_pokemon >── pokemon / abilities / items / types(tera)
             │                  └─< gym_battle_pokemon_moves >── moves
```

- `gym_battles.battle_kind`: `gym`, `rematch` (Esmeralda, Platino, HGSS),
  `challenge_mode` (N2B2) o `special` (Volkner en la Zona Combate).
- `gym_battle_pokemon.moves_source`:
  - `game_data`: moveset definido por el juego.
  - `default_levelup`: el juego no lo define y usa los 4 últimos movimientos
    aprendidos por nivel. Se calcula con los datos de aprendizaje de ese juego.
  - `mixed`: lo mismo, pero con los movimientos especiales que la gen 1 sobreescribe.
- `gym_battle_pokemon.ability_source`:
  - `game_data`: la habilidad la indica la fuente.
  - `only_option`: la especie solo podía tener esa habilidad en esa generación.
  - `NULL`: no aplica (gen 1-2 y Let's Go) o depende de la personalidad aleatoria del Pokémon.

Vistas útiles: `v_gym_teams` (un Pokémon por fila, con movimientos en español e
inglés), `v_gym_battles`, `v_pokemon`, `v_pokemon_learnset`, `v_type_chart`.

```sql
SELECT version_group, leader_es, slot, pokemon_es, level, ability_es, held_item_es, moves_es
FROM v_gym_teams
WHERE leader_en = 'Brock' AND battle_kind = 'gym'
ORDER BY gym_battle_id, slot;
```
