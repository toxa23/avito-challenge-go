package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (app *Application) getUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		userId, _ := strconv.Atoi(vars["id"])
		user, err := app.DB.GetUser(int64(userId))
		if err != nil {
			responseError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if user == nil {
			responseError(w, "user not found", http.StatusNotFound)
			return
		}
		ctx = context.WithValue(ctx, userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) getPayload(r *http.Request) (*InputPayload, error) {
	var p InputPayload
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		return nil, errors.New("payload is invalid")
	}
	if p.Amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	return &p, nil
}

func jsonResponse(w http.ResponseWriter, payload interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func responseError(w http.ResponseWriter, msg string, code int) {
	result := map[string]string{"message": msg}
	jsonResponse(w, result, code)
}

func (app *Application) responseBalance(w http.ResponseWriter, r *http.Request, balance int64, currency string) {
	amount := float64(balance)
	if currency == "" {
		currency = Config.Currency
	} else if currency != Config.Currency {
		rate, err := app.getCbrRate(currency)
		if err == nil {
			amount *= rate
		}
	}
	result := Balance{Amount: amount, Currency: currency}
	jsonResponse(w, result, http.StatusOK)
}
