package main

import (
	"database/sql"
	"log"
	"net/http"

	"encoding/json"

	"github.com/go-chi/chi"
	_ "github.com/go-sql-driver/mysql"
)

type Route struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Source      JSON   `json:"source"`
	Destination JSON   `json:"destination"`
}

type JSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func main() {
	r := chi.NewRouter()
	db, err := sql.Open("mysql", "root:password@tcp(mysql:3306)/mydb")
	if err != nil {
		log.Println("err: ", err)
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS routes (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			source_lat FLOAT,
			source_lng FLOAT,
			dest_lat FLOAT,
			dest_lng FLOAT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	initRoutes(db, r)

	log.Println("Server is running on :8080")
	http.ListenAndServe(":8080", r)
}

func initRoutes(db *sql.DB, r chi.Router) {
	r.Post("/api/routes", createRouteHandler(db))
	r.Get("/api/routes", listRoutesHandler(db))
}

func listRoutesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, source_lat, source_lng, dest_lat, dest_lng FROM routes")
		if err != nil {
			log.Println("err: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var routes []Route
		for rows.Next() {
			var route Route
			err := rows.Scan(&route.ID, &route.Name, &route.Source.Lat, &route.Source.Lng, &route.Destination.Lat, &route.Destination.Lng)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			routes = append(routes, route)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(routes)
	}
}

func createRouteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var route Route
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO routes (name, source_lat, source_lng, dest_lat, dest_lng) VALUES (?, ?, ?, ?, ?)",
			route.Name, route.Source.Lat, route.Source.Lng, route.Destination.Lat, route.Destination.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		route.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(route)
	}
}
