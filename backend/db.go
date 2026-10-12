package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql
var schema string

func migrate(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// В этом файле только простые CREATE TABLE; multiStatements не требуется.
	for _, statement := range strings.Split(schema, ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("создание таблиц: %w", err)
		}
	}
	log.Print("Таблицы users, sessions, requests готовы")
	return nil
}

func seedAdmin(db *sqlx.DB) error {
	login := env("ADMIN_LOGIN", "admin01")
	password := os.Getenv("ADMIN_PASSWORD")
	name := env("ADMIN_NAME", "Администратор")
	if strings.TrimSpace(login) == "" || len(login) > 64 || strings.TrimSpace(name) == "" {
		return errors.New("заполните ADMIN_LOGIN и ADMIN_NAME")
	}
	// Пароль администратора задаётся заданием и может отличаться от правил регистрации.
	if strings.TrimSpace(password) == "" || len(password) > 72 {
		return errors.New("задайте ADMIN_PASSWORD в .env: от 1 до 72 байт")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var existingID int64
	err := db.GetContext(ctx, &existingID, "SELECT id FROM users WHERE login = ?", login)
	if err == nil {
		return errors.New("логин уже существует; seed-admin не перезаписывает пользователя")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("сначала выполните migrate: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		"INSERT INTO users(login, password_hash, full_name, phone, email, role) VALUES(?, ?, ?, '', '', 'admin')",
		login, hash, name)
	if err != nil {
		return err
	}
	log.Print("Администратор создан: ", login)
	return nil
}
