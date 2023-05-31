package garage

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type CarRepository struct {
	Client *mongo.Client
}

func (repo *CarRepository) collection() *mongo.Collection {
	return repo.Client.Database("dingo_car_api").Collection("cars")
}

func (repo *CarRepository) FindAll() (*[]Car, error) {
	var cars []Car
	cur, err := repo.collection().Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	if err = cur.All(context.TODO(), &cars); err != nil {
		log.Fatal(err)
	}

	return &cars, err
}

func (repo *CarRepository) FindByID(id string) (*Car, error) {
	var car Car
	filter := bson.D{{"_id", id}}
	err := repo.collection().FindOne(nil, filter).Decode(&car)
	return &car, err
}

func (repo *CarRepository) Insert(car *Car) error {
	_, err := repo.collection().InsertOne(nil, &car)
	return err
}

func (repo *CarRepository) Update(car *Car) error {
	filter := bson.D{{"_id", car.ID}}
	update := bson.D{{"$set", bson.D{{"brand", car.Brand},
		{"color", car.Color}}}}

	_, err := repo.collection().UpdateOne(nil, filter, update)

	return err
}

func (repo *CarRepository) Delete(id string) error {
	filter := bson.D{{"_id", id}}
	opts := options.Delete().SetHint(bson.D{{"_id", id}})

	_, err := repo.collection().DeleteOne(nil, filter, opts)
	return err
}

func (repo *CarRepository) IsNotFoundErr(err error) bool {
	return err != nil && err.Error() == "mongo: no documents in result"
}

func (repo *CarRepository) IsAlreadyExistErr(err error) bool {
	return false
}
