package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pwera/di/garage"
	"github.com/pwera/di/helpers"
	"github.com/sarulabs/di"
)

func GetCarListHandler(w http.ResponseWriter, r *http.Request) {
	manager := di.Get(r, "car-manager").(*garage.CarManager)
	cars, err := manager.GetAll()

	if err == nil {
		helpers.JSONResponse(w, 200, cars)
		return
	}

	helpers.JSONResponse(w, 500, map[string]interface{}{
		"error": "Internal Error",
	})
}

// PostCarHandler is the handler that adds a new car.
func PostCarHandler(w http.ResponseWriter, r *http.Request) {
	var input *garage.Car

	err := helpers.ReadJSONBody(r, &input)
	if err != nil {
		helpers.JSONResponse(w, 400, map[string]interface{}{
			"error": "Could not decode request body.",
		})
		return
	}
	now := time.Now()
	manager := di.Get(r, "car-manager").(*garage.CarManager)
	car, err := manager.Create(input)
	fmt.Println("Times spend here {}", time.Now().Sub(now).String())

	if err == nil {
		helpers.JSONResponse(w, 200, car)
		return
	}

	switch e := err.(type) {
	case *helpers.ErrValidation:
		helpers.JSONResponse(w, 400, map[string]interface{}{
			"error": e.Error(),
		})
	default:
		helpers.JSONResponse(w, 500, map[string]interface{}{
			"error": "Internal Error",
		})
	}
}

// GetCarHandler is the handler that prints the characteristics of a car.
func GetCarHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["carId"]

	manager := di.Get(r, "car-manager").(*garage.CarManager)
	car, err := manager.Get(id)

	if err == nil {
		helpers.JSONResponse(w, 200, car)
		return
	}

	switch e := err.(type) {
	case *helpers.ErrNotFound:
		helpers.JSONResponse(w, 404, map[string]interface{}{
			"error": e.Error(),
		})
	default:
		helpers.JSONResponse(w, 500, map[string]interface{}{
			"error": "Internal Error",
		})
	}
}

// PutCarHandler is the handler that updates a car.
func PutCarHandler(w http.ResponseWriter, r *http.Request) {
	var input *garage.Car

	err := helpers.ReadJSONBody(r, &input)
	if err != nil {
		helpers.JSONResponse(w, 400, map[string]interface{}{
			"error": "Could not decode request body.",
		})
		return
	}

	id := mux.Vars(r)["carId"]

	manager := di.Get(r, "car-manager").(*garage.CarManager)
	car, err := manager.Update(id, input)

	if err == nil {
		helpers.JSONResponse(w, 200, car)
		return
	}

	switch e := err.(type) {
	case *helpers.ErrValidation:
		helpers.JSONResponse(w, 400, map[string]interface{}{
			"error": e.Error(),
		})
	case *helpers.ErrNotFound:
		helpers.JSONResponse(w, 404, map[string]interface{}{
			"error": e.Error(),
		})
	default:
		helpers.JSONResponse(w, 500, map[string]interface{}{
			"error": "Internal Error",
		})
	}
}

// DeleteCarHandler is the handler that removes a car from the database.
func DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["carId"]

	manager := di.Get(r, "car-manager").(*garage.CarManager)
	err := manager.Delete(id)

	if err == nil {
		w.WriteHeader(204)
		return
	}

	helpers.JSONResponse(w, 500, map[string]interface{}{
		"error": "Internal Error",
	})
}
