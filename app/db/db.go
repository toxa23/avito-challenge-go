package db

import (
	"avito-challenge/app/models"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

type PostgresService struct {
	DB *sqlx.DB
}

func NewPostgresService(connectionStr string) *PostgresService {
	db, err := sqlx.Open("postgres", connectionStr)
	if err != nil {
		log.Panicf(err.Error())
	}

	return &PostgresService{
		DB: db,
	}
}

func (svc *PostgresService) GetUser(userId int64) (*models.User, error) {
	result := &models.User{}
	query := `SELECT id, balance FROM "user" WHERE id=$1`
	err := svc.DB.Get(result, query, userId)
	if err == sql.ErrNoRows {
		err = nil
		result = nil
	}
	return result, err
}

func (svc *PostgresService) addBalance(tx *sql.Tx, userId, amount int64, details string) (int64, error) {
	query := `UPDATE "user" SET balance = balance + $2 WHERE id=$1 RETURNING balance`
	balance := int64(0)
	err := tx.QueryRow(query, userId, amount).Scan(&balance)
	if err == nil {
		query = `INSERT INTO "transaction" (user_id, amount, details) VALUES($1, $2, $3)`
		_, err = tx.Exec(query, userId, amount, details)
	}
	return balance, err
}

func (svc *PostgresService) AddBalance(userId, amount int64, details string) (int64, error) {
	tx, err := svc.DB.Begin()
	balance, err := svc.addBalance(tx, userId, amount, details)
	if err == nil {
		err = tx.Commit()
	} else {
		tx.Rollback()
	}
	return balance, err
}

func (svc *PostgresService) SubtractBalance(userId, amount int64, details string) (int64, error) {
	return svc.AddBalance(userId, -amount, details)
}

func (svc *PostgresService) TransferMoney(fromUser, toUser, amount int64) error {
	tx, err := svc.DB.Begin()
	_, err = svc.addBalance(tx, fromUser, -amount, fmt.Sprintf("transfer to %d", toUser))
	if err == nil {
		_, err = svc.addBalance(tx, toUser, amount, fmt.Sprintf("transfer from %d", fromUser))
	}
	if err == nil {
		err = tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

func (svc *PostgresService) GetUserTransactions(
	userId int64, sortBy string, page, limit int) ([]models.Transaction, error) {

	result := []models.Transaction{}
	offset := (page - 1) * limit
	query := `SELECT id, amount, created_at, details 
			  FROM "transaction" WHERE user_id=$1 ORDER BY %s OFFSET $2 LIMIT $3`
	query = fmt.Sprintf(query, sortBy)
	err := svc.DB.Select(&result, query, userId, offset, limit)
	return result, err
}
