

# SalesTracker

### **CRUD**

* `POST /items` — создать запись
* `GET /items` — получить список всех записей
* `GET /items/{id}` — получить запись по ID
* `PUT /items/{id}` — обновить запись
* `DELETE /items/{id}` — удалить запись

### **Analytics**

###### Не до конца понял задание, поэтому принял решение для взаимодействия с аналитикой сделать параметр типа операции обязательной

* `GET /analytics?type=...&from=...&to=...`
  Возвращает:
* сумму
* среднее
* количество
* медиану
* 90-й перцентиль

---

## Запуск проекта

### 1. Клонировать репозиторий

```sh
git clone https://github.com/v1adis1av28/Level3.git
cd Level3/SalesTracker
```

### 2. Запустить Docker

```sh
docker-compose up --build
```

Сервис поднимется на:

* Backend → **[http://localhost:8080](http://localhost:8080)**
* Frontend → **[http://localhost:8080/static/index.html](http://localhost:8080/static/index.html)**

---

### Получить аналитику

```bash
curl "http://localhost:8080/analytics?type=food&from=2025-01-01&to=2025-01-31"
```

---

## Требования к данным

* `price` > 0
* `name` — строка
* `type` — строка
* `date` (опционально) — валидная дата
