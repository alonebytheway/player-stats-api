<div align="center">

  <img src="https://capsule-render.vercel.app/api?type=waving&color=timeGradient&height=220&section=header&text=Player%20Stats%20API&fontSize=60&fontAlignY=35&desc=Backend%20Service%20on%20Go&descAlignY=58&descAlign=50" alt="header banner" />

  <br>

  ![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
  ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql&logoColor=white)
  ![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)
  ![CI/CD](https://img.shields.io/badge/CI-GitHub_Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white)

  <br>
  
  **RESTful API сервис для отслеживания игровой статистики, проведения дуэлей и формирования таблицы лидеров.**
  <br>
  <i>Написан на Go с соблюдением принципов Clean Architecture.</i>

</div>

---

## ⚡ Главные фичи

Мы не просто храним данные, мы делаем это надежно и быстро:

- 🏎️ **Асинхронность:** Таблица лидеров обновляется в фоне (Goroutines + Channels) и отдается из In-Memory кэша без задержек.
- 🛡️ **Надежность:** Встроенная защита транзакций от Deadlock'ов при конкурентных запросах.
- 🔐 **Безопасность:** Доступ к методам закрыт через `AuthMiddleware` (ApiKeyAuth). Защита от падений через `Recoverer`.
- 📦 **DevOps Ready:** Полная контейнеризация (Docker-Compose), управление через `Makefile` и настроенный пайплайн тестирования (GitHub Actions).
- 📚 **Документация:** Автогенерируемый интерактивный интерфейс Swagger.

---

## 🗄️ Маршрутизация (API Endpoints)

> 💡 *Доступ к защищенным эндпоинтам требует передачи заголовка `Authorization`.*

| Метод | Путь | Описание | Доступ |
| :--- | :--- | :--- | :---: |
| `GET` | `/swagger/*` | Интерактивная документация | 🟢 Всем |
| `GET` | `/leaderboard` | Получить топ игроков (кэш) | 🔴 Авторизация |
| `GET` | `/players` | Получить список всех игроков | 🔴 Авторизация |
| `POST` | `/players` | Зарегистрировать нового игрока | 🔴 Авторизация |
| `PATCH` | `/players/{name}`| Частичное обновление (убийства, смерти) | 🔴 Авторизация |
| `POST` | `/players/duel` | Записать результаты дуэли | 🔴 Авторизация |

---

## 🚀 Быстрый старт

Проект запускается одной командой. Убедитесь, что у вас установлены **Docker** и **Make**.

**1. Склонируйте репозиторий:**
```bash
git clone [https://github.com/alonebytheway/player-stats-api.git](https://github.com/alonebytheway/player-stats-api.git)
cd player-stats-api

**2. Запустите магию Docker:**
```bash
make up

**3. Откройте документацию:**

Перейдите в браузере по адресу 👉 http://localhost:8080/swagger/index.html

**🛠 Пульт управления (Makefile)**
Если вы хотите поработать с кодом локально, используйте эти шорткаты:

```bash
make run    # Запустить сервер локально (без Docker)
make test   # Прогнать все Unit-тесты
make swag   # Обновить документацию Swagger
make up     # Поднять БД и Сервер в Docker
make down   # Остановить и удалить контейнеры
make tidy   # Причесать зависимости go.mod
