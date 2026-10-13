package main

import (
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Booking struct {
	ID       int64  `json:"id" db:"id"`
	FullName string `json:"fullName" db:"full_name"`
	Room     string `json:"room" db:"room"`
	StartsAt string `json:"startsAt" db:"starts_at"`
}

func main() {
	// Подставь свои настройки MySQL.
	config := mysql.NewConfig()
	config.User = "YOUR_MYSQL_USER"
	config.Passwd = "YOUR_MYSQL_PASSWORD"
	config.Net = "tcp"
	config.Addr = "127.0.0.1:3306"
	config.DBName = "demo_simple"
	config.ParseTime = true
	db, err := sqlx.Connect("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("GET /api/requests", func(w http.ResponseWriter, r *http.Request) {
		bookings := []Booking{}
		err := db.Select(&bookings, "SELECT id, full_name, room, DATE_FORMAT(starts_at, '%Y-%m-%d') AS starts_at FROM bookings ORDER BY id DESC")
		if err != nil {
			serverError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bookings)
	})
	http.HandleFunc("POST /api/requests", func(w http.ResponseWriter, r *http.Request) {
		var booking Booking
		if !readBooking(w, r, &booking) {
			return
		}
		result, err := db.Exec("INSERT INTO bookings (full_name, room, starts_at) VALUES (?, ?, ?)", booking.FullName, booking.Room, booking.StartsAt)
		if err != nil {
			serverError(w, err)
			return
		}
		booking.ID, err = result.LastInsertId()
		if err != nil {
			serverError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(booking)
	})
	http.HandleFunc("PUT /api/requests/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := readID(w, r)
		if !ok {
			return
		}
		var booking Booking
		if !readBooking(w, r, &booking) {
			return
		}
		_, err := db.Exec("UPDATE bookings SET full_name = ?, room = ?, starts_at = ? WHERE id = ?", booking.FullName, booking.Room, booking.StartsAt, id)
		if err != nil {
			serverError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	http.HandleFunc("DELETE /api/requests/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, ok := readID(w, r)
		if !ok {
			return
		}
		_, err := db.Exec("DELETE FROM bookings WHERE id = ?", id)
		if err != nil {
			serverError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	log.Println("Go: http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}

func readBooking(w http.ResponseWriter, r *http.Request, booking *Booking) bool {
	if err := json.NewDecoder(r.Body).Decode(booking); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return false
	}
	if strings.TrimSpace(booking.FullName) == "" || strings.TrimSpace(booking.Room) == "" || booking.StartsAt == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return false
	}
	return true
}

func readID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func serverError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
}
