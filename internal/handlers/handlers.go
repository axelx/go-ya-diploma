package handlers

import (
	"encoding/json"
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

type handler struct {
	Logger *zap.Logger
	db     *sqlx.DB
	chAdd  chan string
}

func New(log *zap.Logger, db *sqlx.DB, chAdd chan string) handler {
	return handler{
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
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userID := user.GetIDviaCookie(req)

		o, err := orders.FindOrder(h.db, h.Logger, order)

		if o.UserID > 0 && o.UserID == userID {
			h.Logger.Info("AddOrders : заказ существует уже у этого пользователя", zap.String("order", order))
			res.WriteHeader(http.StatusOK)
			return
		} else if o.UserID > 0 && o.UserID != userID {
			h.Logger.Info("AddOrders : заказ существует уже НО у другого пользователя", zap.String("order", order))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		if err != nil {
			h.Logger.Info("AddOrders : добавляем новый заказ", zap.String("order", order))
			err = orders.AddOrder(h.db, h.Logger, userID, order, 0, h.chAdd)
			if err != nil {
				h.Logger.Error("Error AddOrders :", zap.String("about ERR", err.Error()))
				http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
				return
			}
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
func (h *handler) Orders() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userID := user.GetIDviaCookie(req)
		os, err := orders.FindOrders(h.db, h.Logger, userID)

		if len(os) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		ordersJSON, err := json.Marshal(os)
		size, err := res.Write(ordersJSON)

		if err != nil {
			h.Logger.Error("Orders", zap.String("StatusInternalServerError", err.Error()))
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Balance() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userID := user.GetIDviaCookie(req)
		ubs, err := user.Balance(h.db, h.Logger, userID)
		balanceJSON, err := json.Marshal(ubs)
		size, err := res.Write(balanceJSON)

		if err != nil {
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Withdraw() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ro, _ := io.ReadAll(req.Body)
		var dat map[string]interface{}
		err := json.Unmarshal(ro, &dat)
		if err != nil {
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}
		var order string
		order = dat["order"].(string)

		wdrw := dat["sum"]
		sumWithdraw := wdrw.(float64)

		if !orders.LunaCheck(order, h.Logger) {
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userID := user.GetIDviaCookie(req)

		ubs, err := user.Balance(h.db, h.Logger, userID)
		avBalance := ubs.Current - ubs.Withdrawn

		if avBalance < sumWithdraw {
			http.Error(res, "StatusPaymentRequired", http.StatusPaymentRequired)
			return
		}

		o, err := orders.FindOrder(h.db, h.Logger, order)

		if o.UserID > 0 && o.UserID == userID {
			h.Logger.Info("AddOrders : заказ существует уже у этого пользователя", zap.String("order", order))
			res.WriteHeader(http.StatusOK)
			return
		} else if o.UserID > 0 && o.UserID != userID {
			h.Logger.Info("AddOrders : заказ существует уже НО у другого пользователя", zap.String("order", order))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		if err != nil {
			h.Logger.Info("AddOrders : добавляем новый заказ", zap.String("order", order))
			err = orders.AddOrder(h.db, h.Logger, userID, order, sumWithdraw, h.chAdd)
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Withdrawals() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userID := user.GetIDviaCookie(req)
		os, err := orders.FindOrders(h.db, h.Logger, userID)

		if len(os) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		ordersJSON, err := json.Marshal(os)
		size, err := res.Write(ordersJSON)

		if err != nil {
			h.Logger.Info("Orders", zap.String("StatusInternalServerError", err.Error()))
			http.Error(res, "StatusInternalServerError", http.StatusInternalServerError)
			return
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
