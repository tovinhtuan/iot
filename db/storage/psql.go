package storage

import (
	"fmt"
	"log"

	"iot/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "changeme"
	dbname   = "iot"
)

type PSQLManager struct {
	*gorm.DB
}

func NewPSQLManager() (*PSQLManager, error) {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host = %s user = %s password = %s dbname = %s port = %d sslmode = disable TimeZone = Asia/Ho_Chi_Minh",
		host, user, password, dbname, port)))
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&model.Sensor{},
	)
	if err != nil {
		return nil, err
	}
	return &PSQLManager{db.Debug()}, nil
}
func (m *PSQLManager) Close() {
	sqlDB, _ := m.Debug().DB()
	err := sqlDB.Close()
	if err != nil {
		log.Fatalf("Could not close storage, err: %v", err)
	}
}
