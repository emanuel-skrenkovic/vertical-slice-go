package sqlmigration

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	rootPath string

	conf config.Config
	db   *sqlx.DB
)

const (
	testDir = "./TEST_TEMP"

	mainTableUpMigration = `
		CREATE TABLE main_table1 (
			id INTEGER PRIMARY KEY
		);`

	mainTableDownMigration = `
		DROP TABLE main_table1;`

	depdendantTableUpMigration = `
		CREATE TABLE dependant_table (
			id serial PRIMARY KEY,
			main_id INTEGER,

			CONSTRAINT fk_main FOREIGN KEY(main_id) REFERENCES main_table1(id)
		);`

	depdendantTableDownMigration = `
		ALTER TABLE dependant_table DROP CONSTRAINT fk_main;
		DROP TABLE dependant_table;`

	mainTable2UpMigration = `
		CREATE TABLE main_table2 (
			id serial PRIMARY KEY
		);`

	mainTable2DownMigration = `
		DROP TABLE main_table2;`

	mainTable2mainTable2Migration = `
		CREATE TABLE main_table1_main_table2 (
			main_id_1 INTEGER REFERENCES main_table1(id),
			main_id_2 INTEGER REFERENCES main_table2(id),

			CONSTRAINT pk_main_table1_main_table2 PRIMARY KEY (main_id_1, main_id_2)
		);`
)

func TestMain(m *testing.M) {
	rootPath = os.Args[len(os.Args)-1]
	if rootPath == "" {
		log.Fatal("root directoy path is empty")
	}

	// Otherwise, doesn't work from Goland.
	// Need to pass in through other means than args
	rootPath = "/home/emanuel/dev/vertical-slice-go"

	var err error
	conf, err = config.Load(path.Join(rootPath, "config.env"))
	if err != nil {
		log.Fatal(err)
	}

	db, err = sqlx.Connect("postgres", conf.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	if err := os.RemoveAll(testDir); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func Test_Applies_All_Migrations_In_Directory(t *testing.T) {
	// Arrange
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())
	defer cleanUpTestMigrations()
	defer func() {
		if err := os.RemoveAll(testMigrationsPath); err != nil {
			t.Errorf("cleanup failed to remove directory: %s", testMigrationsPath)
		}
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	// Act
	err := Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	m := migrations(t, db)

	expectedMigrationsFound := 2
	if len(m) != expectedMigrationsFound {
		t.Errorf("expected '%d' migrations found '%d'", expectedMigrationsFound, len(m))
	}
}

func Test_Applied_Migrations_From_Directory_After_Already_Applied_Version(t *testing.T) {
	// Arrange
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())
	defer cleanUpTestMigrations()
	defer func() {
		if err := os.RemoveAll(testMigrationsPath); err != nil {
			t.Errorf("cleanup failed to remove directory: %s", testMigrationsPath)
		}
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	err := Run(testMigrationsPath, conf.DatabaseURL)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Act
	createMigrationFile(t, "main_table2", mainTable2UpMigration, mainTable2DownMigration)
	err = Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	m := migrations(t, db)

	expectedMigrationsFound := 3
	if len(m) != expectedMigrationsFound {
		t.Errorf("expected '%d' migrations found '%d'", expectedMigrationsFound, len(m))
	}
}

func Test_Reverts_All_Attempted_Migrations_On_Failed_Migration_Attempt(t *testing.T) {
	// Arrange
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())
	defer cleanUpTestMigrations()
	defer func() {
		if err := os.RemoveAll(testMigrationsPath); err != nil {
			t.Errorf("cleanup failed to remove directory: %s", testMigrationsPath)
		}
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	// Act
	createMigrationFile(t, "main_table2", "invalid", "SELECT 1;")
	err := Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	if err == nil {
		t.Error("did not receive expected error")
	}

	m := migrations(t, db)

	expectedMigrationsFound := 0
	if len(m) != expectedMigrationsFound {
		t.Errorf("expected '%d' migrations found '%d'", expectedMigrationsFound, len(m))
	}
}

func Test_Reverts_All_Attempted_Migrations_On_Failed_Migration_Attempt_Leaving_Previous_Migrations(t *testing.T) {
	// Arrange
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())
	defer cleanUpTestMigrations()
	defer func() {
		if err := os.RemoveAll(testMigrationsPath); err != nil {
			t.Errorf("cleanup failed to remove directory: %s", testMigrationsPath)
		}
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	if err := Run(testMigrationsPath, conf.DatabaseURL); err != nil {
		t.Errorf("received unexpected error: %v", err)
	}


	// Act
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)
	createMigrationFile(t, "main_table2", "invalid", "SELECT 1;")
	err := Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	if err == nil {
		t.Error("did not receive expected error")
	}

	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)
	m := migrations(t, db)

	expectedMigrationsFound := 1
	if len(m) != expectedMigrationsFound {
		t.Errorf("expected '%d' migrations found '%d'", expectedMigrationsFound, len(m))
	}
}


func cleanUpTestMigrations() {
	db.MustExec("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public")
}

func createMigrationFile(t *testing.T, name, upScript, downScript string) {
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())
	if _, err := os.Stat(testMigrationsPath); err != nil && errors.Is(err, os.ErrNotExist) {
		os.Mkdir(testMigrationsPath, 0755)
	}

	version := getMigrationVersion(t)

	upScriptName := strings.Join([]string{fmt.Sprintf("%d", version), name, "up", "sql"}, ".")
	downScriptName := strings.Join([]string{fmt.Sprintf("%d", version), name, "down", "sql"}, ".")

	if err := os.WriteFile(path.Join(testMigrationsPath, upScriptName), []byte(upScript), 0777); err != nil {
		t.Errorf("unexpected error occurred: %v", err)
	}

	if err := os.WriteFile(path.Join(testMigrationsPath, downScriptName), []byte(downScript), 0777); err != nil {
		t.Errorf("unexpected error occurred: %v", err)
	}
}

func getMigrationVersion(t *testing.T) int {
	testMigrationsPath := path.Join(rootPath, "pkg", "sql-migrations", "test", "migrations", t.Name())

	entities, err := os.ReadDir(testMigrationsPath)
	if err != nil {
		t.Errorf("received unexpected err: %v", err)
	}

	highestVersion := 0
	for _, entry := range entities {
		if entry.IsDir() {
			continue
		}

		parts := strings.Split(entry.Name(), ".")
		if len(parts) != 4 {
			t.Errorf("found unexpected file in test migrations folder %s %s", testMigrationsPath, entry.Name())
		}

		versionPart := parts[0]

		version, err := strconv.Atoi(versionPart)
		if err != nil {
			t.Errorf("received unexpected error: %v", err)
		}

		if version > highestVersion {
			highestVersion = version
		}
	}

	return highestVersion + 1
}

func migrations(t *testing.T, db *sqlx.DB) []migration {
	var m []migration
	if err := db.Select(&m, "SELECT * FROM schema_migration;"); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	return m
}
