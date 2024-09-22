CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    file_name VARCHAR(255),
    file_path TEXT,
    file_size BIGINT,
    format VARCHAR(50),
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

