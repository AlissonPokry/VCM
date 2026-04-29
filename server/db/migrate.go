package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	sqlitemigrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
)

type promptMigration struct {
	version    uint
	identifier string
	path       string
}

type promptSource struct {
	migrations []promptMigration
}

// RunMigrations applies ordered SQL migrations with golang-migrate.
func RunMigrations(conn *sql.DB, migrationsDir string) error {
	sourceDriver, err := newPromptSource(migrationsDir)
	if err != nil {
		return err
	}

	databaseDriver, err := sqlitemigrate.WithInstance(conn, &sqlitemigrate.Config{})
	if err != nil {
		return fmt.Errorf("create sqlite migration driver: %w", err)
	}

	runner, err := migrate.NewWithInstance("prompt-sql", sourceDriver, "sqlite3", databaseDriver)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}

	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

func newPromptSource(migrationsDir string) (*promptSource, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	matcher := regexp.MustCompile(`^([0-9]+)_.+\.sql$`)
	migrations := make([]promptMigration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		match := matcher.FindStringSubmatch(name)
		if len(match) != 2 {
			continue
		}

		parsed, err := strconv.ParseUint(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", name, err)
		}

		migrations = append(migrations, promptMigration{
			version:    uint(parsed),
			identifier: strings.TrimSuffix(name, filepath.Ext(name)),
			path:       filepath.Join(migrationsDir, name),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return &promptSource{migrations: migrations}, nil
}

// Open satisfies the golang-migrate source.Driver interface.
func (s *promptSource) Open(url string) (source.Driver, error) {
	return s, nil
}

// Close satisfies the golang-migrate source.Driver interface.
func (s *promptSource) Close() error {
	return nil
}

// First returns the first migration version.
func (s *promptSource) First() (uint, error) {
	if len(s.migrations) == 0 {
		return 0, source.ErrNotExist
	}
	return s.migrations[0].version, nil
}

// Prev returns the previous migration version.
func (s *promptSource) Prev(version uint) (uint, error) {
	for i := range s.migrations {
		if s.migrations[i].version == version && i > 0 {
			return s.migrations[i-1].version, nil
		}
	}
	return 0, source.ErrNotExist
}

// Next returns the next migration version.
func (s *promptSource) Next(version uint) (uint, error) {
	for i := range s.migrations {
		if s.migrations[i].version == version && i+1 < len(s.migrations) {
			return s.migrations[i+1].version, nil
		}
	}
	return 0, source.ErrNotExist
}

// ReadUp returns the SQL body for an up migration.
func (s *promptSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	migration, ok := s.find(version)
	if !ok {
		return nil, "", source.ErrNotExist
	}

	file, err := os.Open(migration.path)
	if err != nil {
		return nil, "", fmt.Errorf("open migration %s: %w", migration.identifier, err)
	}
	return file, migration.identifier, nil
}

// ReadDown returns an empty down migration because this project only ships forward migrations.
func (s *promptSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	migration, ok := s.find(version)
	if !ok {
		return nil, "", source.ErrNotExist
	}
	return io.NopCloser(strings.NewReader("")), migration.identifier, nil
}

func (s *promptSource) find(version uint) (promptMigration, bool) {
	for _, migration := range s.migrations {
		if migration.version == version {
			return migration, true
		}
	}
	return promptMigration{}, false
}
