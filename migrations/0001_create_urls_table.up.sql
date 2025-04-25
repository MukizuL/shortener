CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    user_id UUID,
    short_url TEXT NOT NULL UNIQUE,
    full_url TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX urls_short_url_hash_idx ON urls USING HASH(short_url);
CREATE INDEX urls_user_id_hash_idx ON urls USING HASH(user_id);