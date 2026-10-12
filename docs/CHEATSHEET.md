# Карта проекта и шпаргалка

Цепочка: **поле Angular → объект TypeScript → HttpClient → JSON → обработчик Go → SQL с `?` → JSON → страница**. Формы и SQL уже связаны; для нового задания меняются поля и правила по цепочке.

## Где менять

| Задача | Файл |
| --- | --- |
| Реквизиты MySQL, порт, администратор | `backend/.env` (создаётся из `.env.example`, в Git не попадает) |
| Таблицы, индексы, связи | `backend/schema.sql` |
| Создание таблиц и администратора | `backend/db.go` |
| Маршруты API и запуск | `backend/main.go` |
| Регистрация, сессии и роли | `backend/auth.go` |
| SQL и операции с заявками | `backend/requests.go` |
| Ограничения полей | `backend/validation.go` |
| Ответы JSON, защита Origin, раздача Angular | `backend/web.go` |
| Название, помещения, подписи статусов/оплаты | `frontend/src/app/config.ts` |
| Цвета, отступы, адаптивность | `frontend/src/styles.scss` |
| Типы JSON | `frontend/src/app/models.ts` |
| API заявок | `frontend/src/app/api.service.ts` |
| Вход/выход, текущий пользователь | `frontend/src/app/auth.service.ts` |
| Страницы и проверки переходов | `frontend/src/app/app.routes.ts`, `guards.ts` |
| Форма регистрации | `frontend/src/app/pages/register.ts`, `register.html` |
| Форма входа | `frontend/src/app/pages/login.ts`, `login.html` |
| Список, удаление, отзывы | `frontend/src/app/pages/requests.ts`, `requests.html` |
| Создание и редактирование заявки | `frontend/src/app/pages/request-form.ts`, `request-form.html` |
| Админка | `frontend/src/app/pages/admin.ts`, `admin.html` |
| Адрес Go для разработки | `frontend/proxy.conf.json` |

`schema.sql` встроен в Go через `go:embed`: после изменения SQL пересоберите сервер. Команда `migrate` не выполняет автоматическое сравнение схем: существующим таблицам нужен явный `ALTER TABLE`.

## Контракт API

Все пути начинаются с `/api`. Тела запросов и ответы — JSON. Пользовательские ошибки возвращаются как `{"error":"Сообщение"}`. Для обращения к защищённым маршрутам нужна cookie, полученная при регистрации или входе. Роль и владелец проверяются в Go.

| Метод и путь | Вход | Результат |
| --- | --- | --- |
| `GET /api/health` | — | `{"status":"ok"}` при доступной БД |
| `POST /api/auth/register` | `login`, `password`, `fullName`, `phone`, `email` | Пользователь и cookie сессии |
| `POST /api/auth/login` | `login`, `password` | Пользователь и cookie сессии |
| `GET /api/auth/me` | Cookie | Текущий пользователь |
| `POST /api/auth/logout` | Cookie | Завершение сессии |
| `GET /api/requests` | Cookie | Массив собственных заявок; пустой — `[]` |
| `POST /api/requests` | `room`, `startsAt`, `payment` | Созданная заявка |
| `PUT /api/requests/{id}` | Те же три поля | Изменённая собственная заявка |
| `DELETE /api/requests/{id}` | — | Удаление собственной заявки |
| `POST /api/requests/{id}/review` | `review` | Заявка с отзывом; статус должен быть `done` |
| `GET /api/admin/requests?status=new` | Необязательный фильтр | Все заявки; нужна роль `admin` |
| `PATCH /api/admin/requests/{id}/status` | `status` | Заявка с новым статусом; нужна роль `admin` |

Пример регистрации:

```json
{
  "login": "student01",
  "password": "practice123",
  "fullName": "Иван Иванов",
  "phone": "8(999)123-45-67",
  "email": "ivan@example.com"
}
```

Пример создания/редактирования:

```json
{"room":"Коворкинг","startsAt":"2026-10-20","payment":"cash"}
```

Заявка в ответе:

```json
{
  "id": 1,
  "userId": 2,
  "fullName": "Иван Иванов",
  "room": "Коворкинг",
  "startsAt": "2026-10-20",
  "payment": "cash",
  "status": "new",
  "review": ""
}
```

`payment`: `cash` или `sbp`. `status`: `new`, `scheduled`, `done`. Дата передаётся строкой `YYYY-MM-DD`; часовой пояс браузера не должен сдвигать её на сутки. `userId`, `fullName` и `status` определяет сервер; при создании заявки их не отправляем.

## Добавление поля: пример «Комментарий»

1. В новой схеме добавьте `comment TEXT NOT NULL` в `requests`. Для уже созданной таблицы выполните подходящий `ALTER TABLE` с учётом существующих записей.
2. Добавьте поле в `requestInput` и серверную проверку в `validation.go`.
3. Добавьте поле в структуру ответа, SQL `INSERT`, `UPDATE`, `SELECT` в `requests.go`. MySQL использует `?`, значения передавайте отдельными аргументами.
4. Добавьте `comment: string` в `BookingInput` в `models.ts`.
5. Добавьте исходное значение в `request-form.ts`, перенос поля при редактировании и элемент с `name="comment" [(ngModel)]="data.comment"` в HTML.
6. Выведите комментарий в списке, добавьте нужный `data-testid` из выданной разметки, проверьте создание и редактирование после обновления страницы.

## Частые конструкции Angular

```html
<!-- Текст из свойства -->
<h1>{{ title }}</h1>

<!-- Двусторонняя связь; FormsModule указан в imports компонента -->
<input name="room" [(ngModel)]="data.room">

<!-- Событие вызывает метод, нужны скобки -->
<button type="button" (click)="reload()">Обновить</button>

<!-- Условие -->
@if (error) { <p role="alert">{{ error }}</p> }

<!-- Список отслеживается по ID записи -->
@for (booking of bookings; track booking.id) {
  <p>{{ booking.room }}</p>
} @empty {
  <p>Заявок пока нет</p>
}
```

Декоратор `@Component(...)` должен стоять непосредственно перед классом. Интерфейсы пишутся выше декоратора или в `models.ts`. После изменения TypeScript страницу обновляет dev-сервер; после изменения Go сервер нужно перезапустить.

## Проверка HTTP без Angular

В Linux/Bash, с работающим сервером:

```bash
curl -i -c /tmp/demoex-cookies.txt \
  -H 'Content-Type: application/json' \
  -d '{"login":"student01","password":"practice123"}' \
  http://127.0.0.1:8080/api/auth/login

curl -i -b /tmp/demoex-cookies.txt \
  -H 'Content-Type: application/json' \
  -d '{"room":"Коворкинг","startsAt":"2026-10-20","payment":"cash"}' \
  http://127.0.0.1:8080/api/requests

curl -i -b /tmp/demoex-cookies.txt http://127.0.0.1:8080/api/requests
```

Пользователь `student01` должен быть сначала зарегистрирован. Файл cookie содержит локальную сессию; не добавляйте его в сдаваемый репозиторий. В браузере запросы видны в DevTools → Network: сначала смотрите HTTP-статус и тело ответа, затем логи Go.

## Границы каркаса

`GET /orders` и импорт интернет-магазина относятся к другому заданию и в каркас не входят. Создайте **отдельную БД**, например `shop_db`, перед импортом выданного SQL. Дамп может содержать `DROP TABLE users`; не импортируйте его в `demoex`.

В `docs/schema.drawio` есть редактируемая схема таблиц шаблона. Сверьте её с фактической БД и экспортируйте диаграмму из СУБД по условиям экзамена. Шаблон не содержит выданных фотографий, шрифтов и исходников графического редактора. Их нужно подготовить по фактическим материалам. Локальные изображения и шрифты складывайте в `frontend/public/assets/` и подключайте по URL `/assets/имя-файла`. Все внешние ресурсы для офлайн-работы подключаются локально.
