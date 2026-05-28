-- Очистка старых данных (опционально) 
-- TRUNCATE trip_attractions, trip_member, trip, review, place, locality, "user" CASCADE;

-- Пользователи (50k)
INSERT INTO "user" (login, nickname, password_hash, country, city)
SELECT 
    'user' || i || '@example.com',
    'nick' || i,
    'hash',
    CASE (i % 10) WHEN 0 THEN 'Russia' WHEN 1 THEN 'USA' ELSE 'France' END,
    'City' || (i % 100)
FROM generate_series(1, 50000) AS i;

-- Локации (1000)
INSERT INTO locality (name, country_id)
SELECT 'Locality ' || i, (i % 10) + 1
FROM generate_series(1, 1000) AS i;

-- Места (100k)
INSERT INTO place (name, description, locality_id, category_id, price, rating, review_count, latitude, longitude)
SELECT 
    'Place ' || i,
    'Description for place ' || i,
    (i % 1000) + 1,
    (i % 8) + 1,
    (random() * 50000)::int,
    (random() * 5)::numeric(2,1),
    (random() * 100)::int,
    random() * 180 - 90,
    random() * 360 - 180
FROM generate_series(1, 100000) AS i;

-- Отзывы (200k)
INSERT INTO review (user_id, place_id, rating, comment, created_at)
SELECT 
    (random() * 50000)::int + 1,
    (random() * 100000)::int + 1,
    (random() * 4 + 1)::int,
    'Review comment ' || i,
    NOW() - (random() * 365 * 24 * 3600)::int * interval '1 second'
FROM generate_series(1, 200000) AS i
ON CONFLICT DO NOTHING;

-- Поездки (30k)
INSERT INTO trip (title, created_by, start_date, end_date, is_public)
SELECT 
    'Trip ' || i,
    (random() * 50000)::int + 1,
    NOW() - (random() * 365 * 24 * 3600)::int * interval '1 second',
    NOW() + (random() * 30 * 24 * 3600)::int * interval '1 second',
    random() > 0.5
FROM generate_series(1, 30000) AS i;

-- Участники поездок (по 2 случайных на поездку)
INSERT INTO trip_member (trip_id, user_id, role)
SELECT 
    t.id,
    (random() * 50000)::int + 1,
    CASE WHEN random() > 0.7 THEN 'companion' ELSE 'viewer' END
FROM trip t
CROSS JOIN generate_series(1, 2)
ON CONFLICT DO NOTHING;

-- Достопримечательности в поездках (100k записей)
INSERT INTO trip_attractions (trip_id, place_id, order_index)
SELECT 
    t.id,
    (random() * 100000)::int + 1,
    order_idx
FROM trip t
CROSS JOIN LATERAL generate_series(1, (random() * 7 + 3)::int) AS order_idx
LIMIT 100000;