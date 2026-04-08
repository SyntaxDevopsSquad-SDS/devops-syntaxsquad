-- Migration 001: Add id to pages, language index and FTS5 full-text search

-- 1. Rename old table
ALTER TABLE pages RENAME TO pages_old;

-- 2. Create new pages table with id as PK
CREATE TABLE pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL DEFAULT 'en' CHECK(language IN ('en', 'da')),
    last_updated TIMESTAMP,
    content TEXT NOT NULL
);

-- 3. Copy data
INSERT INTO pages (title, url, language, last_updated, content)
SELECT title, url, language, last_updated, content FROM pages_old;

-- 4. Drop old table
DROP TABLE pages_old;

-- 5. Index on language
CREATE INDEX IF NOT EXISTS idx_pages_language ON pages(language);

-- 6. FTS5 virtual table
CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(
    title,
    content,
    language UNINDEXED,
    url UNINDEXED,
    content='pages',
    content_rowid='id'
);

-- 7. Populate FTS5 with existing data
INSERT INTO pages_fts(rowid, title, content, language, url)
SELECT id, title, content, language, url FROM pages;

-- 8. Triggers to keep FTS5 in sync (Trigger for future changes)
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
