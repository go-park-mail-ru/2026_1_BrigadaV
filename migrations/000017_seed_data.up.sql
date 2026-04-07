
TRUNCATE album_photo, album, trip_member, trip, review, place_photo, photo, favorite, session, "user", place, locality, country, category RESTART IDENTITY CASCADE;

INSERT INTO country (name) VALUES
    ('Франция'),
    ('Италия'),
    ('Испания'),
    ('Нидерланды'),
    ('Индонезия'),
    ('Бразилия'),
    ('Германия'),
    ('Великобритания'),
    ('Япония'),
    ('Таиланд');

INSERT INTO category (name, description) VALUES
    ('Отель', 'Гостиницы и места размещения'),
    ('Музей', 'Художественные и исторические музеи'),
    ('Историческое место', 'Древние руины и достопримечательности'),
    ('Площадь', 'Городские площади'),
    ('Курорт', 'Курортные зоны и отдых'),
    ('Парк', 'Городские парки и заповедники'),
    ('Ресторан', 'Заведения общественного питания'),
    ('Торговый центр', 'Крупные торговые комплексы');

INSERT INTO locality (name, country_id, latitude, longitude) VALUES
    ('Грамаду', (SELECT id FROM country WHERE name = 'Бразилия'), -29.3733, -50.8762),
    ('Париж', (SELECT id FROM country WHERE name = 'Франция'), 48.8566, 2.3522),
    ('Рим', (SELECT id FROM country WHERE name = 'Италия'), 41.9028, 12.4964),
    ('Барселона', (SELECT id FROM country WHERE name = 'Испания'), 41.3851, 2.1734),
    ('Амстердам', (SELECT id FROM country WHERE name = 'Нидерланды'), 52.3676, 4.9041),
    ('Бали', (SELECT id FROM country WHERE name = 'Индонезия'), -8.4095, 115.1889),
    ('Берлин', (SELECT id FROM country WHERE name = 'Германия'), 52.5200, 13.4050),
    ('Лондон', (SELECT id FROM country WHERE name = 'Великобритания'), 51.5074, -0.1278),
    ('Токио', (SELECT id FROM country WHERE name = 'Япония'), 35.6895, 139.6917),
    ('Бангкок', (SELECT id FROM country WHERE name = 'Таиланд'), 13.7367, 100.5231);

INSERT INTO place (name, description, locality_id, category_id, price) VALUES
    ('Hotel Estalagem St Hubertus', 'Очаровательный отель в Грамаду', 1, 1, 2370000),
    ('Hotel Ritta Höppner', 'Уютный отель в Грамаду', 1, 1, 1138100),
    ('Rodin Musée', 'Музей, посвящённый Огюсту Родену', 2, 2, 126900),
    ('Roman Forum', 'Древний римский форум', 3, 3, 126900),
    ('Basílica de Santa María del Pi', 'Готическая церковь в Барселоне', 4, 3, 199400),
    ('De Hallen Amsterdam', 'Культурный комплекс в Амстердаме', 5, 2, 3398800),
    ('Amnaya Resort Kuta', 'Курорт на Бали', 6, 5, 584400),
    ('Plaça Reial', 'Историческая площадь в Барселоне', 4, 4, 1236900),
    ('Бранденбургские ворота', 'Символ Берлина', 7, 3, 0),
    ('Британский музей', 'Один из крупнейших музеев мира', 8, 2, 0),
    ('Senso-ji', 'Древний буддийский храм в Токио', 9, 3, 0),
    ('Ват Арун', 'Храм рассвета в Бангкоке', 10, 3, 0),
    ('Лувр', 'Крупнейший художественный музей', 2, 2, 170000),
    ('Эйфелева башня', 'Знаменитая башня в Париже', 2, 3, 150000),
    ('Колизей', 'Древний амфитеатр', 3, 3, 160000),
    ('Парк Гуэль', 'Парк с архитектурными объектами Гауди', 4, 6, 100000),
    ('Музей Ван Гога', 'Музей, посвящённый Ван Гогу', 5, 2, 200000);

INSERT INTO photo (file_path) VALUES
    ('mock/place/rcmd1.png'),
    ('mock/place/rcmd2.png'),
    ('mock/place/rcmd3.png'),
    ('mock/place/rcmd4.png'),
    ('mock/place/rcmd5.png'),
    ('mock/place/rcmd6.png'),
    ('mock/place/rcmd7.png'),
    ('mock/place/rcmd8.png'),
    ('mock/place/rcmd9.png'),
    ('mock/place/rcmd10.png'),
    ('mock/place/rcmd11.png'),
    ('mock/place/rcmd12.png'),
    ('mock/place/rcmd13.png'),
    ('mock/place/rcmd14.png'),
    ('mock/place/rcmd15.png'),
    ('mock/place/rcmd16.png'),
    ('mock/place/rcmd17.png');

INSERT INTO place_photo (place_id, photo_id, is_main) VALUES
    (1, 1, true),
    (2, 2, true),
    (3, 3, true),
    (4, 4, true),
    (5, 5, true),
    (6, 6, true),
    (7, 7, true),
    (8, 8, true),
    (9, 9, true),
    (10, 10, true),
    (11, 11, true),
    (12, 12, true),
    (13, 13, true),
    (14, 14, true),
    (15, 15, true),
    (16, 16, true),
    (17, 17, true);


INSERT INTO "user" (email, nickname, avatar_url, password_hash, created_at, updated_at) VALUES
    ('john@example.com', 'johnny', 'mock/user-avatar/john.jpg', 'argon2id$v=19$m=65536,t=1,p=4$LFU4f51KpaFJ85VzwIXZ2Q$NjKqQ4SfxdTnOJz22q+B8sYtNiTcOA4eozfj7mNJtnY', NOW(), NOW()),
    ('jane@example.com', 'jane', 'mock/user-avatar/jane.jpg', 'argon2id$v=19$m=65536,t=1,p=4$LFU4f51KpaFJ85VzwIXZ2Q$NjKqQ4SfxdTnOJz22q+B8sYtNiTcOA4eozfj7mNJtnY', NOW(), NOW()),
    ('admin@example.com', 'admin', 'mock/user-avatar/admin.jpg', 'argon2id$v=19$m=65536,t=1,p=4$LFU4f51KpaFJ85VzwIXZ2Q$NjKqQ4SfxdTnOJz22q+B8sYtNiTcOA4eozfj7mNJtnY', NOW(), NOW());


INSERT INTO favorite (user_id, place_id, created_at) VALUES
    (1, 1, NOW()),
    (1, 3, NOW()),
    (1, 5, NOW()),
    (2, 2, NOW()),
    (2, 4, NOW()),
    (3, 6, NOW()),
    (3, 7, NOW()),
    (3, 8, NOW());

INSERT INTO review (user_id, place_id, rating, comment, visit_date, created_at, updated_at) VALUES
    (1, 1, 5, 'Отличный отель! Очень понравилось обслуживание.', '2025-01-10', NOW(), NOW()),
    (1, 3, 4, 'Интересный музей, но маловато экспонатов.', '2025-02-15', NOW(), NOW()),
    (2, 5, 5, 'Великолепная церковь! Очень красивая архитектура.', '2025-03-20', NOW(), NOW()),
    (2, 7, 4, 'Хороший курорт, но дороговато.', '2025-04-05', NOW(), NOW()),
    (3, 13, 5, 'Лувр – это нечто невероятное!', '2025-05-10', NOW(), NOW()),
    (3, 14, 5, 'Эйфелева башня – символ Парижа, обязательно к посещению.', '2025-05-12', NOW(), NOW()),
    (1, 15, 5, 'Колизей впечатляет!', '2025-06-01', NOW(), NOW()),
    (2, 16, 4, 'Красивый парк, но много туристов.', '2025-06-15', NOW(), NOW());

INSERT INTO trip (title, description, start_date, end_date, created_by, is_public, created_at, updated_at) VALUES
    ('Поездка в Европу', 'Посещение Парижа, Рима и Барселоны', '2025-06-01', '2025-06-15', 1, true, NOW(), NOW()),
    ('Отдых на Бали', 'Пляжный отдых и экскурсии', '2025-07-10', '2025-07-20', 2, false, NOW(), NOW()),
    ('Культурный тур по Европе', 'Музеи и исторические места', '2025-08-01', '2025-08-14', 1, true, NOW(), NOW()),
    ('Японское приключение', 'Путешествие по Японии', '2025-09-05', '2025-09-15', 3, true, NOW(), NOW());

INSERT INTO trip_member (trip_id, user_id, role, joined_at) VALUES
    (1, 1, 'owner', NOW()),
    (1, 2, 'viewer', NOW()),
    (2, 2, 'owner', NOW()),
    (2, 1, 'editor', NOW()),
    (3, 1, 'owner', NOW()),
    (3, 2, 'editor', NOW()),
    (3, 3, 'viewer', NOW()),
    (4, 3, 'owner', NOW());

INSERT INTO album (trip_id, name, description, cover_photo_id, created_at, updated_at) VALUES
    (1, 'Париж', 'Фотографии из Парижа', 3, NOW(), NOW()),
    (1, 'Рим', 'Фотографии из Рима', 4, NOW(), NOW()),
    (1, 'Барселона', 'Фотографии из Барселоны', 5, NOW(), NOW()),
    (2, 'Бали', 'Пляжи и закаты', 7, NOW(), NOW()),
    (3, 'Музеи Европы', 'Фотографии из музеев', 13, NOW(), NOW());

INSERT INTO album_photo (album_id, photo_id, order_index, created_at) VALUES
    (1, 3, 0, NOW()),
    (1, 14, 1, NOW()),
    (2, 4, 0, NOW()),
    (2, 15, 1, NOW()),
    (3, 5, 0, NOW()),
    (3, 8, 1, NOW()),
    (4, 7, 0, NOW()),
    (5, 13, 0, NOW());