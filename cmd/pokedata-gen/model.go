package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// Table es una tabla en memoria lista para volcarse a SQLite o a un .sql.
// Los valores son nil, int64, float64 o string.
type Table struct {
	Name string
	Cols []string
	Rows [][]any
}

func (t *Table) Add(vals ...any) {
	if len(vals) != len(t.Cols) {
		panic(fmt.Sprintf("%s: %d valores para %d columnas", t.Name, len(vals), len(t.Cols)))
	}
	for i, v := range vals {
		if n, ok := v.(int); ok {
			vals[i] = int64(n)
		}
	}
	t.Rows = append(t.Rows, vals)
}

// Dataset agrupa todas las tablas generadas.
type Dataset struct {
	byName map[string]*Table
}

func NewDataset() *Dataset { return &Dataset{byName: map[string]*Table{}} }

// T declara (o devuelve) una tabla con sus columnas.
func (d *Dataset) T(name string, cols ...string) *Table {
	if t, ok := d.byName[name]; ok {
		return t
	}
	t := &Table{Name: name, Cols: cols}
	d.byName[name] = t
	return t
}

// ---------------------------------------------------------------------
// Lectura de CSV de PokeAPI
// ---------------------------------------------------------------------

type CSV struct {
	name string
	idx  map[string]int
	rows [][]string
}

type Rec struct {
	c *CSV
	r []string
}

func readCSV(path string) *CSV {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	rd := csv.NewReader(f)
	rd.FieldsPerRecord = -1
	rd.TrimLeadingSpace = true // hay filas con `, "texto"`
	rd.LazyQuotes = true
	all, err := rd.ReadAll()
	if err != nil {
		panic(fmt.Errorf("%s: %w", path, err))
	}
	c := &CSV{name: path, idx: map[string]int{}}
	for i, h := range all[0] {
		c.idx[h] = i
	}
	c.rows = all[1:]
	return c
}

func (c *CSV) Each(fn func(Rec)) {
	for _, r := range c.rows {
		fn(Rec{c, r})
	}
}

func (r Rec) Str(col string) string {
	i, ok := r.c.idx[col]
	if !ok {
		panic(fmt.Sprintf("%s: no existe la columna %q", r.c.name, col))
	}
	if i >= len(r.r) {
		return ""
	}
	return r.r[i]
}

func (r Rec) Int(col string) int {
	s := r.Str(col)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("%s: %s=%q no es entero", r.c.name, col, s))
	}
	return n
}

// S devuelve el texto o nil si está vacío.
func (r Rec) S(col string) any {
	if s := r.Str(col); s != "" {
		return s
	}
	return nil
}

// I devuelve el entero o nil si está vacío.
func (r Rec) I(col string) any {
	if r.Str(col) == "" {
		return nil
	}
	return int64(r.Int(col))
}

// B normaliza booleanos ("True"/"1") a 0/1.
func (r Rec) B(col string) any {
	switch r.Str(col) {
	case "":
		return nil
	case "1", "True", "true":
		return int64(1)
	default:
		return int64(0)
	}
}
