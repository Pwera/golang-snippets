package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pwera/di/handlers"
	"github.com/pwera/di/logging"
	"github.com/pwera/di/middlewares"
	"github.com/pwera/di/services"
	"github.com/sarulabs/di"
)

func main() {

	defer func() {
		err := logging.Logger.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}()

	builder, err := di.NewBuilder()
	if err != nil {
		logging.Logger.Fatal(err.Error())
	}

	err = builder.Add(services.Services...)
	if err != nil {
		logging.Logger.Fatal(err.Error())
	}

	app := builder.Build()
	defer func() {
		if err2 := app.Delete(); err2 != nil {
			fmt.Println(err)
		}
	}()

	r := mux.NewRouter()

	m := func(h http.HandlerFunc) http.HandlerFunc {
		return middlewares.PanicRecoveryMiddleware(
			di.HTTPMiddleware(h, app, func(msg string) {
				logging.Logger.Error(msg)
			}),
			logging.Logger,
		)
	}
	//manager := di.Get(r, "car-manager").(*garage.CarManager)
	//manager.GetAll()
	r.HandleFunc("/cars", m(handlers.GetCarListHandler)).Methods("GET")
	r.HandleFunc("/cars", m(handlers.PostCarHandler)).Methods("POST")
	r.HandleFunc("/cars/{carId}", m(handlers.GetCarHandler)).Methods("GET")
	r.HandleFunc("/cars/{carId}", m(handlers.PutCarHandler)).Methods("PUT")
	r.HandleFunc("/cars/{carId}", m(handlers.DeleteCarHandler)).Methods("DELETE")

	port := "8080"
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	logging.Logger.Info("Listening on port " + port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Error(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logging.Logger.Info("Stopping the http server")
	if err := srv.Shutdown(ctx); err != nil {
		logging.Logger.Error(err.Error())
	}
}
