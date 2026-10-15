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

	opts := runOptions{
		Mode:    *mode,
		Sources: *sources,
		DBPath:  *dbPath,
		SQLPath: *sqlPath,
		Refresh: *refresh,
		Offline: *offline,
		Verbose: *verbose,
		MySQL:   mysqlCfg,
	}
	if err := run(opts, logf); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runOptions agrupa los parámetros de ejecución obtenidos de la línea de comandos.
type runOptions struct {
	Mode    string
	Sources string
	DBPath  string
	SQLPath string
	Refresh bool
	Offline bool
	Verbose bool
	MySQL   MySQLConfig
}

func run(opts runOptions, logf func(string, ...any)) (err error) {
	modeLower := strings.ToLower(strings.TrimSpace(opts.Mode))
	if err := validateMode(modeLower, opts.Mode, opts.DBPath, opts.SQLPath, opts.MySQL); err != nil {
		return err
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

	if err := ensureSources(ctx, opts.Sources, opts.Refresh, opts.Offline, opts.Verbose, logf); err != nil {
		return err
	}

	tables, err := buildTables(opts.Sources, logf)
	if err != nil {
		return err
	}

	if modeLower == "sqlite" {
		err = writeSQLiteOutputs(opts.DBPath, opts.SQLPath, tables, logf)
	} else {
		err = writeMySQLOutputs(ctx, opts.SQLPath, opts.MySQL, tables, logf)
	}
	if err != nil {
		return err
	}

	logf("listo en %s", time.Since(start).Round(time.Millisecond))
	return nil
}

// validateMode comprueba que el modo es válido y que tiene los parámetros que necesita.
func validateMode(modeLower, mode, dbPath, sqlPath string, mysqlCfg MySQLConfig) error {
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
	return nil
}

// buildTables carga las fuentes y devuelve las tablas ordenadas listas para volcar.
func buildTables(sources string, logf func(string, ...any)) ([]*Table, error) {
	ds := NewDataset()
	logf("cargando PokeAPI...")
	L := loadPokeAPI(ds, &Src{dir: sources, cache: map[string]*CSV{}})
	logf("procesando líderes de gimnasio...")
	if err := buildGyms(ds, L, sources, logf); err != nil {
		return nil, err
	}
	tables, err := orderedTables(ds)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, t := range tables {
		total += len(t.Rows)
	}
	logf("%d tablas, %d filas", len(tables), total)
	return tables, nil
}

func writeSQLiteOutputs(dbPath, sqlPath string, tables []*Table, logf func(string, ...any)) error {
	if sqlPath != "" {
		logf(WRITING, sqlPath)
		if err := writeSQLiteFile(sqlPath, tables); err != nil {
			return err
		}
	}
	if dbPath != "" {
		logf(WRITING, dbPath)
		return writeSQLite(dbPath, tables, logf)
	}
	return nil
}

func writeMySQLOutputs(ctx context.Context, sqlPath string, mysqlCfg MySQLConfig, tables []*Table, logf func(string, ...any)) error {
	if sqlPath != "" {
		logf("escribiendo %s...", sqlPath)
		if err := writeMySQLFile(sqlPath, tables); err != nil {
			return err
		}
	}
	logf("conectando y volcando a MySQL/MariaDB (%s:%d/%s)...", mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database)
	return writeMySQL(ctx, mysqlCfg, tables, logf)
}
