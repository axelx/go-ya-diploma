package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"github.com/axelx/go-ya-diploma/internal/middleware"
	"github.com/axelx/go-ya-diploma/internal/orders"
	"github.com/axelx/go-ya-diploma/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type searcher interface {
	SearchOne(*sqlx.DB, *zap.Logger, string) (int, string)
	SearchMany(string) ([]int, []string)
}
type creater interface {
	Create(*sqlx.DB, *zap.Logger, string, string) error
}

type handler struct {
	ordS   searcher
	usrS   searcher
	ordC   creater
	usrC   creater
	Logger *zap.Logger
	db     *sqlx.DB
	chAdd  chan string
}

func New(ord, usr searcher, ust creater, log *zap.Logger, db *sqlx.DB, chAdd chan string) handler {
	return handler{
		ordS: ord,
		usrS: usr,
		//ordC:   ort,
		usrC:   ust,
		Logger: log,
		db:     db,
		chAdd:  chAdd,
	}
}

func (h *handler) Router() chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger(h.Logger))

	r.Post("/api/user/register", h.UserRegister())
	r.Post("/api/user/login", h.UserAuth())
	r.Post("/api/user/orders", middleware.CheckAuth(h.AddOrders(), h.Logger))
	r.Get("/api/user/orders", middleware.CheckAuth(h.Orders(), h.Logger))
	r.Get("/api/user/balance", middleware.CheckAuth(h.Balance(), h.Logger))
	r.Post("/api/user/balance/withdraw", middleware.CheckAuth(h.Withdraw(), h.Logger))
	r.Get("/api/user/withdrawals", middleware.CheckAuth(h.Withdrawals(), h.Logger))
	return r
}

func (h *handler) AddOrders() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		h.Logger.Info("AddOrders :")

		ro, _ := io.ReadAll(req.Body)
		order := string(ro)
		if !orders.LunaCheck(order, h.Logger) {
			h.Logger.Info("AddOrders : не прошёл проверку lunacheck ", zap.String("order", order))
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userIDcookie := user.GetIDviaCookie(req)

		//o, err := orders.FindOrder(h.db, h.Logger, order)
		usrID, ordN := h.find(h.ordS, order)
		fmt.Println("----", usrID, "-", ordN)

		if usrID > 0 && usrID == userIDcookie {
			fmt.Println("----", usrID, ordN)
			h.Logger.Info("AddOrders : заказ существует уже у этого пользователя", zap.String("order", order))
			res.WriteHeader(http.StatusOK)
			return
		} else if usrID > 0 && usrID != userIDcookie {
			fmt.Println("----", usrID, ordN)
			h.Logger.Info("AddOrders : заказ существует уже НО у другого пользователя", zap.String("order", order))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		if ordN == "" {
			h.Logger.Info("AddOrders : добавляем новый заказ", zap.String("order", order))
			err := orders.AddOrder(h.db, h.Logger, userIDcookie, order, 0, h.chAdd)
			if err != nil {
				h.Logger.Info("Error AddOrders :", zap.String("about ERR", err.Error()))
				http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
				return
			}
			res.WriteHeader(http.StatusAccepted)
		}

		h.Logger.Info("sending HTTP response",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
func (h *handler) Orders() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userID := user.GetIDviaCookie(req)
		os, err := orders.FindOrders(h.db, h.Logger, userID, h.chAdd)
		fmt.Println("----handlers Orders()", userID, os)
		if err != nil {
			h.Logger.Info("handler Orders", zap.String("orders.FindOrders", err.Error()))
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)

		if len(os) == 0 {
			_, err := res.Write([]byte("[]"))
			if err != nil {
				h.Logger.Info("Orders: not found any orders", zap.String("StatusInternalServerError", err.Error()))
				http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
				return
			}
			return
		} else {
			ordersJSON, err := json.Marshal(os)
			if err != nil {
				h.Logger.Info("handler Orders", zap.String("json.Marshal(os)", err.Error()))
			}
			_, err = res.Write(ordersJSON)
			if err != nil {
				h.Logger.Info("Orders", zap.String("StatusInternalServerError", err.Error()))
				http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
				return
			}
		}

		h.Logger.Info("sending HTTP response",
			//zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Balance() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Content-Type", "application/json")
		userID := user.GetIDviaCookie(req)
		ubs, err := user.Balance(h.db, h.Logger, userID)
		if err != nil {
			h.Logger.Info("handler Balance", zap.String("user.Balance", err.Error()))
		}
		balanceJSON, err := json.Marshal(ubs)
		if err != nil {
			h.Logger.Info("handler Balance", zap.String("json.Marshal(ubs)", err.Error()))
		}
		size, err := res.Write(balanceJSON)

		if err != nil {
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}

		h.Logger.Info("sending HTTP response",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Withdraw() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("------Withdraw--start---")
		ro, _ := io.ReadAll(req.Body)
		var dat map[string]interface{}
		err := json.Unmarshal(ro, &dat)
		if err != nil {
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}
		order := dat["order"].(string)
		wdrw := dat["sum"]
		sumWithdraw := wdrw.(float64)
		fmt.Println("------Withdraw----", order, wdrw)

		if !orders.LunaCheck(order, h.Logger) {
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userID := user.GetIDviaCookie(req)

		//ubs, err := user.Balance(h.db, h.Logger, userID)
		//if err != nil {
		//	h.Logger.Info("handler Withdraw", zap.String("user.Balance", err.Error()))
		//}
		//avBalance := ubs.Current - ubs.Withdrawn
		//if avBalan	ce < sumWithdraw {
		//	http.Error(res, "StatusPaymentRequired", http.StatusPaymentRequired)
		//	return
		//}
		fmt.Println("------Withdraw----4")

		o, err := orders.FindOrder(h.db, h.Logger, order)
		if err != nil {
			h.Logger.Info("handler Withdraw", zap.String("orders.FindOrder", err.Error()))
		}
		fmt.Println("------Withdraw----5")

		if o.UserID > 0 && o.UserID == userID {
			h.Logger.Info("AddOrders : заказ существует уже у этого пользователя", zap.String("order", order))
			res.WriteHeader(http.StatusOK)
			return
		} else if o.UserID > 0 && o.UserID != userID {
			h.Logger.Info("AddOrders : заказ существует уже НО у другого пользователя", zap.String("order", order))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		fmt.Println("------Withdraw----6")

		if err != nil {
			h.Logger.Info("AddOrders : добавляем новый заказ", zap.String("order", order))
			err = orders.AddOrder(h.db, h.Logger, userID, order, sumWithdraw, h.chAdd)
			if err != nil {
				h.Logger.Info("handler Withdraw", zap.String("orders.AddOrder", err.Error()))
			}
		}

		h.Logger.Info("sending HTTP response",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Withdrawals() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userID := user.GetIDviaCookie(req)
		os, err := orders.FindOrders(h.db, h.Logger, userID, h.chAdd)
		if err != nil {
			h.Logger.Info("handler Withdrawals", zap.String("orders.FindOrders", err.Error()))
		}

		res.Header().Set("Content-Type", "application/json")
		if len(os) == 0 {
			_, err := res.Write([]byte("[]"))

			if err != nil {
				h.Logger.Info("Withdrawals", zap.String("res.Write([]byte([]))", err.Error()))
				http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
				return
			}
			res.WriteHeader(http.StatusNoContent)
			return
		}

		ordersJSON, err := json.Marshal(os)
		if err != nil {
			h.Logger.Info("handler Withdrawals", zap.String("json.Marshal(os)", err.Error()))
		}
		size, err := res.Write(ordersJSON)

		if err != nil {
			h.Logger.Info("Withdrawals", zap.String("StatusInternalServerError", err.Error()))
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}

		h.Logger.Info("sending HTTP response",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) find(se searcher, findStr string) (int, string) {
	fmt.Println("---- find - SearchOne", se, findStr)

	i, s := se.SearchOne(h.db, h.Logger, findStr)
	fmt.Println("func (h *handler) find(se searcher, findStr string) (int, string)", i, s)
	return i, s
}

func (h *handler) findMany(se searcher) {
	i, s := se.SearchMany("_findMany")
	fmt.Println(i, s)
}

func (h *handler) create(c creater, firstStr, secondStr string) error {
	err := c.Create(h.db, h.Logger, firstStr, secondStr)
	return err
}
