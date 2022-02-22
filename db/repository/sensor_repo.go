package repository

import (
	"iot/db/storage"
	"iot/model"
	"time"
)

type SensorRepo struct {
	db *storage.PSQLManager
}
type SensorRepository interface {
	InsertSensor(sensor *model.Sensor) error
	GetSensorById(Id int64) (*model.Sensor, error)
	GetSensorByTime(timeRequest time.Time) (*model.Sensor, error)
}

func NewSensorRepository(db *storage.PSQLManager) SensorRepository {
	return &SensorRepo{
		db: db,
	}
}
func (s *SensorRepo) InsertSensor(sensor *model.Sensor) error {
	if err := s.db.Create(sensor).Error; err != nil {
		return err
	}
	return nil
}
func (s *SensorRepo) GetSensorById(id int64) (*model.Sensor, error) {
	sensor := &model.Sensor{}
	if err := s.db.Where(&model.Sensor{Id: id}).First(sensor).Error; err != nil {
		return nil, err
	}
	return sensor, nil
}
func (s *SensorRepo) GetSensorByTime(timeRequest time.Time) (*model.Sensor, error) {
	sensor := &model.Sensor{}
	if err := s.db.Where(&model.Sensor{CreatedAt: timeRequest}).First(sensor).Error; err != nil {
		return nil, err
	}
	return sensor, nil
}
