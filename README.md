# pokedata

`pokedata-gen` es un binario Go estático (sin cgo) que descarga las fuentes y genera
una base de datos SQLite o MySQL/MariaDB con todos los datos de Pokémon y los equipos de los
líderes de gimnasio de todas las generaciones. Si se le pide, vuelca también un
script `.sql` equivalente adaptado al motor seleccionado.

```sh
go build -o pokedata-gen ./cmd/pokedata-gen

# Modo SQLite (por defecto)
./pokedata-gen                         # descarga las fuentes y crea pokedata.db
./pokedata-gen --mode sqlite           # equivalente al anterior
./pokedata-gen -v                      # muestra el progreso de la descarga cada 5s
./pokedata-gen -sql pokedata.sql       # además, escribe el script SQL para SQLite
./pokedata-gen -offline                # solo con la caché de ./sources
./pokedata-gen -refresh                # vuelve a descargar todas las fuentes
./pokedata-gen -db /data/poke.db -sources /tmp/src -q

# Modo MySQL / MariaDB (instancia local o remota)
./pokedata-gen --mode mysql --host 192.168.1.100 --port 3306 --username root --password mi_pass --database pokedata --create-database
./pokedata-gen --mode mysql --host 127.0.0.1 --username user --password pass --database pokedata
./pokedata-gen --mode mysql --sql pokedata_mysql.sql -offline   # vuelca script SQL para MySQL
```

Con el script:
- SQLite: `sqlite3 pokedata.db < pokedata.sql`
- MySQL / MariaDB: `mysql -u <usuario> -p <base_de_datos> < pokedata_mysql.sql`

Opciones para MySQL / MariaDB:
- `--mode`: `sqlite` (por defecto) o `mysql` (también acepta `mariadb`).
- `--host`: dirección del servidor MySQL (por defecto `127.0.0.1`, admite formato `host:puerto`).
- `--port`: puerto TCP (por defecto `3306`).
- `--username` / `--user`: usuario con permisos en la base de datos (por defecto `root`).
- `--password`: contraseña en texto plano enviada por el proceso que invoca al binario.
- `--database`: nombre de la base de datos de destino (por defecto `pokedata`).
- `--create-database`: crea automáticamente la base de datos en el servidor si aún no existe.
- `--tls`: modo TLS/SSL (`preferred` por defecto, `true`, `skip-verify` o `false`). Resuelve requisitos de conexión segura (`--require_secure_transport=ON`).
- `--ssl`: fuerza TLS/SSL (equivalente a `--tls true`).

La generación falla, sin dejar datos a medias, si algún nombre de las fuentes
no se puede resolver, si hay errores de conexión o si alguna vista no es válida.

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
