package main

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ServerConfig struct {
	DB *gorm.DB
}

// generateID generates a UUID without hyphens
func generateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// setHeaders Sets all headers for JSON endpoints
func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
}

// RegisterRoutes registers all routes for the server
func (s *ServerConfig) RegisterRoutes() {
	http.HandleFunc("/user", s.HandleUser)
	http.HandleFunc("/screech", s.HandleScreech)
	http.HandleFunc("/screeches", s.ListScreeches)
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	config := &ServerConfig{
		DB: db,
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		panic("failed to migrate schema")
	}
	err = db.AutoMigrate(&Screech{})
	if err != nil {
		panic("failed to migrate schema")
	}

	config.RegisterRoutes()

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("failed to start server")
	}
}
