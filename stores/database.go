package stores

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Open(ctx context.Context) (*sql.DB, error) {
	env, _ := ctx.Value("env").(map[string]string)

	db, err := sql.Open("pgx", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		env["DB_HOST"], env["DB_USERNAME"], env["DB_PASSWORD"], env["DB_NAME"], env["DB_PORT"], env["SSL_MODE"]))
	if err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	fmt.Println("😉😉😉Connected to DB😉😉😉")
	return db, nil
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()

	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("Migrate error: %v", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("Goose Up error: %v", err)
	}

	return nil
}
