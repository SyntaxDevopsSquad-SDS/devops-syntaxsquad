-- Initial database schema for PostgreSQL

CREATE TABLE IF NOT EXISTS users (
    id                        SERIAL PRIMARY KEY,
    username                  TEXT NOT NULL UNIQUE,
    email                     TEXT NOT NULL UNIQUE,
    password                  TEXT NOT NULL,
    force_password_reset      BOOLEAN DEFAULT FALSE,
    password_reset_required_at TIMESTAMP
);

INSERT INTO users (username, email, password)
VALUES ('admin', 'admin@whoknows.com', '$2a$10$v/spwONyDHojGbiU6V36BOcKJ/bSt9kO2pl41JJ/CMo0ZcruhWwvq')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at    TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON password_reset_tokens(token);

CREATE TABLE IF NOT EXISTS pages (
    id           SERIAL PRIMARY KEY,
    title        TEXT NOT NULL UNIQUE,
    url          TEXT NOT NULL UNIQUE,
    language     TEXT NOT NULL DEFAULT 'en' CHECK(language IN ('en', 'da')),
    last_updated TIMESTAMP,
    content      TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pages_language ON pages(language);

-- PostgreSQL fulltekst-søgning med tsvector
ALTER TABLE pages ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE INDEX IF NOT EXISTS idx_pages_search_vector ON pages USING GIN(search_vector);

CREATE OR REPLACE FUNCTION pages_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', NEW.title || ' ' || NEW.content);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER pages_search_vector_trigger
    BEFORE INSERT OR UPDATE ON pages
    FOR EACH ROW EXECUTE FUNCTION pages_search_vector_update();

INSERT INTO pages (title, url, language, content) VALUES
('Fortran',    'http://web.archive.org/web/20081220110619/http://en.wikipedia.org:80/wiki/Fortran',    'en', 'Fortran'),
('Algorithm',  'http://web.archive.org/web/20081217070911/http://en.wikipedia.org:80/wiki/Algorithm',  'en', 'Algorithm'),
('MATLAB',     'http://web.archive.org/web/20090110165251/http://en.wikipedia.org:80/wiki/Matlab',     'en', 'MATLAB'),
('JavaScript', 'http://web.archive.org/web/20081218123622/http://en.wikipedia.org:80/wiki/Javascript', 'en', 'JavaScript'),
('Database',   'http://web.archive.org/web/20081219060743/http://en.wikipedia.org:80/wiki/Database',   'en', 'Database')
ON CONFLICT DO NOTHING;
