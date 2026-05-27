package models

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Login        string    `json:"login" gorm:"uniqueIndex;not null" binding:"required"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Password     string    `json:"password,omitempty" gorm:"-"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Terminal struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SerialNumber string    `json:"serial_number" gorm:"uniqueIndex;not null" binding:"required"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type CardKey struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null" binding:"required"`
	KeyValue    string    `json:"key_value" gorm:"not null" binding:"required"`
	Description string    `json:"description"`
	Cards       []Card    `json:"cards,omitempty" gorm:"foreignKey:KeyID" binding:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Card struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Number    string    `json:"number" gorm:"uniqueIndex;not null" binding:"required"`
	Balance   int64     `json:"balance"`
	Blocked   bool      `json:"blocked"`
	OwnerName string    `json:"owner_name"`
	UserID    uint      `json:"user_id" gorm:"index"`
	KeyID     uint      `json:"key_id"`
	Key       CardKey   `json:"key,omitempty" gorm:"foreignKey:KeyID" binding:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Transaction struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Amount     int64     `json:"amount" binding:"required"`
	CardID     uint      `json:"card_id"`
	Card       Card      `json:"card,omitempty" gorm:"foreignKey:CardID" binding:"-"`
	TerminalID uint      `json:"terminal_id"`
	Terminal   Terminal  `json:"terminal,omitempty" gorm:"foreignKey:TerminalID" binding:"-"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
