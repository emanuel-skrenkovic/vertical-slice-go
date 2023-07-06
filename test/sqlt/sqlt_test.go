package sqlt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/config"
	"github.com/eskrenkovic/vertical-slice-go/internal/test"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/eskrenkovic/tql"
	_ "github.com/lib/pq"
)

var db *sql.DB

func TestMain(m *testing.M) {
	rootPath := "../../"

	localConfigPath := path.Join(rootPath, "config.local.env")
	if _, err := os.Stat(localConfigPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(localConfigPath)
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Fatal(err)
				}
			}()

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

	conf, err := config.Load()
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

	db, err = sql.Open("postgres", conf.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec("CREATE TABLE test (id text, nullable text);"); err != nil {
		log.Fatal(err)
	}

	_ = m.Run()

	if err := recover(); err != nil {
		log.Println(err)
	}

	if _, err := db.Exec("DROP TABLE test;"); err != nil {
		log.Println(err)
	}

	if err := fixture.Stop(); err != nil {
		log.Fatal(err)
	}
}

type result struct {
	ID       string  `db:"id"`
	Nullable *string `db:"nullable"`
}

func Test_QueryOne(t *testing.T) {
	// Arrange
	id := uuid.New()
	nullable := uuid.New()

	_, err := db.Exec(fmt.Sprintf("INSERT INTO test VALUES ('%s', '%s');", id.String(), nullable.String()))
	require.NoError(t, err)

	// Act
	r, err := tql.QueryFirst[result](context.Background(), db, "SELECT id, nullable FROM test;")

	// Assert
	require.NoError(t, err)
	require.Equal(t, id.String(), r.ID)
	require.NotNil(t, r.Nullable)
	require.Equal(t, nullable.String(), *r.Nullable)
}

func Test_QueryOne_String(t *testing.T) {
	// Arrange
	id := uuid.New()
	nullable := uuid.New()

	_, err := db.Exec(fmt.Sprintf("INSERT INTO test (id, nullable) VALUES ('%s', '%s');", id.String(), nullable.String()))
	require.NoError(t, err)

	// Act
	r, err := tql.QueryFirst[string](context.Background(), db, "SELECT id FROM test WHERE id = $1;", id)

	// Assert
	require.NoError(t, err)
	require.Equal(t, id.String(), r)
}

func Test_QueryOne_String_Pointer(t *testing.T) {
	// Arrange
	id := uuid.New()
	nullable := uuid.New()

	_, err := db.Exec(fmt.Sprintf("INSERT INTO test (id, nullable) VALUES ('%s', '%s');", id.String(), nullable.String()))
	require.NoError(t, err)

	// Act
	r, err := tql.QueryFirst[*string](context.Background(), db, "SELECT id FROM test WHERE id = $1;", id)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, id.String(), *r)
}

func Test_QueryOne_Int_Pointer(t *testing.T) {
	// Arrange
	id := uuid.New()
	nullable := uuid.New()

	_, err := db.Exec(fmt.Sprintf("INSERT INTO test (id, nullable) VALUES ('%s', '%s');", id.String(), nullable.String()))
	require.NoError(t, err)

	// Act
	r, err := tql.QueryFirst[*int](context.Background(), db, "SELECT 420;")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 420, *r)
}

func Test_Query(t *testing.T) {
	// Arrange
	_, err := db.Exec("INSERT INTO test (id, nullable) VALUES ('asdf', 'fdsa');")
	require.NoError(t, err)

	// Act
	r, err := tql.Query[result](context.Background(), db, "SELECT id, nullable FROM test;")

	// Assert
	require.NoError(t, err)
	require.Len(t, r, 5)
	require.Equal(t, "asdf", r[4].ID)
	require.NotNil(t, r[4].Nullable)
	require.Equal(t, "fdsa", *r[4].Nullable)
}
func Test_Query_Basic_Type(t *testing.T) {
	// Arrange
	tx, _ := db.BeginTx(context.Background(), &sql.TxOptions{})

	// Act
	r, err := tql.Query[string](context.Background(), tx, "SELECT id FROM test;")

	require.NoError(t, tx.Commit())

	// Assert
	require.NoError(t, err)
	require.Len(t, r, 5)
	require.Equal(t, "asdf", r[4])
	require.NotNil(t, r[4])
}

func Test_Query_Basic_Type_Pointer(t *testing.T) {
	// Act
	r, err := tql.Query[*string](context.Background(), db, "SELECT id FROM test;")

	// Assert
	require.NoError(t, err)
	require.Len(t, r, 5)
	require.Equal(t, "asdf", *r[4])
	require.NotNil(t, r[4])
}

func Test_Query_Basic_Type_Pointer_Null(t *testing.T) {
	// Act
	r, err := tql.QueryFirst[*string](context.Background(), db, "SELECT NULL;")

	// Assert
	require.NoError(t, err)
	require.Nil(t, r)
}

func Test_Query_Empty_Result(t *testing.T) {
	_, err := db.Exec("INSERT INTO test VALUES ('asdf', 'fdsa');")
	require.NoError(t, err)

	// Act
	r, err := tql.Query[result](context.Background(), db, "SELECT * FROM test WHERE id = '';")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, r)
}

func Test_Exec(t *testing.T) {
	// Act
	const insertStmt = "INSERT INTO test (id, nullable) VALUES (:test, :test2);"
	_, err := tql.Exec(context.Background(), db, insertStmt, map[string]any{
		"test":  "totally_new_id",
		"test2": "totally_new_id_2",
	})

	// Assert
	require.NoError(t, err)
	r, err := tql.QueryFirst[result](context.Background(), db, "SELECT * FROM test WHERE id = $1;", "totally_new_id")

	require.NotEmpty(t, r)
	require.Equal(t, "totally_new_id", r.ID)
	require.Equal(t, "totally_new_id_2", *r.Nullable)
	require.NoError(t, err)
	require.NoError(t, err)
}

func Test_Exec_With_Struct(t *testing.T) {
	// Act
	id := uuid.NewString()
	userID := uuid.NewString()
	params := struct {
		ID     string `db:"test"`
		UserID string `db:"test2"`
	}{
		ID:     id,
		UserID: userID,
	}
	const insertStmt = "INSERT INTO test (id, nullable) VALUES (:test, :test2);"
	_, err := tql.Exec(context.Background(), db, insertStmt, params)

	// Assert
	require.NoError(t, err)
	r, err := tql.QueryFirst[result](context.Background(), db, "SELECT * FROM test WHERE id = $1;", id)

	require.NotEmpty(t, r)
	require.Equal(t, id, r.ID)
	require.Equal(t, userID, *r.Nullable)
	require.NoError(t, err)
	require.NoError(t, err)
}

func Test_Exec_Not_Named(t *testing.T) {
	// Arrange
	id := uuid.NewString()
	userID := uuid.NewString()
	const insertStmt = "INSERT INTO test (id, nullable) VALUES ($1, $2);"

	// Act
	_, err := tql.Exec(context.Background(), db, insertStmt, id, userID)

	// Assert
	require.NoError(t, err)
	r, err := tql.QueryFirst[result](context.Background(), db, "SELECT * FROM test WHERE id = $1;", id)

	require.NotEmpty(t, r)
	require.Equal(t, id, r.ID)
	require.Equal(t, userID, *r.Nullable)
	require.NoError(t, err)
	require.NoError(t, err)
}

func Test_Exec_Mixed_Named_Positional(t *testing.T) {
	// Arrange
	id := uuid.NewString()
	userID := uuid.NewString()

	// Act
	const insertStmt = "INSERT INTO test (id, nullable) VALUES ($1, :test2);"
	_, err := tql.Exec(context.Background(), db, insertStmt, id, userID, map[string]any{"test2": "asdf"})

	// Assert
	require.Error(t, err)
	require.Equal(t, "mixed positional and named parameters", err.Error())
	//require.ErrorIs(t, err, fmt.Errorf("mixed positional and named parameters"))

	r, err := tql.QueryFirst[result](context.Background(), db, "SELECT * FROM test WHERE id = $1;", id)
	require.NoError(t, err)
	require.Empty(t, r)
}
