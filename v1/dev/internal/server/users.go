package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"slices"
	"sync"
)

var mutex sync.RWMutex

var (
	ErrSavData = errors.New("Error saving data")
	ErrReaData = errors.New("Error reading data")
	ErrFetData = errors.New("Error fetching data")
)

type Field struct {
	Name string `json:"Name"`
	Age  int    `json:"Age"`
}

type MockedDbData struct {
	Fields []Field
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	f, err := os.ReadFile(s.dbPath)
	if err != nil {
		s.logger.Error("failed to read db file", "err", err)
		http.Error(w, ErrFetData.Error(), http.StatusInternalServerError)
		return
	}

	var data MockedDbData
	if err := json.Unmarshal(f, &data); err != nil {
		s.logger.Error("failed to parse db file", "err", err)
		http.Error(w, ErrFetData.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	var body Field
	var dbData MockedDbData

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.ReadFile(s.dbPath)
	if err != nil {
		s.logger.Error("failed to read db file", "err", err)
		http.Error(w, ErrReaData.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(f, &dbData); err != nil {
		s.logger.Error("failed to parse db file", "err", err)
		http.Error(w, ErrReaData.Error(), http.StatusInternalServerError)
		return
	}

	dbData.Fields = append(dbData.Fields, body)

	updated, err := json.Marshal(dbData)
	if err != nil {
		s.logger.Error("failed to marshal db data", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(s.dbPath, updated, 0644); err != nil {
		s.logger.Error("failed to write db file", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var dbData MockedDbData

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name query parameter", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.ReadFile(s.dbPath)
	if err != nil {
		s.logger.Error("failed to read db file", "err", err)
		http.Error(w, ErrReaData.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(f, &dbData); err != nil {
		s.logger.Error("failed to parse db file", "err", err)
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}

	idx := slices.IndexFunc(dbData.Fields, func(f Field) bool {
		return f.Name == name
	})
	if idx == -1 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	returned := dbData.Fields[idx]
	dbData.Fields = slices.Delete(dbData.Fields, idx, idx+1)

	updated, err := json.Marshal(dbData)
	if err != nil {
		s.logger.Error("failed to marshal db data", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(s.dbPath, updated, 0644); err != nil {
		s.logger.Error("failed to write db file", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(returned)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var dbData MockedDbData
	var userData Field

	mutex.Lock()
	defer mutex.Unlock()

	id := r.URL.Query().Get("name")

	f, err := os.ReadFile(s.dbPath)
	if err != nil {
		s.logger.Error("Error while reading the db", "error", err)
		http.Error(w, ErrReaData.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(f, &dbData)
	if err != nil {
		s.logger.Error("Error while decoding the db", "error", err)
		http.Error(w, "Error while decoding the db", http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		s.logger.Error("Error while decoding the request", "error", err)
		http.Error(w, "Error while decoding the request", http.StatusBadRequest)
		return
	}

	idx := slices.IndexFunc(dbData.Fields, func(f Field) bool {
		return f.Name == id
	})
	if idx == -1 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	dbData.Fields[idx] = userData
	updated, err := json.Marshal(dbData)

	if err != nil {
		s.logger.Error("failed to marshal db data", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(s.dbPath, updated, 0644); err != nil {
		s.logger.Error("failed to write db file", "err", err)
		http.Error(w, ErrSavData.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userData)
}
