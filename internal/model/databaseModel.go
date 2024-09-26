package model

import (
	"gorm.io/gorm"
)

type Admin struct {
	ID        uint `gorm:"primaryKey"`
	Username  string
	Password  string
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	gorm.Model
	Name           string `gorm:"column:name;type:varchar(255)" validate:"required" json:"name"`
	Email          string `gorm:"column:email;type:varchar(255);unique" validate:"required,email" json:"email"`
	PhoneNumber    string `gorm:"column:phone_number;type:varchar(255);unique" validate:"required,e164" json:"phone_number"`
	Picture        string `gorm:"column:picture;type:text" json:"picture"`
	Blocked        bool   `gorm:"column:blocked;type:bool" json:"blocked"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255)" validate:"required" json:"hashed_password"`
}
