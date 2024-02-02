CREATE TABLE IF NOT EXISTS recipes (
    recipe_id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    recipe_text TEXT NOT NULL,
    publisher_username VARCHAR(50) NOT NULL,
    published_date DATE NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
    );