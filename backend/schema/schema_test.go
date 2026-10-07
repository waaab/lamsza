// Package schema holds no Go code, only this guard over the committed dump.
package schema

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The backend creates tables and adds columns on boot (CREATE TABLE IF NOT
// EXISTS, ALTER TABLE ... ADD COLUMN IF NOT EXISTS). A dev database therefore
// gets them, but a database built from backend/schema/ by
// scripts/db-bootstrap.sh (CI, a new machine, every scratch test database)
// only gets what was dumped. admin_audit_log was added on boot without a
// re-dump, so a fresh database had no audit table for lamsza-admin to write to.
//
// This test fails when boot DDL names a table or column that 001_schema.sql
// does not have. Fix it by running scripts/db-dump-schema.sh against a dev
// database that has booted the new code, and committing the result.

var (
	bootCreateTable = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(?:public\.)?(\w+)`)
	bootAlterTable  = regexp.MustCompile("(?i)ALTER\\s+TABLE\\s+(?:IF\\s+EXISTS\\s+)?(?:ONLY\\s+)?(?:public\\.)?(\\w+)([^;`]*)")
	bootAddColumn   = regexp.MustCompile(`(?i)ADD\s+COLUMN\s+IF\s+NOT\s+EXISTS\s+(\w+)`)
	dumpTable       = regexp.MustCompile(`(?s)CREATE TABLE public\.(\w+) \((.*?)\n\);`)
	dumpColumn      = regexp.MustCompile(`(?m)^\s+(\w+)\s`)
)

// dumpedTables maps each table in 001_schema.sql to its column names.
func dumpedTables(t *testing.T) map[string]map[string]bool {
	t.Helper()
	raw, err := os.ReadFile("001_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	tables := map[string]map[string]bool{}
	for _, m := range dumpTable.FindAllStringSubmatch(string(raw), -1) {
		cols := map[string]bool{}
		for _, c := range dumpColumn.FindAllStringSubmatch(m[2], -1) {
			cols[c[1]] = true
		}
		tables[m[1]] = cols
	}
	if len(tables) == 0 {
		t.Fatal("found no CREATE TABLE in 001_schema.sql; has the dump format changed?")
	}
	return tables
}

// bootSources returns the non-test Go files of the backend, keyed by path.
// migrations/ is a hand-applied historical record, not boot code.
func bootSources(t *testing.T) map[string]string {
	t.Helper()
	sources := map[string]string{}
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "migrations" || d.Name() == "schema" || d.Name() == "data" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sources[path] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return sources
}

func TestBootDDLIsInTheSchemaDump(t *testing.T) {
	tables := dumpedTables(t)
	var missing []string
	seenTables, seenColumns := 0, 0
	for path, src := range bootSources(t) {
		for _, m := range bootCreateTable.FindAllStringSubmatch(src, -1) {
			seenTables++
			if _, ok := tables[m[1]]; !ok {
				missing = append(missing, "table "+m[1]+" ("+path+")")
			}
		}
		for _, m := range bootAlterTable.FindAllStringSubmatch(src, -1) {
			table := m[1]
			for _, c := range bootAddColumn.FindAllStringSubmatch(m[2], -1) {
				seenColumns++
				cols, ok := tables[table]
				if !ok || !cols[c[1]] {
					missing = append(missing, "column "+table+"."+c[1]+" ("+path+")")
				}
			}
		}
	}
	if seenTables == 0 {
		t.Fatal("found no boot CREATE TABLE at all; is the source walk broken?")
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("boot DDL creates things backend/schema/001_schema.sql does not have, so a database built by "+
			"scripts/db-bootstrap.sh lacks them. Run scripts/db-dump-schema.sh and commit the result.\n  %s",
			strings.Join(missing, "\n  "))
	}
	t.Logf("checked %d boot CREATE TABLE and %d ADD COLUMN statements against %d dumped tables", seenTables, seenColumns, len(tables))
}
