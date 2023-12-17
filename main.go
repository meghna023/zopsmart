package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Item represents your entity structure.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if not exists
	createTable := `
		CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT
		);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var newItem Item
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into the database
	_, err = db.Exec("INSERT INTO items (id, name) VALUES (?, ?)", newItem.ID, newItem.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM items")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemID := params["id"]

	var updatedItem Item
	err := json.NewDecoder(r.Body).Decode(&updatedItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE items SET name = ? WHERE id = ?", updatedItem.Name, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	itemID := params["id"]

	_, err := db.Exec("DELETE FROM items WHERE id = ?", itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	initDB()

	r := mux.NewRouter()

	// CRUD Endpoints
	r.HandleFunc("/items", createItem).Methods("POST")
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items/{id}", updateItem).Methods("PUT")
	r.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")

	port := 8080
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
