package core

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// подключение к базе
// миграции

func InitDB(url string, lg *zap.Logger) (*sqlx.DB, error) {

	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		lg.Error("Error not connect to db", zap.String("about ERR", err.Error()))
		return db, err
	}
	db.SetMaxOpenConns(10)

	// миграции
	err = createTable(db)
	if err != nil {
		lg.Error("Error not connect to db", zap.String("about ERR", err.Error()))
		return db, err
	}

	return db, nil
}
