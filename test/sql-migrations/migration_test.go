package main

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
	sqlmigration "github.com/eskrenkovic/vertical-slice-go/internal/sql-migrations"
	"github.com/eskrenkovic/vertical-slice-go/internal/test"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

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
	args := os.Args

	if len(args) < 2 {
		log.Fatal("root path is required")
	}
	rootPath = args[len(args)-1]
	os.Setenv(config.RootPathEnv, rootPath)

	localConfigPath := path.Join(rootPath, "config.local.env")
	if _, err := os.Stat(localConfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(localConfigPath)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			if _, err := f.Write([]byte("SKIP_INFRASTRUCTURE=false")); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := godotenv.Load(path.Join(rootPath, "config.local.env")); err != nil {
		log.Fatal(err)
	}

	if err := godotenv.Load(path.Join(rootPath, "config.env")); err != nil {
		log.Fatal(err)
	}

	var err error
	conf, err = config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fixture, err := test.NewLocalTestFixture(path.Join(rootPath, "docker-compose.yml"), conf.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := fixture.Start(); err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := fixture.Stop(); err != nil {
			log.Fatal(err)
		}
	}()

	db, err = sqlx.Connect("postgres", conf.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	_ = m.Run()

	if err := os.RemoveAll(testDir); err != nil {
		log.Fatal(err)
	}
}

func Test_Applies_All_Migrations_In_Directory(t *testing.T) {
	// Arrange
	testMigrationsPath := migrationPath(t)
	defer cleanUpTestMigrations()
	defer func() {
		require.NoError(t, os.RemoveAll(testMigrationsPath))
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	// Act
	err := sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	require.NoError(t, err)

	m := getMigrations(t, db)

	expectedMigrationsFound := 2
	require.Equal(t, expectedMigrationsFound, len(m))
}

func Test_Applied_Migrations_From_Directory_After_Already_Applied_Version(t *testing.T) {
	// Arrange
	testMigrationsPath := migrationPath(t)
	defer cleanUpTestMigrations()
	defer func() {
		require.NoError(t, os.RemoveAll(testMigrationsPath))
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	err := sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)
	require.NoError(t, err)

	// Act
	createMigrationFile(t, "main_table2", mainTable2UpMigration, mainTable2DownMigration)
	err = sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	require.NoError(t, err)

	m := getMigrations(t, db)

	expectedMigrationsFound := 3
	require.Equal(t, expectedMigrationsFound, len(m))
}

func Test_Reverts_All_Attempted_Migrations_On_Failed_Migration_Attempt(t *testing.T) {
	// Arrange
	testMigrationsPath := migrationPath(t)
	defer cleanUpTestMigrations()
	defer func() {
		require.NoError(t, os.RemoveAll(testMigrationsPath))
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)

	// Act
	createMigrationFile(t, "main_table2", "invalid", "SELECT 1;")
	err := sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	require.Error(t, err)

	m := getMigrations(t, db)

	expectedMigrationsFound := 0
	require.Equal(t, expectedMigrationsFound, len(m))
}

func Test_Reverts_All_Attempted_Migrations_On_Failed_Migration_Attempt_Leaving_Previous_Migrations(t *testing.T) {
	// Arrange
	testMigrationsPath := migrationPath(t)
	defer cleanUpTestMigrations()
	defer func() {
		require.NoError(t, os.RemoveAll(testMigrationsPath))
	}()

	createMigrationFile(t, "main_table", mainTableUpMigration, mainTableDownMigration)
	err := sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)
	require.NoError(t, err)

	// Act
	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)
	createMigrationFile(t, "main_table2", "invalid", "SELECT 1;")
	err = sqlmigration.Run(testMigrationsPath, conf.DatabaseURL)

	// Assert
	require.Error(t, err)

	createMigrationFile(t, "dependant_table", depdendantTableUpMigration, depdendantTableDownMigration)
	m := getMigrations(t, db)

	expectedMigrationsFound := 1
	require.Equal(t, expectedMigrationsFound, len(m))
}

func cleanUpTestMigrations() {
	db.MustExec("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public")
}

func createMigrationFile(t *testing.T, name, upScript, downScript string) {
	testMigrationsPath := migrationPath(t)
	if _, err := os.Stat(testMigrationsPath); err != nil && errors.Is(err, os.ErrNotExist) {
		os.Mkdir(testMigrationsPath, 0755)
	}

	version := getMigrationVersion(t)

	upScriptName := strings.Join([]string{fmt.Sprintf("%d", version), name, "up", "sql"}, ".")
	downScriptName := strings.Join([]string{fmt.Sprintf("%d", version), name, "down", "sql"}, ".")

	err := os.WriteFile(path.Join(testMigrationsPath, upScriptName), []byte(upScript), 0777)
	require.NoError(t, err)

	err = os.WriteFile(path.Join(testMigrationsPath, downScriptName), []byte(downScript), 0777)
	require.NoError(t, err)
}

func getMigrationVersion(t *testing.T) int {
	testMigrationsPath := migrationPath(t)

	entities, err := os.ReadDir(testMigrationsPath)
	require.NoError(t, err)

	highestVersion := 0
 	for _, entry := range entities {
		if entry.IsDir() {
			continue
		}

		parts := strings.Split(entry.Name(), ".")
		require.Equal(t, 4, len(parts))

		versionPart := parts[0]

		version, err := strconv.Atoi(versionPart)
		require.NoError(t, err)

		if version > highestVersion {
			highestVersion = version
		}
	}

	return highestVersion + 1
}

func getMigrations(t *testing.T, db *sqlx.DB) []sqlmigration.Migration {
	var m []sqlmigration.Migration
	err := db.Select(&m, "SELECT * FROM schema_migration;")
	require.NoError(t, err)

	return m
}

func migrationPath(t *testing.T) string {
	return path.Join(rootPath, "test", "sql-migrations", "temp", t.Name())
}
