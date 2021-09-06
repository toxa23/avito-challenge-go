package app

import (
	"avito-challenge/app/models"
	"net/http"
	"strconv"
	"strings"
)

type contextKey int

const (
	userKey contextKey = iota
)

func (app *Application) getBalance(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*models.User)
	currency := r.URL.Query().Get("currency")
	app.responseBalance(w, r, user.Balance, currency)
}

func (app *Application) getTransactions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*models.User)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = Config.PerPage
	}
	sortBy := strings.ToLower(r.URL.Query().Get("sort"))
	if sortBy == "" {
		sortBy = "-created_at"
	}
	sortDir := ""
	if strings.HasPrefix(sortBy, "-") {
		sortDir = " desc"
		sortBy = strings.TrimPrefix(sortBy, "-")
	}
	if sortBy != "amount" && sortBy != "created_at" {
		msg := "invalid 'sort' attribute. Valid values are: ['created_at', 'amount']"
		responseError(w, msg, http.StatusBadRequest)
		return
	}
	sortBy += sortDir
	result, err := app.DB.GetUserTransactions(user.Id, sortBy, page, limit)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

func (app *Application) addBalance(w http.ResponseWriter, r *http.Request) {
	data, err := app.getPayload(r)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := r.Context().Value(userKey).(*models.User)
	balance, err := app.DB.AddBalance(user.Id, data.Amount, data.Details)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.responseBalance(w, r, balance, "")
}

func (app *Application) subtractBalance(w http.ResponseWriter, r *http.Request) {
	data, err := app.getPayload(r)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := r.Context().Value(userKey).(*models.User)
	if user.Balance < data.Amount {
		responseError(w, "not enough money on user balance", http.StatusBadRequest)
		return
	}
	balance, err := app.DB.SubtractBalance(user.Id, data.Amount, data.Details)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.responseBalance(w, r, balance, "")
}

func (app *Application) transfer(w http.ResponseWriter, r *http.Request) {
	data, err := app.getPayload(r)
	if err != nil {
		responseError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fromUser := r.Context().Value(userKey).(*models.User)
	if fromUser.Balance < data.Amount {
		responseError(w, "not enough money on user balance", http.StatusBadRequest)
		return
	}
	toUser, err := app.DB.GetUser(data.UserId)
	if err != nil || toUser == nil {
		responseError(w, "recipient not found", http.StatusNotFound)
		return
	}
	if fromUser.Id == toUser.Id {
		responseError(w, "cannot transfer money to herself", http.StatusBadRequest)
		return
	}
	err = app.DB.TransferMoney(fromUser.Id, toUser.Id, data.Amount)
	if err != nil {
		responseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result := map[string]string{"status": "responseBalance"}
	jsonResponse(w, result, http.StatusOK)
}
