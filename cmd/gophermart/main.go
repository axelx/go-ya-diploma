package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"

	"io/ioutil"
	"sync"
	"time"

	//"fmt"
	"github.com/axelx/go-ya-diploma/internal/config"
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/handlers"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"github.com/axelx/go-ya-diploma/internal/orders"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	var wg sync.WaitGroup

	conf := config.NewConfigServer()
	lg := logger.Initialize("info")
	lg.Info("Running server", zap.String("config", conf.String()))

	db, err := core.InitDB(conf.FlagDatabaseDSN, lg)
	if err != nil {
		lg.Error("Error not connect to db", zap.String("about ERR", err.Error()))
	}

	chAddOrder := make(chan string, 100)
	chNewOrder := make(chan string, 100)
	chProcOrder := make(chan string, 500)
	countPerMin := 10 //00

	wg.Add(1)
	//Горутина добавления в accrual
	go func() {
		sleepMillisecond := 55 * 1000 / 60
		for {
			addToAccural(conf.FlagAccrualSystemAddress, <-chAddOrder, chNewOrder)
			time.Sleep(time.Millisecond * time.Duration(sleepMillisecond))
		}
	}()
	wg.Add(1)
	go func(countPerMin *int, db *sqlx.DB, lg *zap.Logger) {
		// точка отправки сообщений.

		for {
			sleepMillisecond := 55 * 1000 / *countPerMin
			select {
			case o := <-chNewOrder:
				fmt.Println("select chNewOrder:")
				checkAccural(conf.FlagAccrualSystemAddress, o, chProcOrder, countPerMin, db, lg)
			case o := <-chProcOrder:
				fmt.Println("select chProcOrder:", *countPerMin)
				checkAccural(conf.FlagAccrualSystemAddress, o, chProcOrder, countPerMin, db, lg)
			}

			time.Sleep(time.Millisecond * time.Duration(sleepMillisecond))
		}
	}(&countPerMin, db, lg)

	hd := handlers.New(lg, db, chAddOrder)
	if err := http.ListenAndServe(conf.FlagRunAddr, hd.Router()); err != nil {
		panic(err)
	}

	wg.Wait()

}

func checkAccural(urlAccrualServer, order string, chProcOrder chan string, countPerMin *int, db *sqlx.DB, lg *zap.Logger) {

	client := &http.Client{}
	resp2, err := client.Get(urlAccrualServer + "api/orders/" + order)
	if err != nil {
		fmt.Println("Error reporting metrics:", err)
		// скорее всего тут ошибку должен получить
		//*countPerMin += 5
	} else {
		fmt.Println("checkAccural")
		body, _ := ioutil.ReadAll(resp2.Body)
		resp2.Body.Close()

		var dat map[string]string
		json.Unmarshal(body, &dat)
		orders.UpdateStatus(db, lg, order, dat["status"])
		if dat["status"] == "PROCESSED" {
			fmt.Println("добавляем в канал процесс", order)
			chProcOrder <- order
		}
	}

}

func addToAccural(urlAccrualServer, order string, chNewOrder chan string) {

	client := &http.Client{}

	orderJSON, err := json.Marshal(map[string]string{"order": order})

	if err != nil {
		fmt.Println(err)
	}

	resp, err := client.Post(urlAccrualServer+"api/orders", "application/json", bytes.NewBuffer(orderJSON))
	if err != nil {
		fmt.Println("Error reporting metrics:", err)
	}

	fmt.Println("addAccrual", string(orderJSON), resp.StatusCode, 888)
	if resp.StatusCode == 202 {
		fmt.Println("добавляем в канал новые", order)
		chNewOrder <- order
	}

	resp.Body.Close()
}
