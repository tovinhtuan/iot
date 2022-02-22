package model

import "time"

type Sensor struct {
	Id          int64   `gorm:"primaryKey" json:"id"`
	Flight      bool    `gorm:"type:boolean" json:"flight"`
	Temperature float64 `gorm:"type:numeric(4,2);" json:"temperature"`
	Humidity    float64 `gorm:"type:numeric(4,2);" json:"humidity"`
	CreatedAt   time.Time 
}
