package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type app struct {
	db             *sqlx.DB
	secureCookie   bool
	allowedOrigins map[string]bool
	staticDir      string
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Не удалось прочитать .env: ", err)
	}
	if len(os.Args) > 1 && os.Args[1] != "migrate" && os.Args[1] != "seed-admin" {
		log.Fatal("Команды: go run . | go run . migrate | go run . seed-admin")
	}

	// Config корректно обрабатывает специальные символы в пароле.
	config := mysql.NewConfig()
	config.User = env("DB_USER", "demo")
	config.Passwd = os.Getenv("DB_PASSWORD")
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(env("DB_HOST", "127.0.0.1"), env("DB_PORT", "3306"))
	config.DBName = env("DB_NAME", "demoex")
	config.ParseTime = true
	config.Loc = time.UTC
	config.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
	config.Timeout = 5 * time.Second
	config.ReadTimeout = 10 * time.Second
	config.WriteTimeout = 10 * time.Second
	db, err := sqlx.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		log.Fatal("Подключение к MySQL: ", err)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			err = migrate(db)
		case "seed-admin":
			err = seedAdmin(db)
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	secure, err := strconv.ParseBool(env("COOKIE_SECURE", "false"))
	if err != nil {
		log.Fatal("COOKIE_SECURE должен быть true или false")
	}
	a := &app{
		db: db, secureCookie: secure,
		allowedOrigins: make(map[string]bool),
		staticDir:      env("STATIC_DIR", "../frontend/dist/frontend/browser"),
	}
	for _, origin := range strings.Split(env("ALLOWED_ORIGINS", "http://localhost:4200,http://127.0.0.1:4200"), ",") {
		a.allowedOrigins[strings.TrimSpace(origin)] = true
	}
	server := &http.Server{
		Addr: env("ADDR", "127.0.0.1:8080"), Handler: a.routes(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	stop, stopCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopCancel()
	go func() {
		<-stop.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Print("Завершение сервера: ", err)
		}
	}()
	log.Printf("Сервер: http://%s; API: /api/health", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
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
	mux.HandleFunc("/", a.spa)
	return a.protect(mux)
}
