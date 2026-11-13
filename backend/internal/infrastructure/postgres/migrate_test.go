// ===== FILE: backend/internal/infrastructure/postgres/migrate_test.go =====
package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   []string
	}{
		{
			name:   "simple statements",
			script: "CREATE TABLE a (id INT);\nCREATE TABLE b (id INT);",
			want:   []string{"CREATE TABLE a (id INT)", "CREATE TABLE b (id INT)"},
		},
		{
			name:   "semicolon inside string literal",
			script: "INSERT INTO t VALUES ('a;b');",
			want:   []string{"INSERT INTO t VALUES ('a;b')"},
		},
		{
			name:   "escaped quote inside literal",
			script: "INSERT INTO t VALUES ('it''s; fine');",
			want:   []string{"INSERT INTO t VALUES ('it''s; fine')"},
		},
		{
			name:   "semicolon inside line comment",
			script: "-- drop everything; not really\nCREATE TABLE a (id INT);",
			want:   []string{"-- drop everything; not really\nCREATE TABLE a (id INT)"},
		},
		{
			name:   "semicolon inside block comment",
			script: "/* a; b */\nCREATE TABLE a (id INT);",
			want:   []string{"/* a; b */\nCREATE TABLE a (id INT)"},
		},
		{
			name:   "dollar quoted function body",
			script: "CREATE FUNCTION f() RETURNS void AS $$\nBEGIN\n  DELETE FROM t;\nEND;\n$$ LANGUAGE plpgsql;",
			want: []string{
				"CREATE FUNCTION f() RETURNS void AS $$\nBEGIN\n  DELETE FROM t;\nEND;\n$$ LANGUAGE plpgsql",
			},
		},
		{
			name:   "tagged dollar quote",
			script: "DO $body$ BEGIN PERFORM 1; END $body$;",
			want:   []string{"DO $body$ BEGIN PERFORM 1; END $body$"},
		},
		{
			name:   "trailing comment without semicolon is dropped",
			script: "SELECT 1;\n-- done",
			want:   []string{"SELECT 1"},
		},
		{
			name:   "trailing statement without semicolon is kept",
			script: "SELECT 1;\nSELECT 2",
			want:   []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:   "quoted identifier keeps its semicolon",
			script: `CREATE TABLE "odd;name" (id INT);`,
			want:   []string{`CREATE TABLE "odd;name" (id INT)`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSQL(tt.script)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d statements %q, want %d", len(got), got, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("statement %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestListUpMigrations(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"000002_phase_a.up.sql":     "SELECT 2;",
		"000002_phase_a.down.sql":   "SELECT 0;",
		"000010_later.up.sql":       "SELECT 10;",
		"000001_init_schema.up.sql": "SELECT 1;",
		"notes.txt":                 "ignored",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := listUpMigrations(dir)
	if err != nil {
		t.Fatalf("listUpMigrations() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d migrations, want 3: %+v", len(got), got)
	}
	for i, want := range []int{1, 2, 10} {
		if got[i].Version != want {
			t.Errorf("migration %d version = %d, want %d", i, got[i].Version, want)
		}
	}
}

func TestListUpMigrations_RejectsDuplicateVersion(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"000002_a.up.sql", "000002_b.up.sql"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := listUpMigrations(dir); err == nil {
		t.Error("listUpMigrations() should reject duplicate versions")
	}
}

// ===== END FILE =====
