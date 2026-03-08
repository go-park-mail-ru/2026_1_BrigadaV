INSERT INTO country (name) VALUES 
    ('France'), 
    ('Italy'), 
    ('USA');

INSERT INTO locality (name, country_id, latitude, longitude) VALUES
    ('Paris', 1, 48.8566, 2.3522),
    ('Rome', 2, 41.9028, 12.4964),
    ('New York', 3, 40.7128, -74.0060);

INSERT INTO category (name, description) VALUES
    ('Museum', 'Art and history museums'),
    ('Park', 'City parks and nature reserves'),
    ('Restaurant', 'Places to eat');

INSERT INTO place (name, description, locality_id, category_id) VALUES
    ('Eiffel Tower', 'Famous tower', 1, 2),
    ('Colosseum', 'Ancient amphitheater', 2, 1),
    ('Statue of Liberty', 'Gift from France', 3, 2);

INSERT INTO place_photo (place_id, file_path, is_main) VALUES
    (1, '/photos/eiffel.jpg', true),
    (1, '/photos/eiffel_night.jpg', false),
    (2, '/photos/colosseum.jpg', true),
    (3, '/photos/statue.jpg', true);