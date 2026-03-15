# 1. Описание схемы базы данных и нормализация

## 1.1. Общие сведения
Данный документ описывает структуру базы данных сервиса планирования путешествий Guidely. Все отношения приведены к нормальной форме Бойса-Кодда (BCNF). Для каждого отношения приведены атрибуты, функциональные зависимости, ключи и доказательства соответствия нормальным формам в соответствии с требованиями.

---

## 2. Таблица user – пользователи системы

### 2.1. Атрибуты
1. id (BIGINT)
2. email (TEXT)
3. nickname (TEXT)
4. avatar_url (TEXT)
5. password_hash (TEXT)
6. created_at (TIMESTAMPTZ)
7. updated_at (TIMESTAMPTZ)

### 2.2. Ограничения
1. email: уникален, CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
2. nickname: уникален, CHECK (char_length(nickname) >= 3 AND char_length(nickname) <= 50)
3. password_hash: TEXT, длина не ограничена

### 2.3. Функциональные зависимости
1. {id} → email, nickname, avatar_url, password_hash, created_at, updated_at
2. {email} → id, nickname, avatar_url, password_hash, created_at, updated_at
3. {nickname} → id, email, avatar_url, password_hash, created_at, updated_at

### 2.4. Ключи
1. Ключи-кандидаты: id, email, nickname
2. Первичный ключ: id

### 2.5. Нормальные формы
1. **1НФ:**  
   - Все атрибуты атомарны: каждое поле содержит неделимые значения (например, email не разбит на части).
   - Нет повторяющихся групп или массивов.
   - Схема данных не опирается на порядок строк или столбцов.
   - Благодаря первичному ключу (id) строки уникальны.
   - Каждое пересечение строки и столбца содержит ровно одно значение из предметной области.
2. **2НФ:**  
   - Отношение находится в 1НФ.
   - Отсутствует составной ключ, поэтому частичные зависимости невозможны.
3. **3НФ:**  
   - Отношение находится в 2НФ.
   - Нет транзитивных зависимостей: все неключевые атрибуты (email, nickname, avatar_url, password_hash, created_at, updated_at) зависят только от первичного ключа id. Например, nickname не зависит от email, а avatar_url не зависит от nickname.
4. **BCNF:**  
   - Все детерминанты (id, email, nickname) являются ключами-кандидатами. Следовательно, отношение находится в BCNF.

---

## 3. Таблица session – сессии пользователей

### 3.1. Атрибуты
1. id (BIGINT)
2. user_id (BIGINT)
3. session_token (TEXT)
4. expires_at (TIMESTAMPTZ)
5. created_at (TIMESTAMPTZ)

### 3.2. Ограничения
1. session_token уникален
2. user_id REFERENCES user(id) ON DELETE CASCADE

### 3.3. Функциональные зависимости
1. {id} → user_id, session_token, expires_at, created_at
2. {session_token} → id, user_id, expires_at, created_at

### 3.4. Ключи
1. Ключи-кандидаты: id, session_token
2. Первичный ключ: id
3. Внешний ключ: user_id → user(id)

### 3.5. Нормальные формы
1. **1НФ:** Все атрибуты атомарны; строки уникальны благодаря PK; порядок не важен.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Все детерминанты являются ключами-кандидатами.

---

## 4. Таблица favorite – избранное пользователей

### 4.1. Атрибуты
1. user_id (BIGINT)
2. place_id (BIGINT)
3. created_at (TIMESTAMPTZ)

### 4.2. Функциональные зависимости
1. {user_id, place_id} → created_at

### 4.3. Ключи
1. Ключи-кандидаты: (user_id, place_id)
2. Первичный ключ: (user_id, place_id)
3. Внешние ключи: user_id → user(id), place_id → place(id)

### 4.4. Нормальные формы
1. **1НФ:** Атомарность; строки уникальны благодаря составному PK.
2. **2НФ:** Единственный неключевой атрибут (created_at) зависит от полного составного ключа; частичных зависимостей нет.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант (user_id, place_id) является ключом.

---

## 5. Таблица review – отзывы

### 5.1. Атрибуты
1. id (BIGINT)
2. user_id (BIGINT)
3. place_id (BIGINT)
4. rating (SMALLINT)
5. comment (TEXT)
6. visit_date (DATE)
7. created_at (TIMESTAMPTZ)
8. updated_at (TIMESTAMPTZ)

### 5.2. Ограничения
1. rating CHECK (rating >= 1 AND rating <= 5)
2. UNIQUE (user_id, place_id)

### 5.3. Функциональные зависимости
1. {id} → user_id, place_id, rating, comment, visit_date, created_at, updated_at

### 5.4. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id
3. Внешние ключи: user_id → user(id), place_id → place(id)

### 5.5. Нормальные формы
1. **1НФ:** Атомарность, уникальность строк.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант id является ключом.

---

## 6. Таблица category – категории

### 6.1. Атрибуты
1. id (BIGINT)
2. name (TEXT)
3. description (TEXT)
4. created_at (TIMESTAMPTZ)

### 6.2. Ограничения
1. name уникален, CHECK (char_length(name) >= 1 AND char_length(name) <= 100)

### 6.3. Функциональные зависимости
1. {id} → name, description, created_at
2. {name} → id, description, created_at

### 6.4. Ключи
1. Ключи-кандидаты: id, name
2. Первичный ключ: id

### 6.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Все детерминанты являются ключами-кандидатами.

---

## 7. Таблица country – страны

### 7.1. Атрибуты
1. id (BIGINT)
2. name (TEXT)
3. created_at (TIMESTAMPTZ)

### 7.2. Ограничения
1. name уникален, CHECK (char_length(name) >= 2 AND char_length(name) <= 100)

### 7.3. Функциональные зависимости
1. {id} → name, created_at
2. {name} → id, created_at

### 7.4. Ключи
1. Ключи-кандидаты: id, name
2. Первичный ключ: id

### 7.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Все детерминанты являются ключами-кандидатами.

---

## 8. Таблица locality – населённые пункты

### 8.1. Атрибуты
1. id (BIGINT)
2. name (TEXT)
3. country_id (BIGINT)
4. latitude (NUMERIC(10,8))
5. longitude (NUMERIC(11,8))
6. created_at (TIMESTAMPTZ)

### 8.2. Ограничения
1. name CHECK (char_length(name) >= 1 AND char_length(name) <= 200)
2. latitude CHECK (latitude >= -90 AND latitude <= 90)
3. longitude CHECK (longitude >= -180 AND longitude <= 180)
4. UNIQUE(name, country_id)

### 8.3. Функциональные зависимости
1. {id} → name, country_id, latitude, longitude, created_at
2. {name, country_id} → id, latitude, longitude, created_at

### 8.4. Ключи
1. Ключи-кандидаты: id, (name, country_id)
2. Первичный ключ: id
3. Внешний ключ: country_id → country(id)

### 8.5. Нормальные формы
1. **1НФ:** Все атрибуты атомарны; нет повторяющихся групп; наличие первичного ключа гарантирует уникальность.
2. **2НФ:** Рассмотрим составной ключ (name, country_id). Все неключевые атрибуты (latitude, longitude, created_at) зависят от полного ключа, а не от его части. Например, не может быть зависимости {name} → latitude, так как одно и то же название города может встречаться в разных странах с разными координатами. Следовательно, частичные зависимости отсутствуют. Для простого ключа id также нет частичных зависимостей.
3. **3НФ:** Нет транзитивных зависимостей: координаты напрямую зависят от локации, а не от других неключевых атрибутов (например, от country_id, который сам является внешним ключом).
4. **BCNF:** Все детерминанты (id и (name, country_id)) являются ключами-кандидатами.

---

## 9. Таблица place – достопримечательности

### 9.1. Атрибуты
1. id (BIGINT)
2. name (TEXT)
3. description (TEXT)
4. locality_id (BIGINT)
5. category_id (BIGINT)
6. price (BIGINT)
7. created_at (TIMESTAMPTZ)
8. updated_at (TIMESTAMPTZ)

### 9.2. Ограничения
1. name CHECK (char_length(name) >= 1 AND char_length(name) <= 255)
2. price CHECK (price >= 0)

### 9.3. Функциональные зависимости
1. {id} → name, description, locality_id, category_id, price, created_at, updated_at

### 9.4. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id
3. Внешние ключи: locality_id → locality(id), category_id → category(id)

### 9.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей (все неключевые атрибуты зависят только от id).
4. **BCNF:** Детерминант id является ключом.

---

## 10. Таблица photo – фотографии

### 10.1. Атрибуты
1. id (BIGINT)
2. file_path (TEXT)
3. created_at (TIMESTAMPTZ)

### 10.2. Функциональные зависимости
1. {id} → file_path, created_at

### 10.3. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id

### 10.4. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант id является ключом.

---

## 11. Таблица place_photo – связь мест и фотографий

### 11.1. Атрибуты
1. id (BIGINT)
2. place_id (BIGINT)
3. photo_id (BIGINT)
4. is_main (BOOLEAN)
5. created_at (TIMESTAMPTZ)

### 11.2. Функциональные зависимости
1. {id} → place_id, photo_id, is_main, created_at

### 11.3. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id
3. Внешние ключи: place_id → place(id), photo_id → photo(id)

### 11.4. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант id является ключом.

---

## 12. Таблица trip – поездки

### 12.1. Атрибуты
1. id (BIGINT)
2. title (TEXT)
3. description (TEXT)
4. start_date (DATE)
5. end_date (DATE)
6. created_by (BIGINT)
7. is_public (BOOLEAN)
8. created_at (TIMESTAMPTZ)
9. updated_at (TIMESTAMPTZ)

### 12.2. Ограничения
1. title CHECK (char_length(title) >= 1 AND char_length(title) <= 255)
2. CHECK (start_date <= end_date)

### 12.3. Функциональные зависимости
1. {id} → title, description, start_date, end_date, created_by, is_public, created_at, updated_at

### 12.4. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id
3. Внешний ключ: created_by → user(id)

### 12.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант id является ключом.

---

## 13. Таблица trip_member – участники поездок

### 13.1. Атрибуты
1. trip_id (BIGINT)
2. user_id (BIGINT)
3. role (TEXT)
4. joined_at (TIMESTAMPTZ)

### 13.2. Ограничения
1. role CHECK (role IN ('owner', 'editor', 'viewer'))

### 13.3. Функциональные зависимости
1. {trip_id, user_id} → role, joined_at

### 13.4. Ключи
1. Ключи-кандидаты: (trip_id, user_id)
2. Первичный ключ: (trip_id, user_id)
3. Внешние ключи: trip_id → trip(id), user_id → user(id)

### 13.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Все неключевые атрибуты (role, joined_at) зависят от полного составного ключа; частичных зависимостей нет (например, role не зависит только от trip_id).
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант (trip_id, user_id) является ключом.

---

## 14. Таблица album – альбомы

### 14.1. Атрибуты
1. id (BIGINT)
2. trip_id (BIGINT)
3. name (TEXT)
4. description (TEXT)
5. cover_photo_id (BIGINT)
6. created_at (TIMESTAMPTZ)
7. updated_at (TIMESTAMPTZ)

### 14.2. Ограничения
1. name CHECK (char_length(name) >= 1 AND char_length(name) <= 255)

### 14.3. Функциональные зависимости
1. {id} → trip_id, name, description, cover_photo_id, created_at, updated_at

### 14.4. Ключи
1. Ключи-кандидаты: id
2. Первичный ключ: id
3. Внешние ключи: trip_id → trip(id), cover_photo_id → photo(id)

### 14.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Нет составного ключа.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант id является ключом.

---

## 15. Таблица album_photo – связь альбомов и фотографий

### 15.1. Атрибуты
1. album_id (BIGINT)
2. photo_id (BIGINT)
3. order_index (SMALLINT)
4. created_at (TIMESTAMPTZ)

### 15.2. Ограничения
1. order_index CHECK (order_index >= 0)

### 15.3. Функциональные зависимости
1. {album_id, photo_id} → order_index, created_at

### 15.4. Ключи
1. Ключи-кандидаты: (album_id, photo_id)
2. Первичный ключ: (album_id, photo_id)
3. Внешние ключи: album_id → album(id), photo_id → photo(id)

### 15.5. Нормальные формы
1. **1НФ:** Атомарность.
2. **2НФ:** Все неключевые атрибуты зависят от полного составного ключа; частичных зависимостей нет.
3. **3НФ:** Нет транзитивных зависимостей.
4. **BCNF:** Детерминант (album_id, photo_id) является ключом.

