INSERT INTO category (name, description) VALUES
    ('Музей', 'Художественные, исторические музеи'),
    ('Парк', 'Городские парки и заповедники'),
    ('Ресторан', 'Места для питания');

INSERT INTO location (city, country, latitude, longitude) VALUES
    ('Париж', 'Франция', 48.8566, 2.3522),
    ('Рим', 'Италия', 41.9028, 12.4964),
    ('Нью-Йорк', 'США', 40.7128, -74.0060);

INSERT INTO attraction (name, description, location_id, category_id) VALUES
    ('Эйфелева башня', 'Знаменитая башня', 1, 2),
    ('Колизей', 'Древний амфитеатр', 2, 1),
    ('Статуя Свободы', 'Подарок Франции', 3, 2);

INSERT INTO photo (attraction_id, file_path, is_main) VALUES
    (1, '/photos/eiffel.jpg', true),
    (2, '/photos/colosseum.jpg', true),
    (3, '/photos/statue.jpg', true);