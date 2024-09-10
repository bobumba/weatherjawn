package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "modernc.org/sqlite"
)

var sqlStatement string = `select temperature from airfeelings order by datetime desc LIMIT 1`
var db *sql.DB

type WeatherRecord struct {
	ID                 int
	DateTime           string
	Temperature        float64
	Humidity           float64
	BarometricPressure float64
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func CurrentTempHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	s := fmt.Sprintf("%f", CurrentTemp(db))
	w.Write([]byte(s))
}
func CurrentRecordHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	wr := CurrentRecord(db)
	err := json.NewEncoder(w).Encode(wr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
func AddRecordDB(wr WeatherRecord) (bool, error) {
	update, err := db.Begin()
	if err != nil {
		return false, err
	}
	statement, err := update.Prepare("INSERT INTO airfeelings (datetime, temperature, humidity, barometricpressure) VALUES (?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	defer statement.Close()
	_, err = statement.Exec(wr.DateTime, wr.Temperature, wr.Humidity, wr.BarometricPressure)
	if err != nil {
		return false, err
	}
	update.Commit()
	return true, nil
}
func AddRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record WeatherRecord
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		log.Fatalf("fuckthis: %v", err)
	}
	_, err = AddRecordDB(record)
	if err != nil {
		log.Fatalf("fuckthis2: %v", err)
	}
	w.Write([]byte("ok"))
}
func CurrentTemp(db *sql.DB) float64 {
	var curTempF float64
	sqlStatement := `select temperature from airfeelings order by id desc LIMIT 1`
	err := db.QueryRow(sqlStatement).Scan(&curTempF)
	if err != nil {
		log.Fatalf("Error executing SQL statement: %v", err)
	}
	return curTempF
}
func CurrentRecord(db *sql.DB) WeatherRecord {
	var curWR WeatherRecord
	sqlStatement := `select * from airfeelings order by id desc LIMIT 1`
	err := db.QueryRow(sqlStatement).Scan(&curWR.ID, &curWR.DateTime, &curWR.Temperature, &curWR.Humidity, &curWR.BarometricPressure)
	if err != nil {
		log.Fatalf("Error executing SQL statement: %v", err)
	}
	return curWR
}
func main() {
	// Example SQL statement. Replace this with the actual statement you want to execute.

	// Open the SQLite database file.
	var err error
	db, err = sql.Open("sqlite", "weatherjawn.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Execute the SQL statement.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//#r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//#	w.Write([]byte("root."))
	//#})
	r.Get("/", CurrentTempHandler)
	r.Get("/cur", CurrentRecordHandler)
	r.Post("/", AddRecordHandler)
	http.ListenAndServe(":51101", r)
}
