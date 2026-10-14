// pokedata-gen descarga las fuentes (PokeAPI, pret, Serebii) y genera una base
// de datos SQLite o MySQL/MariaDB con todos los datos de Pokémon y los equipos de los líderes de
// gimnasio de todas las generaciones. Opcionalmente vuelca también un script .sql.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

type createDBFlag struct {
	enabled bool
	dbName  string
}

func (f *createDBFlag) String() string {
	if f.enabled {
		if f.dbName != "" {
			return f.dbName
		}
		return "true"
	}
	return "false"
}

func (f *createDBFlag) Set(val string) error {
	valLower := strings.ToLower(strings.TrimSpace(val))
	if valLower == "true" || valLower == "1" || valLower == "" {
		f.enabled = true
		return nil
	}
	if valLower == "false" || valLower == "0" {
		f.enabled = false
		return nil
	}
	// Si se pasa un nombre de base de datos directamente (ej: --create-database=mi_db)
	f.enabled = true
	f.dbName = val
	return nil
}

func (f *createDBFlag) IsBoolFlag() bool {
	return true
}

func main() {
	var (
		mode    = flag.String("mode", "sqlite", "motor de base de datos de salida: sqlite o mysql (o mariadb)")
		sources = flag.String("sources", "sources", "directorio de caché de las fuentes descargadas")
		dbPath  = flag.String("db", "pokedata.db", "fichero SQLite de salida (modo sqlite; vacío = no generarlo)")
		sqlPath = flag.String("sql", "", "fichero .sql de salida (vacío = no generarlo)")
		refresh = flag.Bool("refresh", false, "volver a descargar todas las fuentes aunque estén en caché")
		offline = flag.Bool("offline", false, "no descargar nada; fallar si falta alguna fuente")
		quiet   = flag.Bool("q", false, "no mostrar progreso")
		verbose = flag.Bool("v", false, "mostrar progreso de la descarga cada 5 segundos")

		// Parámetros MySQL / MariaDB
		host      = flag.String("host", "127.0.0.1", "host de MySQL/MariaDB (puede incluir :puerto)")
		port      = flag.Int("port", 3306, "puerto de MySQL/MariaDB")
		username  = flag.String("username", "root", "usuario de MySQL/MariaDB")
		userAlias = flag.String("user", "", "alias para -username")
		password  = flag.String("password", "", "contraseña de MySQL/MariaDB en texto plano")
		database  = flag.String("database", "pokedata", "nombre de la base de datos MySQL/MariaDB")
		tlsMode   = flag.String("tls", "preferred", "modo TLS/SSL para MySQL: preferred (por defecto), true, skip-verify o false")
		ssl       = flag.Bool("ssl", false, "forzar conexión SSL/TLS (equivalente a -tls true)")
	)

	var createDB createDBFlag
	flag.Var(&createDB, "create-database", "crear la base de datos MySQL si no existe (admite valor booleano o nombre de bd)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s [opciones]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	logf := func(format string, a ...any) {
		if !*quiet {
			fmt.Fprintf(os.Stderr, format+"\n", a...)
		}
	}

	user := *username
	if *userAlias != "" {
		user = *userAlias
	}

	dbName := *database
	if createDB.dbName != "" {
		dbName = createDB.dbName
	}

	tlsVal := *tlsMode
	if *ssl && (tlsVal == "" || tlsVal == "preferred" || tlsVal == "false") {
		tlsVal = "true"
	}

	mysqlCfg := MySQLConfig{
		Host:           *host,
		Port:           *port,
		User:           user,
		Password:       *password,
		Database:       dbName,
		CreateDatabase: createDB.enabled,
		TLS:            tlsVal,
	}

	if err := run(*mode, *sources, *dbPath, *sqlPath, *refresh, *offline, *verbose, mysqlCfg, logf); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(mode, sources, dbPath, sqlPath string, refresh, offline, verbose bool, mysqlCfg MySQLConfig, logf func(string, ...any)) (err error) {
	modeLower := strings.ToLower(strings.TrimSpace(mode))
	switch modeLower {
	case "sqlite":
		if dbPath == "" && sqlPath == "" {
			return fmt.Errorf("en modo sqlite indica al menos -db o -sql")
		}
	case "mysql", "mariadb":
		if mysqlCfg.Database == "" {
			return fmt.Errorf("en modo mysql debes especificar el nombre de la base de datos (--database)")
		}
	default:
		return fmt.Errorf("modo no válido: %q (usa 'sqlite' o 'mysql')", mode)
	}

	// Los parsers señalan datos inesperados con panic: se convierten en error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	start := time.Now()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := ensureSources(ctx, sources, refresh, offline, verbose, logf); err != nil {
		return err
	}

	ds := NewDataset()
	logf("cargando PokeAPI...")
	L := loadPokeAPI(ds, &Src{dir: sources, cache: map[string]*CSV{}})
	logf("procesando líderes de gimnasio...")
	if err := buildGyms(ds, L, sources, logf); err != nil {
		return err
	}
	tables, err := orderedTables(ds)
	if err != nil {
		return err
	}
	total := 0
	for _, t := range tables {
		total += len(t.Rows)
	}
	logf("%d tablas, %d filas", len(tables), total)

	switch modeLower {
	case "sqlite":
		if sqlPath != "" {
			logf("escribiendo %s...", sqlPath)
			if err := writeSQLiteFile(sqlPath, tables); err != nil {
				return err
			}
		}
		if dbPath != "" {
			logf("escribiendo %s...", dbPath)
			if err := writeSQLite(dbPath, tables, logf); err != nil {
				return err
			}
		}
	case "mysql", "mariadb":
		if sqlPath != "" {
			logf("escribiendo %s...", sqlPath)
			if err := writeMySQLFile(sqlPath, tables); err != nil {
				return err
			}
		}
		logf("conectando y volcando a MySQL/MariaDB (%s:%d/%s)...", mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database)
		if err := writeMySQL(ctx, mysqlCfg, tables, logf); err != nil {
			return err
		}
	}

	logf("listo en %s", time.Since(start).Round(time.Millisecond))
	return nil
}
