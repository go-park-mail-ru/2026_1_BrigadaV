INSERT INTO category (name, description) VALUES
    ('Museum', 'Artistic, historical museums'),
    ('Park', 'City parks and reserves'),
    ('Restaurant', 'Places for dining');

INSERT INTO location (city, country, latitude, longitude) VALUES
    ('Paris', 'France', 48.8566, 2.3522),
    ('Rome', 'Italy', 41.9028, 12.4964),
    ('New York', 'USA', 40.7128, -74.0060);

INSERT INTO attraction (name, description, location_id, category_id) VALUES
    ('Eiffel Tower', 'Famous tower', 1, 2),
    ('Colosseum', 'Ancient amphitheater', 2, 1),
    ('Statue of Liberty', 'Gift from France', 3, 2);

INSERT INTO photo (attraction_id, file_path, is_main) VALUES
    (1, '/photos/eiffel.jpg', true),
    (2, '/photos/colosseum.jpg', true),
    (3, '/photos/statue.jpg', true);
