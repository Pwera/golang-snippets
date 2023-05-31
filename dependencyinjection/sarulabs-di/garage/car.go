package garage

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"

	"github.com/pwera/di/helpers"
)

type Car struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Brand string             `json:"brand" bson:"brand"`
	Color string             `json:"color" bson:"color"`
}

var colorsByBrand = map[string][]string{
	"audi":    {"black", "white", "yellow"},
	"porsche": {"black", "yellow"},
	"bmw":     {"red", "white"},
}

func brands() []string {
	var brands []string
	for brand := range colorsByBrand {
		brands = append(brands, brand)
	}
	return brands
}

func ValidateCar(car *Car) error {
	colors, ok := colorsByBrand[car.Brand]
	if !ok {
		return helpers.NewErrValidation(
			"Brand `" + car.Brand + "` does not exist. Available brands: " +
				strings.Join(brands(), ", "))
	}

	for _, color := range colors {
		if color == car.Color {
			return nil
		}
	}

	return helpers.NewErrValidation(
		"Color `" + car.Color + "` does not exist for `" + car.Brand +
			"`. Available colors: " + strings.Join(colors, ", "),
	)
}
