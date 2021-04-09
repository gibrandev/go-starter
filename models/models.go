package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string `gorm:"primaryKey;autoIncrement:false"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	Password  string `json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Token struct {
	ID           string `gorm:"primaryKey;autoIncrement:false"`
	Sub          string `gorm:"index"`
	Ip           string
	Iss          string
	LastAccessAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

func (u *Token) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}
