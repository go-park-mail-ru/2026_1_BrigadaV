# Описание схемы базы данных и нормализации

## user
Attributes: id, email, password_hash, full_name, created_at, updated_at
Functional dependencies: {id} → all other attributes, {email} → all other attributes
Candidate keys: id, email
Primary key: id
1NF: all attributes atomic
2NF: no composite key, therefore no partial dependencies
3NF: no transitive dependencies
BCNF: all determinants are candidate keys

## session
Attributes: id, user_id, session_token, expires_at, created_at
Functional dependencies: {id} → all other attributes, {session_token} → all other attributes
Candidate keys: id, session_token
Primary key: id
Foreign key: user_id references user(id) on delete cascade
1NF: atomic
2NF: no composite key
3NF: no transitive dependencies
BCNF: all determinants are candidate keys

## favorite
Attributes: user_id, place_id, created_at
Functional dependencies: {user_id, place_id} → created_at
Candidate keys: (user_id, place_id)
Primary key: (user_id, place_id)
Foreign keys: user_id references user(id), place_id references place(id)
1NF: atomic
2NF: all non-key attributes depend on whole key (no partial dependencies)
3NF: no transitive dependencies
BCNF: determinant (user_id, place_id) is the key

## category
Attributes: id, name, description, created_at
Functional dependencies: {id} → all other attributes, {name} → all other attributes
Candidate keys: id, name
Primary key: id
1NF,2NF,3NF,BCNF satisfied

## country
Attributes: id, name, created_at
Functional dependencies: {id} → all other attributes, {name} → all other attributes
Candidate keys: id, name
Primary key: id
1NF,2NF,3NF,BCNF satisfied

## city
Attributes: id, name, country_id, created_at
Functional dependencies: {id} → all other attributes, {name, country_id} → id
Candidate keys: id, (name, country_id)
Primary key: id
Foreign key: country_id references country(id)
1NF: atomic
2NF: no partial dependencies (all non-key attributes depend on full key)
3NF: no transitive dependencies
BCNF: all determinants are keys

## place
Attributes: id, name, description, city_id, category_id, created_at, updated_at
Functional dependencies: {id} → all other attributes
Candidate keys: id
Primary key: id
Foreign keys: city_id references city(id), category_id references category(id)
1NF,2NF,3NF,BCNF satisfied

## place_photo
Attributes: id, place_id, file_path, is_main, created_at
Functional dependencies: {id} → all other attributes
Candidate keys: id
Primary key: id
Foreign key: place_id references place(id) on delete cascade
1NF,2NF,3NF,BCNF satisfied