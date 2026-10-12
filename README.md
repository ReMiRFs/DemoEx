# Универсальный каркас: Go + MySQL + Angular + SCSS

Рабочая основа для подготовки к демоэкзамену: регистрация, вход, сессия в cookie, личные заявки с добавлением/изменением/удалением, отзыв после завершения и админка со сменой статуса. Go выполняет проверку прав и сохраняет данные в MySQL. Angular показывает страницы и вызывает API.

Каркас нужно адаптировать к выданному варианту: поля, ограничения, тексты, `data-testid`, дизайн и артефакты сдачи. [Карта кода и API](docs/CHEATSHEET.md) · [Проверка перед экзаменом](docs/EXAM-CHECKLIST.md).

## Быстрый путь на компьютере с Linux

Нужны Go **1.22+**, MySQL **8.0+**, Node **22.12+ в ветке 22.x** или **24.x**, npm, Git. Точные версии зависимостей проекта закреплены в `go.mod`, `go.sum`, `package.json` и `package-lock.json`. Не выполняйте обновление всех зависимостей перед экзаменом.

### 1. ПО и загрузка проекта

Проверьте систему и уже установленное ПО:

```bash
cat /etc/os-release
go version
node --version
npm --version
git --version
mysql --version
```

Для Fedora с `dnf`, если чего-то не хватает:

```bash
sudo dnf install -y golang nodejs npm git curl zip unzip
sudo dnf install -y mysql-server
sudo systemctl enable --now mysqld
```

Команды рассчитаны на Fedora. У Rocky/Alma/RHEL/других систем с `dnf` могут отличаться пакеты и версии; не заменяйте уже настроенный сервер БД. Если площадка выдала общий MySQL, используйте его реквизиты и пропустите установку локального сервера. Версия Node из репозитория должна пройти проверку выше: Node 23/25/26 не подходит выбранному Angular 20.

Дополнительно, если отсутствуют браузер и редакторы для заданий:

```bash
sudo dnf install -y firefox gimp libreoffice
```

Редактор кода, графический исходник, ER-диаграмму и «Автотест» проверьте по требованиям площадки. DBeaver, если используете его, нужно заранее подключить к MySQL, чтобы JDBC-драйвер успел скачаться.

```bash
git clone https://github.com/ReMiRFs/DemoEx.git
cd DemoEx
bash scripts/setup-online.sh
```

Скрипт устанавливает зависимости через `npm ci`, складывает npm-кэш в `.offline/npm`, создаёт `backend/vendor` и проверяет сборку. Он не устанавливает системные пакеты и не меняет БД. Запускайте его **на целевом Linux-компьютере с интернетом**: `node_modules` из Windows содержит другие нативные пакеты.

### 2. База и конфигурация

Войдите в MySQL способом, который настроен на компьютере; например:

```bash
mysql -u root -p
```

Если на свежей локальной установке root использует socket-аутентификацию, вход может выполняться через `sudo mysql`. Не меняйте существующий пароль администратора ради шаблона. Создайте отдельную базу и пользователя (замените `your-local-password`):

```sql
CREATE DATABASE IF NOT EXISTS demoex CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'demo'@'localhost' IDENTIFIED BY 'your-local-password';
CREATE USER IF NOT EXISTS 'demo'@'127.0.0.1' IDENTIFIED BY 'your-local-password';
GRANT ALL PRIVILEGES ON demoex.* TO 'demo'@'localhost';
GRANT ALL PRIVILEGES ON demoex.* TO 'demo'@'127.0.0.1';
```

Если `demo` уже существует, `IF NOT EXISTS` сохраняет прежний пароль — используйте его либо создайте нового пользователя. Для удалённого сервера права и реквизиты выдаёт администратор площадки.

В корне репозитория:

```bash
cp backend/.env.example backend/.env
```

Откройте `backend/.env` в редакторе, заполните `DB_PASSWORD`, при необходимости `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`. Значения должны совпадать с созданной БД. Для локальной обычной установки порт — `3306`.

Заполните `ADMIN_LOGIN`, `ADMIN_PASSWORD`, `ADMIN_NAME` для администратора своего варианта. Эти данные не задаются в Angular. Пароли хешируются bcrypt. Сам файл `.env` исключён из Git.

```bash
cd backend
go run -mod=vendor . migrate
go run -mod=vendor . seed-admin
go run -mod=vendor .
```

`migrate` создаёт недостающие таблицы, сохраняя существующие данные. Изменение `schema.sql` не изменяет колонки в уже созданной БД: для этого нужен отдельный `ALTER TABLE`. `seed-admin` создаёт администратора и отказывается перезаписывать существующий логин.

### 3. Два терминала во время разработки

Терминал №1 — Go, из `DemoEx/backend`:

```bash
go run -mod=vendor .
```

Терминал №2 — Angular, из `DemoEx/frontend`:

```bash
npm start
```

Откройте **http://localhost:4200**. Прокси уже настроен: `/api/**` идёт к Go на `127.0.0.1:8080`. Работайте всё время через один адрес браузера: cookie для `localhost` и `127.0.0.1` различаются. При изменении прокси Angular перезапускается.

Проверьте из третьего терминала:

```bash
curl http://127.0.0.1:8080/api/health
curl http://localhost:4200/api/health
```

Затем зарегистрируйте пользователя, создайте заявку, обновите страницу и убедитесь, что она сохранена. Войдите администратором, измените статус на завершённый, снова войдите пользователем и оставьте отзыв.

### 4. Офлайн и один сервер для сдачи

Пока интернет есть, `setup-online.sh` уже должен отработать без ошибок. Отключите интернет и выполните из корня:

```bash
bash scripts/offline-check.sh
```

Скрипт собирает Go с `GOPROXY=off`, `GOTOOLCHAIN=local`, `-mod=vendor` и Angular из установленных зависимостей. Наличие `vendor` не заменяет установленный компилятор Go. Наличие `package-lock.json` не заменяет `node_modules`.

После сборки можно запускать только Go, который отдаёт готовый Angular:

```bash
cd backend
../bin/server
```

Откройте **http://127.0.0.1:8080**. MySQL должен продолжать работать. После изменения Angular повторите `npm run build` из `frontend`, чтобы обновить файлы на порту 8080. В режиме `npm start` изменения видны автоматически.

Сохраняйте на экзаменационном компьютере `backend/vendor`, `frontend/node_modules` и `.offline/npm`, если это разрешено площадкой. Они не отправляются в GitHub. Кэш предназначен для той ОС и архитектуры, на которой его подготовили; успешная сборка после отключения сети — обязательная практическая проверка. Не запускайте `npm ci` на уже готовом офлайн-окружении без необходимости: он пересоздаёт `node_modules`.

## Windows для домашней подготовки

Установите Go, Node и MySQL подходящих версий. В PowerShell из корня клонированного проекта:

```powershell
Copy-Item backend/.env.example backend/.env
Set-Location backend
go mod download
go mod vendor
go build -mod=vendor -o server.exe .
Set-Location ../frontend
npm.cmd ci --cache ../.offline/npm
npm.cmd run build
```

Создайте БД и заполните `backend/.env`, как описано выше. В двух отдельных терминалах:

```powershell
# Из DemoEx/backend
go run -mod=vendor . migrate
go run -mod=vendor . seed-admin
go run -mod=vendor .
```

```powershell
# Из DemoEx/frontend
npm.cmd start
```

`npm.cmd` работает и при ограничении PowerShell на запуск `npm.ps1`. Одновременно запускайте только одну установку npm для одной папки.

## Необязательный MySQL в Docker

Используйте этот путь, только если Docker уже установлен и разрешён. Основной путь выше обходится без контейнеров.

Скопируйте корневой `.env.example` в корневой `.env`, замените оба пароля, затем:

```bash
docker compose up -d
docker compose ps
```

MySQL 8.4 доступен на **127.0.0.1:3307**, чтобы не занимать порт уже установленного сервера. В `backend/.env` укажите `DB_PORT=3307`, имя БД/пользователя и пароль из корневого `.env`. Корневой `.env` настраивает Docker, `backend/.env` настраивает приложение.

Значения `MYSQL_*` используются при первом создании пустого хранилища. Изменение файла `.env` позже не меняет пароль уже созданной учётной записи. Данные сохраняются в именованном томе. Команда `docker compose stop` останавливает сервер с сохранением данных. Образ нужно скачать и запустить до отключения интернета.

## Если не работает

| Симптом | Проверить |
| --- | --- |
| `All mirrors were tried` | Строки выше содержат причину: DNS, HTTP, TLS. Проверить сеть и выпуск ОС, затем `sudo dnf makecache --refresh`. |
| `ECONNRESET` при npm | Сеть; повторить `npm ci --cache ../.offline/npm --fetch-retries=5` из `frontend` при работающем интернете. |
| `EBUSY`/`EPERM` на Windows | Остановить параллельные установки и dev-сервер, закрыть IDE, повторить одну установку. |
| `Access denied` от MySQL | `DB_USER`, пароль и разрешённый host; проверить `mysql -h 127.0.0.1 -P 3306 -u demo -p demoex`. |
| `Unknown database` | Создать БД или исправить `DB_NAME`; `migrate` создаёт таблицы, но не БД. |
| API 404/502 на порту 4200 | Запущен ли Go; проверить `frontend/proxy.conf.json` и перезапустить Angular. |
| 401 после входа | Один и тот же адрес браузера, cookies, `COOKIE_SECURE=false` для локального HTTP. |
| 403 на изменение | Права пользователя и адрес в `ALLOWED_ORIGINS`; админские маршруты требуют роль admin. |
| На 8080 старая страница | Пересобрать Angular; Go отдаёт `dist`, а не исходники. |

## Проверки

Из `backend`: `go test ./...` и `go vet ./...`. Из `frontend`: `npm run build`.

Для сквозной проверки API запустите сервер с **отдельной тестовой БД**, выполните `migrate` и `seed-admin`, затем из корня в Bash:

```bash
API_URL=http://127.0.0.1:8080 ADMIN_LOGIN=admin01 ADMIN_PASSWORD='ваш-тестовый-пароль' node tests/api-smoke.mjs
```

В PowerShell задайте эти переменные через `$env:API_URL`, `$env:ADMIN_LOGIN`, `$env:ADMIN_PASSWORD`, затем выполните `node tests/api-smoke.mjs`. Проверка создаёт двух тестовых пользователей и заявку, проходит CRUD, проверяет запрет чужих операций, смену статуса и отзыв. Заявка удаляется, тестовые пользователи остаются. Без переменных администратора админские шаги пропускаются.

CI проверяет Go-тесты, `go vet`, сборки и сценарий API с отдельной MySQL 8.4. Эти проверки не заменяют осмотр страниц на 390×844 и экзаменационный «Автотест». Редактируемая схема таблиц: [docs/schema.drawio](docs/schema.drawio).

Официальные справочники: [версии Angular](https://angular.dev/reference/versions), [прокси Angular](https://angular.dev/tools/cli/serve), [npm ci](https://docs.npmjs.com/cli/v11/commands/npm-ci/), [MySQL в Fedora](https://packages.fedoraproject.org/pkgs/mysql8.4/mysql-server/), [образ MySQL](https://hub.docker.com/_/mysql), [Go modules и vendor](https://go.dev/ref/mod#vendoring).
