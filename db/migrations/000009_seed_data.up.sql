INSERT INTO country (name) VALUES 
    ('France'), 
    ('Italy'), 
    ('USA');

INSERT INTO city (name, country_id) VALUES
    ('Paris', 1),
    ('Rome', 2),
    ('New York', 3);

INSERT INTO category (name, description) VALUES
    ('Museum', 'Art and history museums'),
    ('Park', 'City parks and nature reserves'),
    ('Restaurant', 'Places to eat');

INSERT INTO place (name, description, city_id, category_id) VALUES
    ('Eiffel Tower', 'Famous tower', 1, 2),
    ('Colosseum', 'Ancient amphitheater', 2, 1),
    ('Statue of Liberty', 'Gift from France', 3, 2);

INSERT INTO place_photo (place_id, file_path, is_main) VALUES
    (1, '/photos/eiffel.jpg', true),
    (2, '/photos/colosseum.jpg', true),
    (3, '/photos/statue.jpg', true);