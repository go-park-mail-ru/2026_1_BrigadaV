TRUNCATE album_photo, album, trip_member, trip, review, place_photo, photo, favorite, session, "user", place, locality, country, category RESTART IDENTITY CASCADE;

INSERT INTO country (name) VALUES
    ('France'),
    ('Italy'),
    ('Spain'),
    ('Netherlands'),
    ('Indonesia'),
    ('Brazil');

INSERT INTO category (name, description) VALUES
    ('Hotel', 'Hotels and accommodations'),
    ('Museum', 'Art and history museums'),
    ('Historical place', 'Ancient ruins and landmarks'),
    ('Square', 'City squares'),
    ('Resort', 'Resort areas and recreation');

INSERT INTO locality (name, country_id, latitude, longitude) VALUES
    ('Gramado', (SELECT id FROM country WHERE name = 'Brazil'), -29.3733, -50.8762),
    ('Paris', (SELECT id FROM country WHERE name = 'France'), 48.8566, 2.3522),
    ('Rome', (SELECT id FROM country WHERE name = 'Italy'), 41.9028, 12.4964),
    ('Barcelona', (SELECT id FROM country WHERE name = 'Spain'), 41.3851, 2.1734),
    ('Amsterdam', (SELECT id FROM country WHERE name = 'Netherlands'), 52.3676, 4.9041),
    ('Bali', (SELECT id FROM country WHERE name = 'Indonesia'), -8.4095, 115.1889);

INSERT INTO place (name, description, locality_id, category_id, price) VALUES
    ('Hotel Estalagem St Hubertus', 'Charming hotel in Gramado', 1, 1, 2370000),
    ('Hotel Ritta Höppner', 'Cozy hotel in Gramado', 1, 1, 1138100),
    ('Rodin Musée', 'Museum dedicated to Auguste Rodin', 2, 2, 126900),
    ('Roman Forum', 'Ancient Roman forum', 3, 3, 126900),
    ('Basílica de Santa María del Pi', 'Gothic church in Barcelona', 4, 3, 199400),
    ('De Hallen Amsterdam', 'Cultural complex in Amsterdam', 5, 2, 3398800),
    ('Amnaya Resort Kuta', 'Resort in Bali', 6, 5, 584400),
    ('Plaça Reial', 'Historic square in Barcelona', 4, 4, 1236900);

INSERT INTO photo (file_path) VALUES
    ('/photos/hotel_estalagem.jpg'),
    ('/photos/hotel_ritta.jpg'),
    ('/photos/rodin.jpg'),
    ('/photos/roman_forum.jpg'),
    ('/photos/basilica_pi.jpg'),
    ('/photos/de_hallen.jpg'),
    ('/photos/amnaya.jpg'),
    ('/photos/placa_reial.jpg');

INSERT INTO place_photo (place_id, photo_id, is_main) VALUES
    (1, 1, true),
    (2, 2, true),
    (3, 3, true),
    (4, 4, true),
    (5, 5, true),
    (6, 6, true),
    (7, 7, true),
    (8, 8, true);

INSERT INTO "user" (email, nickname, avatar_url, password_hash, created_at, updated_at) VALUES
    ('john@example.com', 'johnny', '/avatars/john.jpg', 'argon2id$v=19$m=65536,t=1,p=4$LFU4f51KpaFJ85VzwIXZ2Q$NjKqQ4SfxdTnOJz22q+B8sYtNiTcOA4eozfj7mNJtnY', NOW(), NOW()),
    ('jane@example.com', 'jane', '/avatars/jane.jpg', 'argon2id$v=19$m=65536,t=1,p=4$LFU4f51KpaFJ85VzwIXZ2Q$NjKqQ4SfxdTnOJz22q+B8sYtNiTcOA4eozfj7mNJtnY', NOW(), NOW());

INSERT INTO favorite (user_id, place_id, created_at) VALUES
    (1, 1, NOW()),
    (1, 3, NOW()),
    (2, 5, NOW());

INSERT INTO review (user_id, place_id, rating, comment, visit_date, created_at, updated_at) VALUES
    (1, 1, 5, 'Excellent hotel!', '2025-01-10', NOW(), NOW()),
    (2, 3, 4, 'Interesting museum', '2025-02-15', NOW(), NOW());

INSERT INTO trip (title, description, start_date, end_date, created_by, is_public, created_at, updated_at) VALUES
    ('Trip to Europe', 'Visiting Paris and Rome', '2025-06-01', '2025-06-15', 1, true, NOW(), NOW()),
    ('Bali Vacation', 'Beach holiday', '2025-07-10', '2025-07-20', 2, false, NOW(), NOW());

INSERT INTO trip_member (trip_id, user_id, role, joined_at) VALUES
    (1, 1, 'owner', NOW()),
    (1, 2, 'viewer', NOW()),
    (2, 2, 'owner', NOW());

INSERT INTO album (trip_id, name, description, cover_photo_id, created_at, updated_at) VALUES
    (1, 'Paris', 'Photos from Paris', 3, NOW(), NOW()),
    (1, 'Rome', 'Photos from Rome', 4, NOW(), NOW()),
    (2, 'Bali', 'Beaches and sunsets', 7, NOW(), NOW());

INSERT INTO album_photo (album_id, photo_id, order_index, created_at) VALUES
    (1, 3, 0, NOW()),
    (1, 8, 1, NOW()),
    (2, 4, 0, NOW()),
    (3, 7, 0, NOW());