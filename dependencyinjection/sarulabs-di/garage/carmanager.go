package garage

import (
	"github.com/pwera/di/helpers"
	"go.uber.org/zap"
)

type CarManager struct {
	Repo   *CarRepository
	Logger *zap.Logger
}

func (m *CarManager) GetAll() (*[]Car, error) {
	cars, err := m.Repo.FindAll()

	if cars == nil {
		cars = &[]Car{}
	} else if err != nil {
		m.Logger.Error(err.Error())
		return nil, err
	}
	return cars, err
}

func (m *CarManager) Get(id string) (car *Car, err error) {
	car, err = m.Repo.FindByID(id)

	if m.Repo.IsNotFoundErr(err) {
		return nil, helpers.NewErrNotFound("Car " + id + " does not exist")
	}

	if err != nil {
		m.Logger.Error(err.Error())
	}
	return car, err
}

func (m *CarManager) Create(car *Car) (*Car, error) {
	_, err := m.Repo.collection().InsertOne(nil, &car)
	if err != nil {
		return nil, err
	}

	return car, err
}

func (m *CarManager) Update(id string, car *Car) (*Car, error) {
	if err := ValidateCar(car); err != nil {
		return nil, err
	}

	err := m.Repo.Update(car)

	if m.Repo.IsNotFoundErr(err) {
		return nil, helpers.NewErrNotFound("Car " + id + " does not exist")
	}

	if err != nil {
		m.Logger.Error(err.Error())
		return nil, err
	}
	return car, err
}

func (m *CarManager) Delete(id string) error {
	err := m.Repo.Delete(id)

	if m.Repo.IsNotFoundErr(err) {
		return nil
	}
	if err != nil {
		m.Logger.Error(err.Error())
	}
	return err
}
