// ===== FILE: backend/internal/infrastructure/postgres/migrate.go =====
package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/nyouaku/SOPQ/backend/pkg/logger"
	"gorm.io/gorm"
)

// RunMigrations menerapkan file migrasi `NNNNNN_*.up.sql` yang belum tercatat
// di tabel schema_migrations. Tiap file dijalankan dalam satu transaksi, jadi
// migrasi yang gagal tidak meninggalkan skema setengah jadi.
//
// Baseline: database yang sudah diinisialisasi lewat entrypoint Postgres
// (docker-entrypoint-initdb.d) punya tabel `users` tapi tidak punya
// `schema_migrations`. Versi 1 dianggap sudah diterapkan supaya 000001 tidak
// dijalankan ulang.
func (d *Database) RunMigrations(dir string) error {
	ctx := context.Background()

	if err := d.ensureMigrationTable(ctx); err != nil {
		return err
	}

	if err := d.baselineInitSchema(ctx); err != nil {
		return err
	}

	files, err := listUpMigrations(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		logger.Warn("Tidak ada file migrasi di %s", dir)
		return nil
	}

	applied, err := d.appliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, m := range files {
		if applied[m.Version] {
			continue
		}

		body, err := os.ReadFile(m.Path)
		if err != nil {
			return fmt.Errorf("baca migrasi %s: %w", m.Path, err)
		}

		logger.Info("Menerapkan migrasi %03d (%s)", m.Version, m.Name)
		if err := d.applyMigration(ctx, m.Version, m.Name, string(body)); err != nil {
			return fmt.Errorf("migrasi %s gagal: %w", m.Path, err)
		}
	}

	return nil
}

type migrationFile struct {
	Version int
	Name    string
	Path    string
}

var migrationNamePattern = regexp.MustCompile(`^(\d+)_(.+)\.up\.sql$`)

// listUpMigrations mengembalikan file `*.up.sql` terurut menurut nomor versi.
func listUpMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("baca direktori migrasi %s: %w", dir, err)
	}

	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationNamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("nomor versi tidak valid pada %s: %w", entry.Name(), err)
		}
		files = append(files, migrationFile{
			Version: version,
			Name:    match[2],
			Path:    filepath.Join(dir, entry.Name()),
		})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Version < files[j].Version })

	for i := 1; i < len(files); i++ {
		if files[i].Version == files[i-1].Version {
			return nil, fmt.Errorf("dua file migrasi memakai versi %d", files[i].Version)
		}
	}

	return files, nil
}

func (d *Database) ensureMigrationTable(ctx context.Context) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INT PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`
	if err := d.DB.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("buat tabel schema_migrations: %w", err)
	}
	return nil
}

func (d *Database) baselineInitSchema(ctx context.Context) error {
	if !d.DB.Migrator().HasTable("users") {
		return nil
	}
	stmt := `
		INSERT INTO schema_migrations (version, name)
		VALUES (1, 'init_schema (baseline)')
		ON CONFLICT (version) DO NOTHING`
	if err := d.DB.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("catat baseline migrasi: %w", err)
	}
	return nil
}

func (d *Database) appliedVersions(ctx context.Context) (map[int]bool, error) {
	var versions []int
	if err := d.DB.WithContext(ctx).
		Model(&struct{ Version int }{}).
		Table("schema_migrations").
		Pluck("version", &versions).Error; err != nil {
		return nil, fmt.Errorf("baca schema_migrations: %w", err)
	}

	applied := make(map[int]bool, len(versions))
	for _, v := range versions {
		applied[v] = true
	}
	return applied, nil
}

func (d *Database) applyMigration(ctx context.Context, version int, name, body string) error {
	return d.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, stmt := range splitSQL(body) {
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("statement gagal (%s): %w", firstLine(stmt), err)
			}
		}
		return tx.Exec(
			`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`,
			version, name,
		).Error
	})
}

// splitSQL memecah skrip menjadi statement per `;`, dengan menghormati
// komentar, string literal, dan blok dollar-quoted ($$ ... $$, $tag$ ... $tag$).
// Wajib: driver Postgres tidak menerima banyak statement dalam satu Exec.
func splitSQL(script string) []string {
	var (
		statements []string
		current    strings.Builder
		dollarTag  string
		inDollar   bool
		inLine     bool
		inBlock    bool
		inSingle   bool
		inDouble   bool
	)

	runes := []rune(script)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		switch {
		case inLine:
			if ch == '\n' {
				inLine = false
			}
			current.WriteRune(ch)
			continue

		case inBlock:
			current.WriteRune(ch)
			if ch == '*' && next == '/' {
				current.WriteRune(next)
				i++
				inBlock = false
			}
			continue

		case inDollar:
			if ch == '$' {
				closing := "$" + dollarTag + "$"
				if hasPrefixAt(runes, i, closing) {
					for _, r := range closing {
						current.WriteRune(r)
					}
					i += len([]rune(closing)) - 1
					dollarTag = ""
					inDollar = false
					continue
				}
			}
			current.WriteRune(ch)
			continue

		case inSingle:
			current.WriteRune(ch)
			if ch == '\'' {
				if next == '\'' { // escaped quote ''
					current.WriteRune(next)
					i++
					continue
				}
				inSingle = false
			}
			continue

		case inDouble:
			current.WriteRune(ch)
			if ch == '"' {
				inDouble = false
			}
			continue
		}

		// Di luar semua konteks khusus.
		switch {
		case ch == '-' && next == '-':
			inLine = true
			current.WriteRune(ch)
		case ch == '/' && next == '*':
			inBlock = true
			current.WriteRune(ch)
		case ch == '\'':
			inSingle = true
			current.WriteRune(ch)
		case ch == '"':
			inDouble = true
			current.WriteRune(ch)
		case ch == '$':
			if tag, length, ok := readDollarTag(runes[i:]); ok {
				dollarTag = tag
				inDollar = true
				for _, r := range runes[i : i+length] {
					current.WriteRune(r)
				}
				i += length - 1
			} else {
				current.WriteRune(ch)
			}
		case ch == ';':
			if stmt := strings.TrimSpace(current.String()); stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if stmt := strings.TrimSpace(current.String()); hasSQLContent(stmt) {
		statements = append(statements, stmt)
	}

	return statements
}

// readDollarTag membaca pembuka `$$` atau `$tag$`; mengembalikan tag (tanpa $),
// panjang teks pembungkus, dan apakah bentuknya valid.
func readDollarTag(runes []rune) (string, int, bool) {
	if len(runes) < 2 || runes[0] != '$' {
		return "", 0, false
	}
	for i := 1; i < len(runes); i++ {
		ch := runes[i]
		if ch == '$' {
			return string(runes[1:i]), i + 1, true
		}
		// Tag hanya boleh huruf, angka, atau underscore.
		if !(ch == '_' ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9')) {
			return "", 0, false
		}
	}
	return "", 0, false
}

func hasPrefixAt(runes []rune, start int, prefix string) bool {
	p := []rune(prefix)
	if start+len(p) > len(runes) {
		return false
	}
	for i, r := range p {
		if runes[start+i] != r {
			return false
		}
	}
	return true
}

// hasSQLContent melaporkan apakah potongan terakhir berisi statement nyata
// (bukan hanya komentar).
func hasSQLContent(stmt string) bool {
	for _, line := range strings.Split(stmt, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		return true
	}
	return false
}

func firstLine(stmt string) string {
	line := strings.TrimSpace(stmt)
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = line[:idx]
	}
	if len(line) > 80 {
		line = line[:80] + "..."
	}
	return line
}

// ===== END FILE =====
