package main

import (
	"context"
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

type WeatherService struct {
	DB *sql.DB
}

func NewWeatherService(d *sql.DB) *WeatherService {
	return &WeatherService{DB: d}
}

func (a *WeatherService) AddRecord(ctx context.Context, record WeatherRecord) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query, err := tx.PrepareContext(ctx, "INSERT INTO airfeelings (datetime, temperature, humidity, barometricpressure) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer query.Close()

	_, err = query.ExecContext(ctx, record.DateTime, record.Temperature, record.Humidity, record.BarometricPressure)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (c *WeatherService) RetrieveRecord(ctx context.Context) (*WeatherRecord, error) {
	var curWR WeatherRecord
	query := `select * from airfeelings order by id desc LIMIT 1`
	err := db.QueryRowContext(ctx, query).Scan(&curWR.ID, &curWR.DateTime, &curWR.Temperature, &curWR.Humidity, &curWR.BarometricPressure)
	if err != nil {
		log.Fatalf("Error retrieving record: %v", err)
	}
	return &curWR, nil
}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (a *WeatherService) AddRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record WeatherRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "could not decode", http.StatusBadRequest)
		log.Printf("could not decode json: %v", err)
		return
	}
	if err := a.AddRecord(r.Context(), record); err != nil {
		http.Error(w, "error adding record", http.StatusBadRequest)
		log.Printf("error adding record: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "record added"})
}

func (a *WeatherService) RetrieveRecordHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cur, err := a.RetrieveRecord(r.Context())
	if err != nil {
		http.Error(w, "error retrieving record", http.StatusInternalServerError)
		log.Printf("eror with db retrieve: %v", err)
	}
	w.Header().Set("Content-type", "application/json")
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

	service := NewWeatherService(db)
	// Execute the SQL statement.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	//#r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//#	w.Write([]byte("root."))
	//#})
	r.Get("/", service.RetrieveRecordHandler)
	r.Post("/", service.AddRecordHandler)
	http.ListenAndServe(":51101", r)
}
