CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    file_name VARCHAR(255),
    file_path TEXT UNIQUE,
    file_size BIGINT,
    format VARCHAR(50),
    overview TEXT,
    poster_url TEXT,
    added_at TIMESTAMP DEFAULT NOW()
);

