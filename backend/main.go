package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Настройки: замени пользователя и пароль MySQL.
const dbUser = "YOUR_MYSQL_USER"
const dbPassword = "YOUR_MYSQL_PASSWORD"
const dbAddress = "127.0.0.1:3306"
const dbName = "demo_exam"
const adminLogin = "Conf2027"
const adminPassword = "Demo77"

type app struct {
	db             *sqlx.DB
	secureCookie   bool
	allowedOrigins map[string]bool
}

func main() {
	config := mysql.NewConfig()
	config.User = dbUser
	config.Passwd = dbPassword
	config.Net = "tcp"
	config.Addr = dbAddress
	config.DBName = dbName
	config.ParseTime = true
	config.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
	db, err := sqlx.Connect("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal("Подключение к MySQL: ", err)
	}
	defer db.Close()
	if err := createAdmin(db); err != nil {
		log.Fatal(err)
	}
	a := &app{db: db, allowedOrigins: map[string]bool{
		"http://localhost:4200": true, "http://127.0.0.1:4200": true,
	}}
	log.Println("Go: http://127.0.0.1:8080. Открой Angular: http://127.0.0.1:4200")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", a.routes()))
}

// При первом запуске создаётся администратор. Существующие записи не меняются.
func createAdmin(db *sqlx.DB) error {
	var role string
	err := db.Get(&role, "SELECT role FROM users WHERE login = ?", adminLogin)
	if err == nil {
		if role != "admin" {
			return errors.New("логин администратора уже занят пользователем")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (login, password_hash, full_name, phone, email, role) VALUES (?, ?, ?, '', '', 'admin')", adminLogin, hash, "Администратор")
	return err
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("POST /api/auth/register", a.register)
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("GET /api/auth/me", a.auth(false, a.me))
	mux.HandleFunc("POST /api/auth/logout", a.logout)
	mux.HandleFunc("GET /api/requests", a.auth(false, a.listRequests))
	mux.HandleFunc("POST /api/requests", a.auth(false, a.createRequest))
	mux.HandleFunc("PUT /api/requests/{id}", a.auth(false, a.updateRequest))
	mux.HandleFunc("DELETE /api/requests/{id}", a.auth(false, a.deleteRequest))
	mux.HandleFunc("POST /api/requests/{id}/review", a.auth(false, a.reviewRequest))
	mux.HandleFunc("GET /api/admin/requests", a.auth(true, a.listAdminRequests))
	mux.HandleFunc("PATCH /api/admin/requests/{id}/status", a.auth(true, a.changeStatus))
	// Неизвестный API никогда не должен возвращать index.html приложения.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		fail(w, http.StatusNotFound, "Маршрут API не найден")
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		fail(w, http.StatusNotFound, "Маршрут API не найден")
	})

	return a.protect(mux)
}
