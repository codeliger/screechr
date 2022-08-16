package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type Screech struct {
	gorm.Model
	UserID   uint
	User     User
	PublicID string `gorm:"unique; not null"`
	Content  string `gorm:"size:1024; not null"`
}

type PublicScreech struct {
	ID       string `json:"id"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content"`
}

func (s *ServerConfig) HandleScreech(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	switch r.Method {
	case "GET":
		s.GetScreech(w, r)
	case "POST":
		s.CreateScreech(w, r)
	case "PUT":
		s.UpdateScreech(w, r)
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *ServerConfig) CreateScreech(w http.ResponseWriter, r *http.Request) {
	parameters := struct {
		Token   string `json:"token"`
		Content string `json:"content"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parameters.Token = strings.TrimSpace(parameters.Token)
	parameters.Content = strings.TrimSpace(parameters.Content)

	if anyEmpty(
		parameters.Token,
		parameters.Content) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := &User{}
	tx := s.DB.First(user, User{Token: parameters.Token})
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	screech := &Screech{
		UserID:   user.ID,
		PublicID: generateID(),
		Content:  parameters.Content,
	}

	tx = s.DB.Create(screech)
	if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	publicScreech := &PublicScreech{
		ID:      screech.PublicID,
		Content: screech.Content,
	}

	err = json.NewEncoder(w).Encode(publicScreech)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ServerConfig) UpdateScreech(w http.ResponseWriter, r *http.Request) {
	parameters := struct {
		Token    string `json:"token"`
		PublicID string `json:"id"`
		Content  string `json:"content"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	parameters.Token = strings.TrimSpace(parameters.Token)
	parameters.PublicID = strings.TrimSpace(parameters.PublicID)
	parameters.Content = strings.TrimSpace(parameters.Content)

	if anyEmpty(parameters.Token, parameters.PublicID, parameters.Content) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	screech := &Screech{}
	tx := s.DB.Preload("User").First(screech, &Screech{PublicID: parameters.PublicID})
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if screech.User.Token != parameters.Token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if screech.Content != parameters.Content {
		screech.Content = parameters.Content
		tx = s.DB.Save(screech)
		if tx.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	publicScreech := &PublicScreech{
		ID:      screech.PublicID,
		Content: screech.Content,
	}

	err = json.NewEncoder(w).Encode(publicScreech)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ServerConfig) GetScreech(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	screech := &Screech{}
	tx := s.DB.Preload("User").First(screech, &Screech{PublicID: id})
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if tx.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	publicScreech := &PublicScreech{
		ID:       screech.PublicID,
		Username: screech.User.Username,
		Content:  screech.Content,
	}

	err := json.NewEncoder(w).Encode(publicScreech)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ServerConfig) ListScreeches(w http.ResponseWriter, r *http.Request) {
	countRaw := strings.TrimSpace(r.URL.Query().Get("count"))
	usernameRaw := strings.TrimSpace(r.URL.Query().Get("username"))
	publicUserIDRaw := strings.TrimSpace(r.URL.Query().Get("user_id"))

	user := &User{}
	if usernameRaw != "" || publicUserIDRaw != "" {
		tx := s.DB.First(user, User{PublicID: publicUserIDRaw, Username: usernameRaw})
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else if tx.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	sortOrderRaw := r.URL.Query().Get("order")
	sortOrder := "desc"
	if strings.EqualFold(sortOrderRaw, "asc") {
		sortOrder = "asc"
	}

	count := parseCount(countRaw)

	screeches := []Screech{}

	orderBy := fmt.Sprintf("created_at %s", sortOrder)

	if user.Username == "" {
		s.DB.Preload("User").Find(&screeches).Order(orderBy).Limit(count)
	} else {
		s.DB.Preload("User").Find(&screeches).Where("username = ?", user.Username).Order(orderBy).Limit(count)
	}

	publicScreeches := []PublicScreech{}

	for _, screech := range screeches {
		publicScreech := PublicScreech{
			ID:       screech.PublicID,
			Content:  screech.Content,
			Username: screech.User.Username,
		}

		publicScreeches = append(publicScreeches, publicScreech)
	}

	err := json.NewEncoder(w).Encode(publicScreeches)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
