CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL UNIQUE,
    full_url TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT now()
);