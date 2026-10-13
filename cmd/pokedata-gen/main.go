// pokedata-gen descarga las fuentes (PokeAPI, pret, Serebii) y genera una base
// de datos SQLite con todos los datos de Pokémon y los equipos de los líderes de
// gimnasio de todas las generaciones. Opcionalmente vuelca también un script .sql.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	var (
		sources = flag.String("sources", "sources", "directorio de caché de las fuentes descargadas")
		dbPath  = flag.String("db", "pokedata.db", "fichero SQLite de salida (vacío = no generarlo)")
		sqlPath = flag.String("sql", "", "fichero .sql de salida (vacío = no generarlo)")
		refresh = flag.Bool("refresh", false, "volver a descargar todas las fuentes aunque estén en caché")
		offline = flag.Bool("offline", false, "no descargar nada; fallar si falta alguna fuente")
		quiet   = flag.Bool("q", false, "no mostrar progreso")
	)
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
	if err := run(*sources, *dbPath, *sqlPath, *refresh, *offline, logf); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(sources, dbPath, sqlPath string, refresh, offline bool, logf func(string, ...any)) (err error) {
	if dbPath == "" && sqlPath == "" {
		return fmt.Errorf("indica al menos -db o -sql")
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

	if err := ensureSources(ctx, sources, refresh, offline, logf); err != nil {
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

	if sqlPath != "" {
		logf("escribiendo %s...", sqlPath)
		if err := writeSQLFile(sqlPath, tables); err != nil {
			return err
		}
	}
	if dbPath != "" {
		logf("escribiendo %s...", dbPath)
		if err := writeSQLite(dbPath, tables, logf); err != nil {
			return err
		}
	}
	logf("listo en %s", time.Since(start).Round(time.Millisecond))
	return nil
}
