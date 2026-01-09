package store

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func InitDb(strSet string) (*sql.DB, error) {
	var ps string
	ps = strSet
	if ps == "" {
		ps = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			`localhost`, `dburl`, `dburl`, `dburl`)
	}
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	// 4. Проверка подключения
	// if err = db.Ping(); err != nil {
	// 	db.Close()
	// 	return nil, fmt.Errorf("failed to connet to BD: %v", err)
	// }
	return db, nil
}
