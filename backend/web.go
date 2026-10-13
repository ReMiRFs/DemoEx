package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Print("Запись ответа: ", err)
	}
}

func fail(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func serverError(w http.ResponseWriter, err error) {
	log.Print("Ошибка БД: ", err)
	fail(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
}

func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		fail(w, http.StatusBadRequest, "Некорректный JSON или неизвестное поле")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		fail(w, http.StatusBadRequest, "Ожидается один объект JSON")
		return false
	}
	return true
}

// Проверка Origin защищает cookie-сессию от запросов с чужого сайта.
// Angular через прокси работает с тем же origin; curl без Origin разрешён.
func (a *app) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
			if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
				if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
					fail(w, http.StatusForbidden, "Запрос с другого сайта запрещён")
					return
				}
				origin := r.Header.Get("Origin")
				if origin != "" && !a.allowedOrigins[origin] {
					scheme := "http"
					if r.TLS != nil {
						scheme = "https"
					}
					u, err := url.Parse(origin)
					if err != nil || u.Scheme != scheme || u.Host != r.Host || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
						fail(w, http.StatusForbidden, "Источник запроса не разрешён")
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.db.PingContext(ctx); err != nil {
		fail(w, http.StatusServiceUnavailable, "База данных недоступна")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
