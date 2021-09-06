package app

import (
	"avito-challenge/app/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var a *Application

func TestMain(m *testing.M) {
	InitConfig()
	a = NewApplication()
	a.Redis = nil
	code := m.Run()
	os.Exit(code)
}

func clearTables() {
	a.DB.DB.Exec(`DELETE FROM "transaction"`)
	a.DB.DB.Exec("ALTER SEQUENCE transaction_id_seq RESTART WITH 1")
	_, err := a.DB.DB.Exec(`DELETE FROM "user"`)
	print(err)
	a.DB.DB.Exec("ALTER SEQUENCE user_id_seq RESTART WITH 1")
}

func createUser(balance int64) int64 {
	lastInsertId := 0
	err := a.DB.DB.QueryRow(`INSERT INTO "user" (balance) VALUES($1) RETURNING id`, balance).Scan(&lastInsertId)
	print(err)
	return int64(lastInsertId)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return rr
}

type fakeRedis struct {
}

func (r *fakeRedis) GetObj(key string) string {
	return `{"rates": {"USD": 2}}`
}
func (r *fakeRedis) SetObj(key, obj string) error {
	return nil
}

func TestUserDoesNotExist(t *testing.T) {
	clearTables()
	req, _ := http.NewRequest("GET", "/v1/user/1/balance", nil)
	response := executeRequest(req)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUserBalance(t *testing.T) {
	clearTables()
	userId := createUser(100)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/user/%d/balance", userId), nil)
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	balance := Balance{Currency: Config.Currency, Amount: 100}
	var result Balance
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, balance, result)
}

func TestUserBalanceUSD(t *testing.T) {
	clearTables()
	userId := createUser(100)
	a.Redis = &fakeRedis{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/user/%d/balance?currency=USD", userId), nil)
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	balance := Balance{Currency: "USD", Amount: 200}
	var result Balance
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, balance, result)
}

func TestTransactions(t *testing.T) {
	clearTables()
	userId := createUser(100)
	a.DB.AddBalance(userId, 30, "")
	time.Sleep(time.Second)
	a.DB.AddBalance(userId, 10, "")
	time.Sleep(time.Second)
	a.DB.AddBalance(userId, 20, "")
	transactions := []models.Transaction{}
	a.DB.DB.Select(&transactions, `SELECT * FROM "transaction" ORDER BY id`)

	// default sort order and page size
	req, _ := http.NewRequest("GET", fmt.Sprintf("/v1/user/%d/transactions", userId), nil)
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	var result []models.Transaction
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, 3, len(result))

	// sort by amount from largest to smallest
	req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/user/%d/transactions?sort=-amount", userId), nil)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, transactions[1], result[2])

	// sort transactions in chronological order
	req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/user/%d/transactions?sort=created_at", userId), nil)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, transactions[0], result[0])
	assert.Equal(t, transactions[1], result[1])
	assert.Equal(t, transactions[2], result[2])
}

func TestUserBalanceAddSuccess(t *testing.T) {
	clearTables()
	userId := createUser(100)
	jsonStr := `{"amount":100}`
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/user/%d/balance/add", userId), bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	balance := Balance{Currency: Config.Currency, Amount: 200}
	var result Balance
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, balance, result)
}

func TestUserBalanceAddBadPayload(t *testing.T) {
	jsonStr := `{"field":"value"}`
	req, _ := http.NewRequest("POST", "/v1/user/1/balance/add", bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestUserBalanceSubtractSuccess(t *testing.T) {
	clearTables()
	userId := createUser(100)
	jsonStr := `{"amount":90}`
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/user/%d/balance/subtract", userId), bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	balance := Balance{Currency: Config.Currency, Amount: 10}
	var result Balance
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.Equal(t, balance, result)
}

func TestUserBalanceSubtractNotEnoughMoney(t *testing.T) {
	clearTables()
	userId := createUser(100)
	jsonStr := `{"amount":110}`
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/user/%d/balance/subtract", userId), bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	user, _ := a.DB.GetUser(userId)
	assert.Equal(t, user.Balance, int64(100))
}

func TestUserBalanceTransferSuccess(t *testing.T) {
	clearTables()
	user1 := createUser(100)
	user2 := createUser(100)
	jsonStr := fmt.Sprintf(`{"amount":100, "user_id": %d}`, user2)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/user/%d/transfer", user1), bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
	user, _ := a.DB.GetUser(user1)
	assert.Equal(t, user.Balance, int64(0))
	user, _ = a.DB.GetUser(user2)
	assert.Equal(t, user.Balance, int64(200))
}

func TestUserBalanceTransferToHerself(t *testing.T) {
	clearTables()
	user1 := createUser(100)
	jsonStr := fmt.Sprintf(`{"amount":100, "user_id": %d}`, user1)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/v1/user/%d/transfer", user1), bytes.NewBufferString(jsonStr))
	response := executeRequest(req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}
