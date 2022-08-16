package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Token     string `gorm:"unique; not null"`
	Username  string `gorm:"unique; not null"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	ImageURL  string
	PublicID  string `gorm:"unique; not null"`
}

type PublicUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageURL  string `json:"image_url"`
	Token     string `json:"token,omitempty"`
}

func (s *ServerConfig) HandleUser(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	switch r.Method {
	case "GET":
		s.GetUser(w, r)
	case "POST":
		s.CreateUser(w, r)
	case "PUT":
		s.UpdateUser(w, r)
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *ServerConfig) GetUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	username := strings.TrimSpace(r.URL.Query().Get("username"))

	if username == "" && id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &User{}
	tx := s.DB.First(user, User{Username: username, PublicID: id})
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	publicUser := &PublicUser{
		ID:        user.PublicID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ImageURL:  user.ImageURL,
	}

	err := json.NewEncoder(w).Encode(publicUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ServerConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	parameters := struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		ImageURL  string `json:"image_url"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parameters.Username = strings.TrimSpace(parameters.Username)
	parameters.FirstName = strings.TrimSpace(parameters.FirstName)
	parameters.LastName = strings.TrimSpace(parameters.LastName)
	parameters.ImageURL = strings.TrimSpace(parameters.ImageURL)

	if anyEmpty(
		parameters.Username,
		parameters.FirstName,
		parameters.LastName,
	) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &User{}

	tx := s.DB.First(user, User{Username: parameters.Username})
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Username = parameters.Username
	user.FirstName = parameters.FirstName
	user.LastName = parameters.LastName
	user.ImageURL = parameters.ImageURL
	user.Token = generateID()
	user.PublicID = generateID()

	tx = s.DB.Create(user)
	if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	publicUser := &PublicUser{
		ID:        user.PublicID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ImageURL:  user.ImageURL,
		Token:     user.Token, // Temporary for ease of use
	}

	err = json.NewEncoder(w).Encode(publicUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ServerConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	parameters := struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		ImageURL  string `json:"image_url"`
		Token     string `json:"token"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parameters.Token = strings.TrimSpace(parameters.Token)
	parameters.ID = strings.TrimSpace(parameters.ID)
	parameters.Username = strings.TrimSpace(parameters.Username)
	parameters.FirstName = strings.TrimSpace(parameters.FirstName)
	parameters.LastName = strings.TrimSpace(parameters.LastName)
	parameters.ImageURL = strings.TrimSpace(parameters.ImageURL)

	if parameters.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &User{}
	tx := s.DB.First(user, &User{PublicID: parameters.ID})
	if tx.Error != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if parameters.Token == "" || user.Token != parameters.Token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	changed := parameters.Username != user.Username ||
		parameters.FirstName != user.FirstName ||
		parameters.LastName != user.LastName ||
		parameters.ImageURL != user.ImageURL

	if changed {
		if parameters.Username != "" {
			user.Username = parameters.Username
		}

		if parameters.FirstName != "" {
			user.FirstName = parameters.FirstName
		}

		if parameters.LastName != "" {
			user.LastName = parameters.LastName
		}

		if parameters.ImageURL != "" {
			user.ImageURL = parameters.ImageURL
		}

		tx = s.DB.Save(user)
		if tx.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	publicUser := &PublicUser{
		ID:        user.PublicID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ImageURL:  user.ImageURL,
	}

	err = json.NewEncoder(w).Encode(publicUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
