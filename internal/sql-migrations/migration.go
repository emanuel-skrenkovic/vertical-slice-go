package sqlmigration

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type migration struct {
	ID         int    `db:"id"`
	Version    int    `db:"version"`
	Name       string `db:"name"`
	UpScript   string
	DownScript string
}

// TODO: only for for postgres right now.
// Isolate away DB specific parts of code.
func Run(migrationsPath string, connectionString string) error {
	if _, err := os.Stat(migrationsPath); err != nil {
		return err
	}

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	migrations := make(map[int]migration, 0)

	for _, entry := range entries {
		// Sanity checks - only root directory, needs to have a name by convention
		// Name convention - migrationnumber.name.up.sql
		//                   migrationnumber.name.down.sql
		// Needs to have both up and down!

		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		parts := strings.Split(entry.Name(), ".")
		if len(parts) != 4 {
			// Doesn't match the naming convention.
			continue
		}

		migrationNumber, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		var m migration
		m = migrations[migrationNumber]

		m.Version = migrationNumber

		migrationName := parts[1]
		m.Name = migrationName

		// TODO: relative paths
		migrationContent, err := os.ReadFile(path.Join(migrationsPath, entry.Name()))
		if err != nil {
			return err
		}

		migrationScriptType := parts[2]
		switch migrationScriptType {
		case "up":
			m.UpScript = string(migrationContent)
		case "down":
			m.DownScript = string(migrationContent)
		default:
			return fmt.Errorf("uncrecognized script type: %s", migrationScriptType)
		}

		migrations[migrationNumber] = m
	}

	if err := validateFoundMigrationFiles(migrations); err != nil {
		return err
	}

	if err := ensureMigrationsSchema(db); err != nil {
		return err
	}

	// TODO: find diff between DB and file-defined migrations
	var alreadyAppliedMigrations []migration
	if err = db.Select(&alreadyAppliedMigrations, getAppliedMigrationsQuery()); err != nil {
		return err
	}

	lastAppliedMigrationVersion := 0
	if len(alreadyAppliedMigrations) > 0 {
		lastAppliedMigrationVersion = alreadyAppliedMigrations[0].Version
	}

	var migrationsToApply []migration
	for migrationVersion, migration := range migrations {
		if migrationVersion <= lastAppliedMigrationVersion {
			continue
		}

		migrationsToApply = append(migrationsToApply, migration)
	}

	if len(migrationsToApply) == 0 {
		return nil
	}

	sort.Slice(migrationsToApply, func(i, j int) bool {
		return migrationsToApply[i].Version < migrationsToApply[j].Version
	})

	newlyAppliedMigrations := []migration{}

	var migrationErr error
	for _, migration := range migrationsToApply {
		tx, err := db.BeginTxx(context.Background(), nil)
		if err != nil {
			return err
		}

		if _, err = tx.Exec(migration.UpScript); err != nil {
			migrationErr = err
			func() {
				if err := tx.Rollback(); err != nil {
					migrationErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}

		_, err = tx.Exec(
			`INSERT INTO
		     schema_migration (version, name)
			 VALUES ($1, $2);`,
			migration.Version,
			migration.Name,
		)
		if err != nil {
			migrationErr = err
			func() {
				if err := tx.Rollback(); err != nil {
					// migrationErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}

		if err := tx.Commit(); err != nil {
			migrationErr = err
			func() {
				if err := tx.Rollback(); err != nil {
					// migrationErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}

		newlyAppliedMigrations = append(newlyAppliedMigrations, migration)
	}

	if migrationErr != nil {
		if err := revertState(db, newlyAppliedMigrations); err != nil {
			return fmt.Errorf("%s: %w", err.Error(), migrationErr)
		}
		return migrationErr
	}

	return nil
}

func validateFoundMigrationFiles(migrations map[int]migration) error {
	var missingScriptsErr error
	for _, migration := range migrations {
		if migration.DownScript == "" {
			return fmt.Errorf("failed to find 'down' script for %s", migration.Name)
		}

		if migration.UpScript == "" {
			return fmt.Errorf("failed to find 'up' script for %s", migration.Name)
		}
	}
	return missingScriptsErr
}

func revertState(db *sqlx.DB, appliedMigrations []migration) error {
	var rollbackErr error
	for i := len(appliedMigrations)-1; i >= 0; i-- {
		migration := appliedMigrations[i]

		tx, err := db.BeginTxx(context.Background(), nil)
		if err != nil {
			return err
		}

		if _, err = tx.Exec(migration.DownScript); err != nil {
			func() {
				if err := tx.Rollback(); err != nil {
					rollbackErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}

		if _, err = tx.Exec("DELETE FROM schema_migration WHERE version = $1", migration.Version); err != nil {
			func() {
				if err := tx.Rollback(); err != nil {
					rollbackErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}

		if err := tx.Commit(); err != nil {
			func() {
				if err := tx.Rollback(); err != nil {
					rollbackErr = fmt.Errorf("failed to roll back transaction: %w", err)
				}
			}()
			break
		}
	}

	return rollbackErr
}

func ensureMigrationsSchema(db *sqlx.DB) error {
	const checkIfSchemaExistsQuery = `
		SELECT count(table_name)
		FROM information_schema.tables
		WHERE table_name = $1;`

	var schemas int
	if err := db.Get(&schemas, checkIfSchemaExistsQuery, "schema_migration"); err != nil {
		return err
	}

	if schemas > 0 {
		return nil
	}

	if _, err := db.Exec(
		`CREATE TABLE schema_migration (
			id serial PRIMARY KEY,
			name text NOT NULL,
			version integer NOT NULL
		)`,
	); err != nil {
		return err
	}

	return nil
}

func getAppliedMigrationsQuery() string {
	const checkIfSchemaExistsQuery = `
		SELECT *
		FROM schema_migration
		ORDER BY version DESC;`

	return checkIfSchemaExistsQuery
}
