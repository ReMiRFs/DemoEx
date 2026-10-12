package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Request struct {
	ID       int64  `db:"id" json:"id"`
	UserID   int64  `db:"user_id" json:"userId"`
	FullName string `db:"full_name" json:"fullName"`
	Room     string `db:"room" json:"room"`
	StartsAt string `db:"starts_at" json:"startsAt"`
	Payment  string `db:"payment" json:"payment"`
	Status   string `db:"status" json:"status"`
	Review   string `db:"review" json:"review"`
}

const selectRequests = `SELECT r.id, r.user_id, u.full_name, r.room,
	DATE_FORMAT(r.starts_at, '%Y-%m-%d') AS starts_at,
	r.payment, r.status, COALESCE(r.review, '') AS review
	FROM requests r JOIN users u ON u.id = r.user_id`

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, http.StatusBadRequest, "Некорректный ID")
		return 0, false
	}
	return id, true
}

func (a *app) listRequests(w http.ResponseWriter, r *http.Request, user User) {
	items := make([]Request, 0) // JSON должен быть [], а не null.
	if err := a.db.SelectContext(r.Context(), &items, selectRequests+" WHERE r.user_id = ? ORDER BY r.id DESC", user.ID); err != nil {
		serverError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func (a *app) listAdminRequests(w http.ResponseWriter, r *http.Request, user User) {
	items := make([]Request, 0)
	query := selectRequests
	args := make([]any, 0)
	if status := r.URL.Query().Get("status"); status != "" {
		if !validStatus(status) {
			fail(w, http.StatusBadRequest, "Неизвестный статус")
			return
		}
		query += " WHERE r.status = ?"
		args = append(args, status)
	}
	if err := a.db.SelectContext(r.Context(), &items, query+" ORDER BY r.id DESC", args...); err != nil {
		serverError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

func (a *app) createRequest(w http.ResponseWriter, r *http.Request, user User) {
	var in requestInput
	if !readJSON(w, r, &in) {
		return
	}
	if err := in.validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	// user_id всегда берётся из сессии, а не из JSON браузера.
	result, err := a.db.ExecContext(r.Context(),
		"INSERT INTO requests(user_id, room, starts_at, payment) VALUES(?, ?, ?, ?)",
		user.ID, in.Room, in.StartsAt, in.Payment)
	if err != nil {
		serverError(w, err)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		serverError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, Request{
		ID: id, UserID: user.ID, FullName: user.FullName, Room: in.Room,
		StartsAt: in.StartsAt, Payment: in.Payment, Status: "new", Review: "",
	})
}

func (a *app) ownRequest(w http.ResponseWriter, r *http.Request, user User, id int64) (Request, bool) {
	var item Request
	err := a.db.GetContext(r.Context(), &item, selectRequests+" WHERE r.id = ? AND r.user_id = ?", id, user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusNotFound, "Заявка не найдена")
		return item, false
	}
	if err != nil {
		serverError(w, err)
		return item, false
	}
	return item, true
}

func (a *app) updateRequest(w http.ResponseWriter, r *http.Request, user User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in requestInput
	if !readJSON(w, r, &in) {
		return
	}
	if err := in.validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, ok := a.ownRequest(w, r, user, id); !ok {
		return
	}
	_, err := a.db.ExecContext(r.Context(),
		"UPDATE requests SET room = ?, starts_at = ?, payment = ? WHERE id = ? AND user_id = ?",
		in.Room, in.StartsAt, in.Payment, id, user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	// В MySQL unchanged UPDATE возвращает 0 affected rows: это не означает 404.
	if item, ok := a.ownRequest(w, r, user, id); ok {
		jsonResponse(w, http.StatusOK, item)
	}
}

func (a *app) deleteRequest(w http.ResponseWriter, r *http.Request, user User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	result, err := a.db.ExecContext(r.Context(), "DELETE FROM requests WHERE id = ? AND user_id = ?", id, user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		serverError(w, err)
		return
	}
	if count == 0 {
		fail(w, http.StatusNotFound, "Заявка не найдена")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) reviewRequest(w http.ResponseWriter, r *http.Request, user User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Review string `json:"review"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	in.Review = strings.TrimSpace(in.Review)
	if in.Review == "" || utf8.RuneCountInString(in.Review) > 2000 {
		fail(w, http.StatusBadRequest, "Отзыв должен содержать от 1 до 2000 символов")
		return
	}
	item, ok := a.ownRequest(w, r, user, id)
	if !ok {
		return
	}
	if item.Status != "done" {
		fail(w, http.StatusConflict, "Отзыв доступен после завершения мероприятия")
		return
	}
	// Условие статуса проверяется и в UPDATE, защищая от одновременной смены статуса.
	_, err := a.db.ExecContext(r.Context(),
		"UPDATE requests SET review = ? WHERE id = ? AND user_id = ? AND status = 'done'", in.Review, id, user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	item, ok = a.ownRequest(w, r, user, id)
	if !ok {
		return
	}
	if item.Status != "done" {
		fail(w, http.StatusConflict, "Статус заявки изменился. Обновите список")
		return
	}
	jsonResponse(w, http.StatusOK, item)
}

func (a *app) changeStatus(w http.ResponseWriter, r *http.Request, user User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	if !validStatus(in.Status) {
		fail(w, http.StatusBadRequest, "Неизвестный статус")
		return
	}
	if _, err := a.db.ExecContext(r.Context(), "UPDATE requests SET status = ? WHERE id = ?", in.Status, id); err != nil {
		serverError(w, err)
		return
	}
	var item Request
	err := a.db.GetContext(r.Context(), &item, selectRequests+" WHERE r.id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusNotFound, "Заявка не найдена")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, item)
}
