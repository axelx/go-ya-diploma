package pg

import (
	"context"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// подключение к базе
// миграции

func InitDB(url string, lg *zap.Logger) (*sqlx.DB, error) {

	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		lg.Error("Error not connect to pg",
			zap.String("URL", url),
			zap.String("about ERR", err.Error()))
		return db, err
	}
	db.SetMaxOpenConns(10)

	// миграции
	err = createTable(db)
	if err != nil {
		lg.Error("Error not connect to pg", zap.String("about ERR", err.Error()))
		return db, err
	}

	return db, nil
}

func DropTablesForTest(db *sqlx.DB, lg *zap.Logger) error {
	_, err := db.ExecContext(context.Background(), `DROP TABLE orders; DROP TABLE users;`)
	lg.Info("Drop table:", zap.String("drop", " success"))
	return err
}
