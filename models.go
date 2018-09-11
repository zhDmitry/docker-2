package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Task struct {
	ID   string `gorm:"primary_key" json:"id"`
	Info string `gorm:"size:2048" json:"info"`

	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
}

type User struct {
	gorm.Model
	Username string `gorm:"size:2048"`
	Password string `gorm:"size:2048"`
}
