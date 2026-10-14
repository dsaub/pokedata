package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registra el driver "sqlite" para database/sql
)

// ---------------------------------------------------------------------
// Volcado a fichero SQLite
// ---------------------------------------------------------------------

func writeSQLite(path string, tables []*Table, logf func(string, ...any)) (err error) {
	tmp := path + ".tmp"
	os.Remove(tmp)
	defer func() {
		if err != nil {
			os.Remove(tmp)
		}
	}()
	db, err := sql.Open("sqlite", tmp)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, p := range []string{"PRAGMA journal_mode = OFF", "PRAGMA synchronous = OFF", "PRAGMA foreign_keys = OFF"} {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}
	if _, err := db.Exec(schemaSQLiteSQL); err != nil {
		return fmt.Errorf("esquema: %w", err)
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, t := range tables {
		start := time.Now()
		if err := insertTable(tx, t); err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: %w", t.Name, err)
		}
		if len(t.Rows) > 50000 {
			logf("  %-32s %8d filas (%s)", t.Name, len(t.Rows), time.Since(start).Round(time.Millisecond))
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if _, err := db.Exec(viewsSQLiteSQL); err != nil {
		return fmt.Errorf("vistas: %w", err)
	}
	if err := checkSQLiteViews(db); err != nil {
		return err
	}
	if err := checkSQLiteForeignKeys(db); err != nil {
		return err
	}
	for _, p := range []string{"ANALYZE", "PRAGMA journal_mode = DELETE"} {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}
	if err := db.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// checkSQLiteViews ejecuta cada vista: SQLite no valida las columnas al crearlas.
func checkSQLiteViews(db *sql.DB) error {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'view'")
	if err != nil {
		return err
	}
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		names = append(names, n)
	}
	rows.Close()
	for _, n := range names {
		var x any
		err := db.QueryRow("SELECT 1 FROM `" + n + "` LIMIT 1").Scan(&x)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("vista %s: %w", n, err)
		}
	}
	return nil
}

func checkSQLiteForeignKeys(db *sql.DB) error {
	rows, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	var msgs []string
	for rows.Next() {
		var table, parent string
		var rowid, fkid sql.NullInt64
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err != nil {
			return err
		}
		if len(msgs) < 20 {
			msgs = append(msgs, fmt.Sprintf("%s(rowid %d) -> %s", table, rowid.Int64, parent))
		}
	}
	if len(msgs) > 0 {
		return fmt.Errorf("claves foráneas rotas:\n  %s", strings.Join(msgs, "\n  "))
	}
	return rows.Err()
}

// ---------------------------------------------------------------------
// Volcado a script .sql para SQLite (para `sqlite3 pokedata.db < pokedata.sql`)
// ---------------------------------------------------------------------

func writeSQLiteFile(path string, tables []*Table) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	fmt.Fprintf(w, "-- pokedata: generado por pokedata-gen el %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "-- Uso: sqlite3 pokedata.db < %s\n\n", filepath.Base(path))
	fmt.Fprintln(w, "PRAGMA foreign_keys = OFF;")
	fmt.Fprintln(w, "BEGIN TRANSACTION;")
	fmt.Fprintln(w, schemaSQLiteSQL)
	for _, t := range tables {
		writeInserts(w, t, sqliteLiteral)
	}
	fmt.Fprintln(w, viewsSQLiteSQL)
	fmt.Fprintln(w, "COMMIT;")
	fmt.Fprintln(w, "PRAGMA foreign_keys = ON;")
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Close()
}

func sqliteLiteral(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		s := strconv.FormatFloat(x, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case string:
		return "'" + strings.ReplaceAll(x, "'", "''") + "'"
	default:
		panic(fmt.Sprintf("tipo no soportado %T", v))
	}
}
