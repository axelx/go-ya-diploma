package handlers

import (
	"encoding/json"
	"github.com/axelx/go-ya-diploma/internal/logger"
	"github.com/axelx/go-ya-diploma/internal/middleware"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/orders"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type orderer interface {
	SearchIDs(string) (int, string)
	SearchMany(int) ([]models.Order, error)
	FindOrder(string) (models.Order, error)
	LunaCheck(string) bool
	Create(int, string, float64, chan string) error
}

type userdo interface {
	SearchOne(string) (int, string)
	Create(string, string) error
	AuthUser(string, string) (http.Cookie, bool)
	GetIDviaCookie(req *http.Request) int
	Balance(int) (models.Balance, error)
}

type handler struct {
	orderer orderer
	userdo  userdo
	Logger  *zap.Logger
	db      *sqlx.DB
	chAdd   chan string
}

func New(orderer orderer, userdo userdo, log *zap.Logger, db *sqlx.DB, chAdd chan string) handler {
	return handler{
		orderer: orderer,
		userdo:  userdo,
		Logger:  log,
		db:      db,
		chAdd:   chAdd,
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
		if !h.orderer.LunaCheck(order) {
			h.Logger.Info("AddOrders : не прошёл проверку lunacheck ", zap.String("order", order))
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userIDcookie := h.userdo.GetIDviaCookie(req)

		usrID, ordN := h.orderer.SearchIDs(order)

		if usrID > 0 && usrID == userIDcookie {
			h.Logger.Info("AddOrders : заказ существует уже у этого пользователя", zap.String("order", order))
			res.WriteHeader(http.StatusOK)
			return
		} else if usrID > 0 && usrID != userIDcookie {
			h.Logger.Info("AddOrders : заказ существует уже НО у другого пользователя", zap.String("order", order))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		if ordN == "" {
			h.Logger.Info("AddOrders : добавляем новый заказ", zap.String("order", order))
			err := h.orderer.Create(userIDcookie, order, 0, h.chAdd)
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

		userID := h.userdo.GetIDviaCookie(req)
		os, err := h.orderer.SearchMany(userID)
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
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}

func (h *handler) Balance() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Content-Type", "application/json")
		userID := h.userdo.GetIDviaCookie(req)
		ubs, err := h.userdo.Balance(userID)
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

		if !h.orderer.LunaCheck(order) {
			http.Error(res, "StatusUnprocessableEntity", http.StatusUnprocessableEntity)
			return
		}
		userID := h.userdo.GetIDviaCookie(req)

		o, err := h.orderer.FindOrder(order)
		if err != nil {
			h.Logger.Info("handler Withdraw", zap.String("orders.FindOrder", err.Error()))
		}

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
			err = h.orderer.Create(userID, order, sumWithdraw, h.chAdd)
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

		userID := h.userdo.GetIDviaCookie(req)
		os, err := orders.FindWithdrawalsOrders(h.db, h.Logger, userID)
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
