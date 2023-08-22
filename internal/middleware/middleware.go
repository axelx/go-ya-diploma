package middleware

import (
	"errors"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func CheckAuth(h http.HandlerFunc, lg *zap.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := r.Cookie("auth")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				log.Println(err)
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}

		if err != nil {
			lg.Error("order FindOrders", zap.String("err", err.Error()))
		}

		h(w, r)
	})
}
