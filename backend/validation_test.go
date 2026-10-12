package main

import (
	"strings"
	"testing"
)

func TestRegistrationValidation(t *testing.T) {
	valid := registration{Login: "student01", Password: "password123", FullName: "Иван Иванов", Phone: "8(999)123-45-67", Email: "test@example.com"}
	cases := []struct {
		name   string
		change func(*registration)
	}{
		{"короткий логин", func(in *registration) { in.Login = "short" }},
		{"кириллица в логине", func(in *registration) { in.Login = "студент" }},
		{"короткий пароль", func(in *registration) { in.Password = "1234567" }},
		{"пароль сверх лимита bcrypt", func(in *registration) { in.Password = strings.Repeat("Я", 37) }},
		{"имя пробелами", func(in *registration) { in.FullName = "   " }},
		{"латиница в имени", func(in *registration) { in.FullName = "Ivan" }},
		{"неверный формат телефона", func(in *registration) { in.Phone = "+79991234567" }},
		{"имя в email", func(in *registration) { in.Email = "Test <test@example.com>" }},
		{"нет домена email", func(in *registration) { in.Email = "test@" }},
	}
	if err := valid.validate(); err != nil {
		t.Fatalf("корректные данные отклонены: %v", err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid
			tc.change(&in)
			if err := in.validate(); err == nil {
				t.Fatal("ожидалась ошибка")
			}
		})
	}
}

func TestRequestValidation(t *testing.T) {
	cases := []struct {
		name  string
		input requestInput
		valid bool
	}{
		{"нормальные данные", requestInput{"Коворкинг", "2026-10-20", "cash"}, true},
		{"високосная дата", requestInput{"Зал", "2028-02-29", "sbp"}, true},
		{"несуществующая дата", requestInput{"Зал", "2026-02-29", "cash"}, false},
		{"неверный формат", requestInput{"Зал", "20.10.2026", "cash"}, false},
		{"дата вне диапазона MySQL", requestInput{"Зал", "0001-01-01", "cash"}, false},
		{"пустое помещение", requestInput{"  ", "2026-10-20", "cash"}, false},
		{"неизвестная оплата", requestInput{"Зал", "2026-10-20", "card"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.validate()
			if (err == nil) != tc.valid {
				t.Fatalf("ожидался valid=%v, получено %v", tc.valid, err)
			}
		})
	}
}
