package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // registra el driver "mysql" para database/sql
)

type MySQLConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Database       string
	CreateDatabase bool
	TLS            string
}

func (c *MySQLConfig) normalize() {
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if strings.Contains(c.Host, ":") {
		if h, p, err := net.SplitHostPort(c.Host); err == nil {
			c.Host = h
			if portNum, err := strconv.Atoi(p); err == nil && portNum > 0 {
				c.Port = portNum
			}
		}
	}
	if c.Port <= 0 {
		c.Port = 3306
	}
	if c.User == "" {
		c.User = "root"
	}
	if c.Database == "" {
		c.Database = "pokedata"
	}
	if c.TLS == "" {
		c.TLS = "preferred"
	}
}

// ---------------------------------------------------------------------
// Volcado a MySQL / MariaDB
// ---------------------------------------------------------------------

func writeMySQL(ctx context.Context, cfg MySQLConfig, tables []*Table, logf func(string, ...any)) error {
	cfg.normalize()

	if cfg.CreateDatabase {
		rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?multiStatements=true&charset=utf8mb4&parseTime=true&timeout=15s&tls=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.TLS)
		rootDB, err := sql.Open("mysql", rootDSN)
		if err != nil {
			return fmt.Errorf("error al conectar con servidor MySQL/MariaDB: %w", err)
		}
		if err := rootDB.PingContext(ctx); err != nil {
			rootDB.Close()
			return fmt.Errorf("error al conectar con servidor MySQL/MariaDB (%s:%d): %w", cfg.Host, cfg.Port, err)
		}
		logf("asegurando base de datos %s...", cfg.Database)
		createStmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.Database)
		if _, err := rootDB.ExecContext(ctx, createStmt); err != nil {
			rootDB.Close()
			return fmt.Errorf("error al crear base de datos %s: %w", cfg.Database, err)
		}
		rootDB.Close()
	}

	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&charset=utf8mb4&parseTime=true&timeout=15s&interpolateParams=true&tls=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.TLS)
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		return fmt.Errorf("error al conectar a la base de datos %s: %w", cfg.Database, err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("error al conectar a la base de datos %s en %s:%d: %w", cfg.Database, cfg.Host, cfg.Port, err)
	}

	db.SetMaxOpenConns(1)

	// Optimización de importación por lotes
	for _, p := range []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"SET UNIQUE_CHECKS = 0",
		"SET AUTOCOMMIT = 0",
	} {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return err
		}
	}

	// Limpiar vistas y tablas existentes de forma limpia e idempotente
	_, _ = db.ExecContext(ctx, "DROP VIEW IF EXISTS v_gym_teams, v_gimnasios, v_gym_battles, v_pokemon_learnset, v_type_chart, v_pokemon_abilities, v_pokemon")
	for i := len(tables) - 1; i >= 0; i-- {
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS `"+tables[i].Name+"`"); err != nil {
			return fmt.Errorf("error al limpiar tabla %s: %w", tables[i].Name, err)
		}
	}

	logf("creando esquema en MySQL/MariaDB...")
	if _, err := db.ExecContext(ctx, schemaMySQLSQL); err != nil {
		return fmt.Errorf("esquema MySQL: %w", err)
	}

	logf("insertando datos...")
	for _, t := range tables {
		start := time.Now()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := insertTable(tx, t); err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: %w", t.Name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", t.Name, err)
		}
		if len(t.Rows) > 50000 {
			logf("  %-32s %8d filas (%s)", t.Name, len(t.Rows), time.Since(start).Round(time.Millisecond))
		}
	}

	logf("creando vistas e índices...")
	if _, err := db.ExecContext(ctx, viewsMySQLSQL); err != nil {
		return fmt.Errorf("vistas MySQL: %w", err)
	}

	if err := checkMySQLViews(ctx, db); err != nil {
		return err
	}

	for _, p := range []string{
		"SET FOREIGN_KEY_CHECKS = 1",
		"SET UNIQUE_CHECKS = 1",
		"SET AUTOCOMMIT = 1",
	} {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

func checkMySQLViews(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT table_name FROM information_schema.views WHERE table_schema = DATABASE()")
	if err != nil {
		return err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return err
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, n := range names {
		var x any
		err := db.QueryRowContext(ctx, "SELECT 1 FROM `"+n+"` LIMIT 1").Scan(&x)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("vista %s: %w", n, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------
// Volcado a script .sql para MySQL / MariaDB
// ---------------------------------------------------------------------

func writeMySQLFile(path string, tables []*Table) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	fmt.Fprintf(w, "-- pokedata: generado por pokedata-gen el %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "-- Uso: mysql -u <usuario> -p <base_de_datos> < %s\n\n", filepath.Base(path))
	fmt.Fprintln(w, "SET NAMES utf8mb4;")
	fmt.Fprintln(w, "SET FOREIGN_KEY_CHECKS = 0;")
	fmt.Fprintln(w, "SET UNIQUE_CHECKS = 0;")
	fmt.Fprintln(w, "SET AUTOCOMMIT = 0;")
	fmt.Fprintln(w, "START TRANSACTION;")
	fmt.Fprintln(w, schemaMySQLSQL)
	for _, t := range tables {
		writeInserts(w, t, mysqlLiteral)
	}
	fmt.Fprintln(w, viewsMySQLSQL)
	fmt.Fprintln(w, "COMMIT;")
	fmt.Fprintln(w, "SET FOREIGN_KEY_CHECKS = 1;")
	fmt.Fprintln(w, "SET UNIQUE_CHECKS = 1;")
	fmt.Fprintln(w, "SET AUTOCOMMIT = 1;")
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Close()
}

func mysqlLiteral(v any) string {
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
		var b strings.Builder
		b.Grow(len(x) + 4)
		b.WriteByte('\'')
		for i := 0; i < len(x); i++ {
			c := x[i]
			switch c {
			case 0:
				b.WriteString(`\0`)
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\\':
				b.WriteString(`\\`)
			case '\'':
				b.WriteString(`\'`)
			case '"':
				b.WriteString(`\"`)
			case '\032': // Ctrl-Z
				b.WriteString(`\Z`)
			default:
				b.WriteByte(c)
			}
		}
		b.WriteByte('\'')
		return b.String()
	default:
		panic(fmt.Sprintf("tipo no soportado %T", v))
	}
}
