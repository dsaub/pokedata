package main

import (
	"bufio"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed sql/schema.sql
var schemaSQL string

//go:embed sql/views.sql
var viewsSQL string

var createTableRe = regexp.MustCompile(`(?m)^CREATE TABLE (\w+)`)

// identRe limita los nombres de tabla y columna que se interpolan en el SQL
// generado; los valores siempre van como parámetros o pasan por sqlLiteral.
var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// orderedTables devuelve las tablas en el orden del esquema (respeta las FK)
// y comprueba que esquema y datos coinciden.
func orderedTables(ds *Dataset) ([]*Table, error) {
	var out []*Table
	seen := map[string]bool{}
	for _, m := range createTableRe.FindAllStringSubmatch(schemaSQL, -1) {
		t, ok := ds.byName[m[1]]
		if !ok {
			return nil, fmt.Errorf("la tabla %s del esquema no tiene datos", m[1])
		}
		if err := checkIdents(t); err != nil {
			return nil, err
		}
		seen[m[1]] = true
		out = append(out, t)
	}
	for name := range ds.byName {
		if !seen[name] {
			return nil, fmt.Errorf("la tabla %s no está en el esquema", name)
		}
	}
	return out, nil
}

func checkIdents(t *Table) error {
	if !identRe.MatchString(t.Name) {
		return fmt.Errorf("nombre de tabla no válido: %q", t.Name)
	}
	for _, c := range t.Cols {
		if !identRe.MatchString(c) {
			return fmt.Errorf("%s: nombre de columna no válido: %q", t.Name, c)
		}
	}
	return nil
}

func batchSize(t *Table) int {
	b := 30000 / len(t.Cols) // límite de variables de SQLite: 32766
	if b > 500 {
		b = 500
	}
	return b
}

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
	if _, err := db.Exec(schemaSQL); err != nil {
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
	if _, err := db.Exec(viewsSQL); err != nil {
		return fmt.Errorf("vistas: %w", err)
	}
	if err := checkViews(db); err != nil {
		return err
	}
	if err := checkForeignKeys(db); err != nil {
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

func insertTable(tx *sql.Tx, t *Table) error {
	bs := batchSize(t)
	one := "(" + strings.TrimSuffix(strings.Repeat("?,", len(t.Cols)), ",") + ")"
	prefix := "INSERT INTO " + t.Name + " (" + strings.Join(t.Cols, ", ") + ") VALUES "
	var full *sql.Stmt
	for i := 0; i < len(t.Rows); i += bs {
		end := min(i+bs, len(t.Rows))
		n := end - i
		var stmt *sql.Stmt
		var err error
		if n == bs && full != nil {
			stmt = full
		} else {
			// Solo se interpolan identificadores validados en orderedTables; los valores van como "?".
			stmt, err = tx.Prepare(prefix + strings.TrimSuffix(strings.Repeat(one+",", n), ",")) // NOSONAR go:S2077
			if err != nil {
				return err
			}
			if n == bs {
				full = stmt
			} else {
				defer stmt.Close()
			}
		}
		args := make([]any, 0, n*len(t.Cols))
		for _, r := range t.Rows[i:end] {
			args = append(args, r...)
		}
		if _, err := stmt.Exec(args...); err != nil {
			return err
		}
	}
	if full != nil {
		full.Close()
	}
	return nil
}

// checkViews ejecuta cada vista: SQLite no valida las columnas al crearlas.
func checkViews(db *sql.DB) error {
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
		err := db.QueryRow("SELECT 1 FROM " + n + " LIMIT 1").Scan(&x)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("vista %s: %w", n, err)
		}
	}
	return nil
}

func checkForeignKeys(db *sql.DB) error {
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
// Volcado a script .sql (para `sqlite3 pokedata.db < pokedata.sql`)
// ---------------------------------------------------------------------

func writeSQLFile(path string, tables []*Table) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	fmt.Fprintf(w, "-- pokedata: generado por pokedata-gen el %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "-- Uso: sqlite3 pokedata.db < pokedata.sql\n\n")
	fmt.Fprintln(w, "PRAGMA foreign_keys = OFF;")
	fmt.Fprintln(w, "BEGIN TRANSACTION;")
	fmt.Fprintln(w, schemaSQL)
	for _, t := range tables {
		bs := batchSize(t)
		prefix := "INSERT INTO " + t.Name + " (" + strings.Join(t.Cols, ", ") + ") VALUES\n"
		for i := 0; i < len(t.Rows); i += bs {
			end := min(i+bs, len(t.Rows))
			w.WriteString(prefix)
			for j, r := range t.Rows[i:end] {
				w.WriteByte('(')
				for k, v := range r {
					if k > 0 {
						w.WriteByte(',')
					}
					w.WriteString(sqlLiteral(v))
				}
				if i+j == end-1 {
					w.WriteString(");\n")
				} else {
					w.WriteString("),\n")
				}
			}
		}
	}
	fmt.Fprintln(w, viewsSQL)
	fmt.Fprintln(w, "COMMIT;")
	fmt.Fprintln(w, "PRAGMA foreign_keys = ON;")
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Close()
}

func sqlLiteral(v any) string {
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
