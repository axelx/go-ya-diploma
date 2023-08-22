package handlers

import (
	"encoding/json"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/user"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (h *handler) UserRegister() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		h.Logger.Debug("decoding request")
		var u models.User
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&u); err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if u.Login == "" || u.Password == "" {
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		usr, err := user.FindUserByLogin(h.db, h.Logger, u.Login)
		if err != nil {
			h.Logger.Info("CreateNewUser :", zap.String("user_id", u.Login))
			user.CreateNewUser(h.db, h.Logger, u.Login, u.Password)
		}
		if usr.Login != "" {
			h.Logger.Info("StatusConflict :", zap.String("user_id", usr.Login))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}

		if cookie, b := user.AuthUser(h.db, h.Logger, u.Login, u.Password); b {
			http.SetCookie(res, &cookie)
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusUnauthorized)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		size, err := res.Write([]byte("HELLO"))
		if err != nil {
			h.Logger.Error("Error UserRegister",
				zap.String("about func", "UserRegister"),
				zap.String("about ERR", err.Error()),
			)
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("size", strconv.Itoa(size)),
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
func (h *handler) UserAuth() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		h.Logger.Debug("decoding request")
		var u models.User
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&u); err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if u.Login == "" || u.Password == "" {
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		if cookie, b := user.AuthUser(h.db, h.Logger, u.Login, u.Password); b {
			http.SetCookie(res, &cookie)
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusUnauthorized)
		}

		h.Logger.Info("sending HTTP response UpdatedMetric",
			zap.String("status", strconv.Itoa(http.StatusOK)),
		)
	}
}
