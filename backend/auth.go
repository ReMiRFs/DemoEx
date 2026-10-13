package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookie = "session"
const sessionDuration = 12 * time.Hour

type User struct {
	ID       int64  `db:"id" json:"id"`
	Login    string `db:"login" json:"login"`
	FullName string `db:"full_name" json:"fullName"`
	Role     string `db:"role" json:"role"`
}

func sessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (a *app) setSession(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: token, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: a.secureCookie,
		Expires: expires, MaxAge: int(sessionDuration.Seconds()),
	})
}

func (a *app) auth(adminOnly bool, next func(http.ResponseWriter, *http.Request, User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil || len(cookie.Value) != 64 {
			fail(w, http.StatusUnauthorized, "Войдите в аккаунт")
			return
		}
		var user User
		err = a.db.GetContext(r.Context(), &user, `
			SELECT u.id, u.login, u.full_name, u.role FROM users u
			JOIN sessions s ON s.user_id = u.id
			WHERE s.token_hash = ? AND s.expires_at > UTC_TIMESTAMP()`, tokenHash(cookie.Value))
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, http.StatusUnauthorized, "Сессия истекла. Войдите снова")
			return
		}
		if err != nil {
			serverError(w, err)
			return
		}
		if adminOnly && user.Role != "admin" {
			fail(w, http.StatusForbidden, "Доступ только для администратора")
			return
		}
		next(w, r, user)
	}
}

func (a *app) register(w http.ResponseWriter, r *http.Request) {
	var in registration
	if !readJSON(w, r, &in) {
		return
	}
	if err := in.validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		serverError(w, err)
		return
	}
	token, err := sessionToken()
	if err != nil {
		serverError(w, err)
		return
	}
	tx, err := a.db.BeginTxx(r.Context(), nil)
	if err != nil {
		serverError(w, err)
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO users(login, password_hash, full_name, phone, email)
		VALUES(?, ?, ?, ?, ?)`, in.Login, hash, in.FullName, in.Phone, in.Email)
	if err != nil {
		var dbError *mysql.MySQLError
		if errors.As(err, &dbError) && dbError.Number == 1062 {
			fail(w, http.StatusConflict, "Логин уже занят")
		} else {
			serverError(w, err)
		}
		return
	}
	userID, err := result.LastInsertId()
	if err != nil {
		serverError(w, err)
		return
	}
	expires := time.Now().UTC().Add(sessionDuration)
	_, err = tx.ExecContext(r.Context(), "INSERT INTO sessions(token_hash, user_id, expires_at) VALUES(?, ?, ?)", tokenHash(token), userID, expires)
	if err != nil {
		serverError(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		serverError(w, err)
		return
	}
	a.setSession(w, token, expires)
	jsonResponse(w, http.StatusCreated, User{ID: userID, Login: in.Login, FullName: in.FullName, Role: "user"})
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	if in.Password == "" || len(in.Password) > 72 || len(in.Login) > 64 {
		fail(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}
	var account struct {
		User
		PasswordHash string `db:"password_hash"`
	}
	err := a.db.GetContext(r.Context(), &account,
		"SELECT id, login, full_name, role, password_hash FROM users WHERE login = ?", strings.TrimSpace(in.Login))
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(in.Password)); err != nil {
		fail(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}
	token, err := sessionToken()
	if err != nil {
		serverError(w, err)
		return
	}
	expires := time.Now().UTC().Add(sessionDuration)
	_, err = a.db.ExecContext(r.Context(), "INSERT INTO sessions(token_hash, user_id, expires_at) VALUES(?, ?, ?)", tokenHash(token), account.ID, expires)
	if err != nil {
		serverError(w, err)
		return
	}
	a.setSession(w, token, expires)
	jsonResponse(w, http.StatusOK, account.User)
}

func (a *app) me(w http.ResponseWriter, r *http.Request, user User) {
	jsonResponse(w, http.StatusOK, user)
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if _, err := a.db.ExecContext(r.Context(), "DELETE FROM sessions WHERE token_hash = ?", tokenHash(cookie.Value)); err != nil {
			serverError(w, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: a.secureCookie,
	})
	w.WriteHeader(http.StatusNoContent)
}
