package models

type User struct {
	Id      int64 `json:"id" db:"id"`
	Balance int64 `json:"balance" db:"balance"`
}

type Transaction struct {
	Id        int64  `json:"id" db:"id"`
	Amount    int64  `json:"amount" db:"amount"`
	Details   string `json:"details" db:"details"`
	CreatedAt string `json:"created_at" db:"created_at"`
}
