# 2026_1_BrigadaV
Репозиторий бекенда команды Бригада V с проектом TripAdvisor

### Участники команды
 1. [Павлов Егор](https://github.com/TOOSTEER) — **Backend Developer / DB Developer**
 2. [Маркарова Юлия](https://github.com/jmark1518) — **Frontend Developer / Designer**
 3. [Лазунин Артём](https://github.com/cielnumerique) — **Team Lead / Backend Developer** 
 4. [Степан Бурматов](https://github.com/BurmatovStepan) — **Frontend Developer**

### Внешние ссылки
 - [Figma](https://www.figma.com/design/TuvbJqnQeBLsmuGjyBXAQC/Tripadvisor-Guidely)
 - [Frontend](https://github.com/frontend-park-mail-ru/2026_1_BrigadaV.git)
 - [Deploy](https://guidely.ru)
 - [Jira](https://ru.yougile.com/team/947473cd578d/TripAdvisor)
 - [Swagger](https://guidely.ru:8080/api/swagger)

### Технологический стек

 *   **Язык разработки:** Go 1.25.0
 *   **Взаимодействие микросервисов:** gRPC (Protocol Buffers)
 *   **Базы данных:** PostgreSQL 16 Alpine 
 *   **Система поиска:** Elasticsearch 8.13.4
 *   **Наблюдаемость (Observability):** Prometheus + Grafana (мониторинг)
 *   **CI/CD:** Автоматизация на базе GitHub Actions

 ### Архитектура системы

 Проект спроектирован по микросервисной архитектуре и состоит из следующих компонентов:
 *   **API Gateway (BFF):** Точка входа для клиентов. Маршрутизирует gRPC-запросы, управляет сессиями, CORS, защитой от CSRF, логированием.
 *   **Auth Service:** Отвечает за регистрацию, аутентификацию и управление сессиями пользователей. Хранит хеши сессионных токенов в PostgreSQL. Коммуникация через gRPC
 *   **Album Service:** Управление альбомами фотографий поездок. Обеспечивает загрузку, привязку и удаление фотографий. Файлы сохраняются в MinIO (S3‑совместимое хранилище) либо локально. Доступен через gRPC.
 *   **Review Service:** Обработка отзывов и рейтингов достопримечательностей. Позволяет создавать, удалять и получать отзывы. Взаимодействие через gRPC.

### Установка и запуск проекта

#### 1. Локальный запуск (Local Run)

Этот способ предназначен для проверки работоспособности и локального тестирования бэкенда на вашем компьютере.

##### Предварительные требования
 *   Установленный **Docker** и **Docker Compose**
 *   **Go 1.25.0** (если планируется запуск без Docker)

##### Быстрый старт через Docker Compose (Рекомендуется)
1.  Создайте локальный файл окружения:
    ```bash
    cp .env.example .env
    ```

2.  Запустите сборку и старт всей инфраструктуры и сервисов в фоне:

    ```bash
    docker compose up -d --build
    ```

3.  Выполните миграции баз данных и наполните их демонстрационными данными (сидами):
    ```bash
    goose -dir ./migrations postgres "DB_URL" up
    ```

После этого бэкенд будет доступен по адресу `http://localhost:8080`, а Swagger - по адресу `http://localhost:8080/swagger/index.html`.

---

#### 2. Развертывание на Production (Deployment)

Деплой всей системы автоматизирован.

##### Архитектура деплоя

 *   **CI/CD:** GitHub Actions (`.github/workflows/ci-cd.yaml`).

##### Процесс релиза (CI/CD)

 CI запускается при пуше в ветку `dev`
 CI/CD запускается при пуше в ветку `main`

1.  **Линтинг и статический анализ**
2.  **Тестирование**
3.  **Сборка (build)**
5.  **Деплой в продуктовую среду (production)**
