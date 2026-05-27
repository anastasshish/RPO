package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"transport-auth-server/backend/internal/models"
)

func Connect() (*gorm.DB, error) {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "./transport.db"
	}
	mig := os.Getenv("MIGRATIONS_DIR")
	if mig == "" {
		mig = "./migrations"
	}
	sqlDB, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(sqlDB, mig); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Println(err)
	}
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	seed(gdb)
	return gdb, nil
}
func seed(gdb *gorm.DB) {
	var c int64
	gdb.Model(&models.User{}).Count(&c)
	if c == 0 {
		h, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		gdb.Create(&models.User{Login: "admin", Name: "Administrator", PasswordHash: string(h), IsAdmin: true})
		h2, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
		gdb.Create(&models.User{Login: "user", Name: "Regular User", PasswordHash: string(h2), IsAdmin: false})
	}
	gdb.Model(&models.Terminal{}).Count(&c)
	if c == 0 {
		gdb.Create(&models.Terminal{SerialNumber: "TERM-001", Name: "Demo terminal", Address: "Test address"})
	}
	gdb.Model(&models.CardKey{}).Count(&c)
	if c == 0 {
		gdb.Create(&models.CardKey{Name: "Default key", KeyValue: "A0A1A2A3A4A5", Description: "Demo MIFARE key"})
	}
	gdb.Model(&models.Card{}).Count(&c)
	if c == 0 {
		var u models.User
		_ = gdb.Where("login = ?", "user").First(&u).Error
		gdb.Create(&models.Card{Number: "0000000001", Balance: 50000, Blocked: false, OwnerName: "Demo Passenger", KeyID: 1, UserID: u.ID})
	}
}
