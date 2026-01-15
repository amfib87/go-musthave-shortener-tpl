package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

func InitDB(strSet string) (db *sql.DB, err error) {
	db, err = sql.Open("pgx", strSet)
	if err != nil {
		return nil, err
	}

	// Применяем миграции
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("ошибка миграций: %w", err)
	}
	return db, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("failed migration: %w", err)
	}

	return nil
}
