package main

import (
	"database/sql"
	"encoding/json"
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

func (a *WeatherRecord) AddRecord() (bool, error) {
	update, err := db.Begin()
	if err != nil {
		return false, err
	}
	statement, err := update.Prepare("INSERT INTO airfeelings (datetime, temperature, humidity, barometricpressure) VALUES (?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	defer statement.Close()
	_, err = statement.Exec(a.DateTime, a.Temperature, a.Humidity, a.BarometricPressure)
	if err != nil {
		return false, err
	}
	update.Commit()
	return true, nil
}

func (c *WeatherRecord) RetrieveRecord() (*WeatherRecord, error) {
	var curWR WeatherRecord
	sqlStatement := `select * from airfeelings order by id desc LIMIT 1`
	err := db.QueryRow(sqlStatement).Scan(&curWR.ID, &curWR.DateTime, &curWR.Temperature, &curWR.Humidity, &curWR.BarometricPressure)
	if err != nil {
		log.Fatalf("Error executing SQL statement: %v", err)
	}
	return &curWR, nil
}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func AddRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record WeatherRecord
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		log.Fatalf("fuckthis: %v", err)
	}
	_, err = record.AddRecord()
	if err != nil {
		log.Fatalf("fuckthis2: %v", err)
	}
	w.Write([]byte("ok"))
}

func RetrieveRecordHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	var wr WeatherRecord
	cur, err := wr.RetrieveRecord()
	if err != nil {
		log.Fatalf("eror with db retrieve: %v", err)
	}
	err = json.NewEncoder(w).Encode(cur)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func RetrieveRecord(db *sql.DB) WeatherRecord {
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
	r.Get("/", RetrieveRecordHandler)
	r.Post("/", AddRecordHandler)
	http.ListenAndServe(":51101", r)
}
