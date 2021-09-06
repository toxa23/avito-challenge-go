package app

import (
	"avito-challenge/app/db"
	"avito-challenge/app/redis"
	"github.com/gorilla/mux"
)

type Application struct {
	DB     *db.PostgresService
	Redis  redis.IRedisService
	Router *mux.Router
}

func NewApplication() *Application {
	app := Application{
		DB:     db.NewPostgresService(Config.DbUrl),
		Redis:  redis.NewRedisService(Config.RedisUrl),
		Router: mux.NewRouter(),
	}

	r := app.Router
	r.HandleFunc("/v1/user/{id}/balance", app.getBalance).Methods("GET")
	r.HandleFunc("/v1/user/{id}/transactions", app.getTransactions).Methods("GET")
	r.HandleFunc("/v1/user/{id}/balance/add", app.addBalance).Methods("POST")
	r.HandleFunc("/v1/user/{id}/balance/subtract", app.subtractBalance).Methods("POST")
	r.HandleFunc("/v1/user/{id}/transfer", app.transfer).Methods("POST")
	r.Use(app.getUserMiddleware)

	return &app
}
