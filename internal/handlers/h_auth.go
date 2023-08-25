package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/user"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (h *handler) UserRegister() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		h.Logger.Debug("decoding request")
		var u models.User

		body, _ := io.ReadAll(req.Body)
		fmt.Println("----1", body)
		err := json.Unmarshal([]byte(body), &u)
		fmt.Println("----2", u)

		if err != nil {
			h.Logger.Debug("cannot decode request JSON body", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if u.Login == "" || u.Password == "" {
			http.Error(res, "StatusBadRequest", http.StatusBadRequest)
			return
		}

		usrID, usrL := h.find(h.usrS, u.Login)
		fmt.Println("----3", usrID, usrL)

		if usrID == 0 {
			fmt.Println("----4", usrID, usrL, usrID == 0)

			h.Logger.Info("CreateNewUser :", zap.String("user_id", u.Login))
			err := h.create(h.usrC, u.Login, u.Password)
			if err != nil {
				h.Logger.Error("CreateNewUser :", zap.String("err", err.Error()))
			}
		} else {
			fmt.Println("----5", usrID, usrL, usrID == 0)

			h.Logger.Info("StatusConflict :", zap.String("user_id", usrL))
			http.Error(res, "StatusConflict", http.StatusConflict)
			return
		}
		res.Header().Set("Content-Type", "application/json")

		if cookie, b := user.AuthUser(h.db, h.Logger, u.Login, u.Password); b {
			fmt.Println("----6", cookie, b)

			http.SetCookie(res, &cookie)
			res.WriteHeader(http.StatusOK)
		} else {
			res.WriteHeader(http.StatusUnauthorized)
		}

		res.WriteHeader(http.StatusOK)
		size, err := res.Write([]byte("{\"login\":\"" + u.Login + "\", \"password\":\"" + u.Password + "\"}"))
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
		fmt.Println("----7 end")

	}
}
func (h *handler) UserAuth() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		h.Logger.Debug("decoding request")
		var u models.User

		body, _ := io.ReadAll(req.Body)
		err := json.Unmarshal([]byte(body), &u)

		if err != nil {
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
