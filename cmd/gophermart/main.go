package main

import (
	//"fmt"
	"github.com/axelx/go-ya-diploma/internal/config"
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/handlers"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	conf := config.NewConfigServer()
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	db, err := core.InitDB(conf.FlagDatabaseDSN, lg)
	if err != nil {
		lg.Error("Error not connect to db", zap.String("about ERR", err.Error()))
	}

	hd := handlers.New(lg, db)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}

}
