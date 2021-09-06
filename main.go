package main

import (
	"avito-challenge/app"
	"context"
	"fmt"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	if err := app.InitConfig(); err != nil {
		log.Panic(err)
	}
	application := app.NewApplication()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{app.Config.CorsOrigin},
		AllowCredentials: true,
	})
	handler := corsHandler.Handler(application.Router)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.HttpPort),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}
	go func() {
		log.Println(fmt.Sprintf("Listening at %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	wait := time.Second * time.Duration(15)
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
