package main

import (
	"bufio"
	"database/sql"
	_ "embed" // necesario para las directivas //go:embed de los .sql
	"fmt"
	"regexp"
	"strings"
)

//go:embed sql/schema.sql
var schemaSQLiteSQL string

//go:embed sql/views.sql
var viewsSQLiteSQL string

//go:embed sql/schema_mysql.sql
var schemaMySQLSQL string

//go:embed sql/views_mysql.sql
var viewsMySQLSQL string

var createTableRe = regexp.MustCompile(`(?m)^CREATE TABLE (\w+)`)

// identRe limita los nombres de tabla y columna que se interpolan en el SQL
// generado; los valores siempre van como parámetros o pasan por la función literal adecuada.
var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// orderedTables devuelve las tablas en el orden del esquema (respeta las FK)
// y comprueba que esquema y datos coinciden.
func orderedTables(ds *Dataset) ([]*Table, error) {
	var out []*Table
	seen := map[string]bool{}
	for _, m := range createTableRe.FindAllStringSubmatch(schemaSQLiteSQL, -1) {
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
	b := 30000 / len(t.Cols) // límite de variables de SQLite (32766) y holgado para MySQL
	if b > 500 {
		b = 500
	}
	return b
}

func insertTable(tx *sql.Tx, t *Table) error {
	bs := batchSize(t)
	one := "(" + strings.TrimSuffix(strings.Repeat("?,", len(t.Cols)), ",") + ")"
	quotedCols := make([]string, len(t.Cols))
	for i, c := range t.Cols {
		quotedCols[i] = "`" + c + "`"
	}
	prefix := "INSERT INTO `" + t.Name + "` (" + strings.Join(quotedCols, ", ") + ") VALUES "
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
			stmt, err = tx.Prepare(prefix + strings.TrimSuffix(strings.Repeat(one+",", n), ","))
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

// writeInserts escribe las filas de la tabla en INSERT de batchSize filas.
func writeInserts(w *bufio.Writer, t *Table, litFn func(any) string) {
	bs := batchSize(t)
	quotedCols := make([]string, len(t.Cols))
	for i, c := range t.Cols {
		quotedCols[i] = "`" + c + "`"
	}
	prefix := "INSERT INTO `" + t.Name + "` (" + strings.Join(quotedCols, ", ") + ") VALUES\n"
	for i := 0; i < len(t.Rows); i += bs {
		end := min(i+bs, len(t.Rows))
		w.WriteString(prefix)
		for j, r := range t.Rows[i:end] {
			writeValues(w, r, litFn)
			if i+j == end-1 {
				w.WriteString(";\n")
			} else {
				w.WriteString(",\n")
			}
		}
	}
}

// writeValues escribe una fila como "(v1,v2,...)".
func writeValues(w *bufio.Writer, r []any, litFn func(any) string) {
	w.WriteByte('(')
	for k, v := range r {
		if k > 0 {
			w.WriteByte(',')
		}
		w.WriteString(litFn(v))
	}
	w.WriteByte(')')
}
