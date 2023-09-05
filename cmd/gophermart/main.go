package main

import (
	"encoding/json"
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/orderservice"
	"github.com/axelx/go-ya-diploma/internal/pg"
	"github.com/axelx/go-ya-diploma/internal/userservice"
	"github.com/axelx/go-ya-diploma/internal/utils"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"strconv"

	"io"
	"sync"
	"time"

	"github.com/axelx/go-ya-diploma/internal/config"
	"github.com/axelx/go-ya-diploma/internal/handlers"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	var wg sync.WaitGroup

	conf := config.NewConfigServer()
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	db, err := pg.InitDB(conf.FlagDatabaseDSN, lg)
	if err != nil {
		lg.Error("Error not connect to pg", zap.String("about ERR", err.Error()))
	}

	ord := orderservice.Order{DB: db, LG: lg}
	usr := userservice.User{DB: db, LG: lg}

	chNewOrder := make(chan string, 100)
	chProcOrder := make(chan string, 500)
	countPerMin := 10 //00

	wg.Add(1)
	go func(countPerMin *int, db *sqlx.DB, lg *zap.Logger) {
		// точка отправки сообщений.
		for {
			sleepMillisecond := 55 * 1000 / *countPerMin
			select {
			case o := <-chNewOrder:
				checkAccural(conf.FlagAccrualSystemAddress, o, chProcOrder, countPerMin, db, lg)
			case o := <-chProcOrder:
				checkAccural(conf.FlagAccrualSystemAddress, o, chProcOrder, countPerMin, db, lg)
			}

			time.Sleep(time.Millisecond * time.Duration(sleepMillisecond))
		}
	}(&countPerMin, db, lg)

	hd := handlers.New(ord, usr, lg, db, chNewOrder)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}

	wg.Wait()

}

func checkAccural(urlAccrualServer, order string, chProcOrder chan string, countPerMin *int, db *sqlx.DB, lg *zap.Logger) {
	client := &http.Client{}
	resp2, err := client.Get(urlAccrualServer + "api/orders/" + order)
	if err != nil {
		lg.Error("main checkAccural", zap.String("response get urlAccrualServer", err.Error()))

		// скорее всего тут ошибку должен получить
		//*countPerMin += 5
	} else {
		body, _ := io.ReadAll(resp2.Body)
		resp2.Body.Close()

		var dat map[string]interface{}
		json.Unmarshal(body, &dat)

		lg.Info("main checkAccural", zap.String("response accrual", string(body)), zap.String("response body.status", strconv.Itoa(resp2.StatusCode)))
		lg.Info("main checkAccural", zap.String("dat[\"status\"]", fmt.Sprintf("%v", dat["status"])))
		orderservice.UpdateStatus(db, lg, order, fmt.Sprintf("%v", dat["status"]), utils.GetFloat(dat["accrual"]))
		if dat["status"] == "PROCESSING" {
			lg.Info("main checkAccural", zap.String("добавляем в канал процесс", order))
			chProcOrder <- order
		}
	}

}
