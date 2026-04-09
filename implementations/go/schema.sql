DROP TABLE IF EXISTS pages_fts;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS users;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  force_password_reset BOOLEAN DEFAULT 0,
  password_reset_required_at TIMESTAMP
);

INSERT INTO users (username, email, password)
    VALUES ('admin', 'admin@whoknows.com', '$2a$10$v/spwONyDHojGbiU6V36BOcKJ/bSt9kO2pl41JJ/CMo0ZcruhWwvq');

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_token ON password_reset_tokens(token);

CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL DEFAULT 'en' CHECK(language IN ('en', 'da')),
    last_updated TIMESTAMP,
    content TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pages_language ON pages(language);

CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(
    title,
    content,
    language UNINDEXED,
    url UNINDEXED,
    content='pages',
    content_rowid='id'
);

CREATE TRIGGER pages_ai AFTER INSERT ON pages BEGIN
    INSERT INTO pages_fts(rowid, title, content, language, url)
    VALUES (new.id, new.title, new.content, new.language, new.url);
END;

CREATE TRIGGER pages_ad AFTER DELETE ON pages BEGIN
    INSERT INTO pages_fts(pages_fts, rowid, title, content, language, url)
    VALUES ('delete', old.id, old.title, old.content, old.language, old.url);
END;

CREATE TRIGGER pages_au AFTER UPDATE ON pages BEGIN
    INSERT INTO pages_fts(pages_fts, rowid, title, content, language, url)
    VALUES ('delete', old.id, old.title, old.content, old.language, old.url);
    INSERT INTO pages_fts(rowid, title, content, language, url)
    VALUES (new.id, new.title, new.content, new.language, new.url);
END;

INSERT OR IGNORE INTO pages (title, url, language, content) VALUES
('Fortran',    'http://web.archive.org/web/20081220110619/http://en.wikipedia.org:80/wiki/Fortran',    'en', 'Fortran'),
('Algorithm',  'http://web.archive.org/web/20081217070911/http://en.wikipedia.org:80/wiki/Algorithm',  'en', 'Algorithm'),
('MATLAB',     'http://web.archive.org/web/20090110165251/http://en.wikipedia.org:80/wiki/Matlab',     'en', 'MATLAB'),
('JavaScript', 'http://web.archive.org/web/20081218123622/http://en.wikipedia.org:80/wiki/Javascript', 'en', 'JavaScript'),
('Database',   'http://web.archive.org/web/20081219060743/http://en.wikipedia.org:80/wiki/Database',   'en', 'Database');
