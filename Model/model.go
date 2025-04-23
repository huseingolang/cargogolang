package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string  `json:"email"`
	Password string  `json:"Password"`
	Orders   []Order `gorm:"foreignKey:UserID "`
}
type Order struct {
	gorm.Model

	Name          string `json:"name"`
	Description   string `json:"description"`
	CurrentStatus string `json:"current_status"`
	CurrentTime   time.Time
	Status        []Status `gorm:"foreignKey:OrderID "`
	Chakes        []Chake  `gorm:"foreignKey:OrderID "`
	UserID        uint     `gorm:"index"`
}
type Chake struct {
	gorm.Model

	Name          string `json:"name"`
	CurrentStatus string `json:"current_status"`
	OrderID       uint   `gorm:"index"`
}
type Status struct {
	gorm.Model
	CurrentStatus string `json:"status"`
	OrderID       uint   `gorm:"index"`
}
type Notification struct {
	gorm.Model
	UserID  uint   `gorm:"index"`
	Message string `json:"message"`
	IsSent  bool   `json:"isSent"`
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	return err == nil

}
