package main

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Правила сосредоточены здесь: на экзамене их легко заменить под новый вариант.
var (
	loginPattern = regexp.MustCompile(`^[A-Za-z0-9]{6,64}$`)
	namePattern  = regexp.MustCompile(`^[А-Яа-яЁё ]+$`)
	phonePattern = regexp.MustCompile(`^8\([0-9]{3}\)[0-9]{3}-[0-9]{2}-[0-9]{2}$`)
)

type registration struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

func (in *registration) validate() error {
	in.Login = strings.TrimSpace(in.Login)
	in.FullName = strings.TrimSpace(in.FullName)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
	if !loginPattern.MatchString(in.Login) {
		return errors.New("Логин: 6–64 латинских буквы или цифры")
	}
	if utf8.RuneCountInString(in.Password) < 8 || len(in.Password) > 72 {
		return errors.New("Пароль: минимум 8 символов, максимум 72 байта")
	}
	if !namePattern.MatchString(in.FullName) || utf8.RuneCountInString(in.FullName) > 255 {
		return errors.New("ФИО: кириллица и пробелы, до 255 символов")
	}
	if !phonePattern.MatchString(in.Phone) {
		return errors.New("Телефон должен иметь вид 8(999)123-45-67")
	}
	address, err := mail.ParseAddress(in.Email)
	if err != nil || address.Address != in.Email || len(in.Email) > 255 {
		return errors.New("Введите корректный email")
	}
	return nil
}

type requestInput struct {
	Room     string `json:"room"`
	StartsAt string `json:"startsAt"`
	Payment  string `json:"payment"`
}

func (in *requestInput) validate() error {
	in.Room = strings.TrimSpace(in.Room)
	if in.Room == "" || utf8.RuneCountInString(in.Room) > 255 {
		return errors.New("Введите помещение: до 255 символов")
	}
	date, err := time.Parse("2006-01-02", in.StartsAt)
	if err != nil || date.Year() < 1000 || date.Year() > 9999 {
		return errors.New("Дата должна иметь вид ГГГГ-ММ-ДД")
	}
	if in.Payment != "cash" && in.Payment != "sbp" {
		return errors.New("Способ оплаты: cash или sbp")
	}
	return nil
}

func validStatus(status string) bool {
	return status == "new" || status == "scheduled" || status == "done"
}
