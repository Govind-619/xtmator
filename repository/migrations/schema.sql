-- Users table
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT,                          -- nullable for Google-only accounts
    google_id     TEXT UNIQUE,                   -- Google subject ID
    auth_provider TEXT NOT NULL DEFAULT 'email', -- 'email' | 'google'
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);


-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    client_name TEXT NOT NULL DEFAULT '',
    location    TEXT NOT NULL DEFAULT '',
    share_token TEXT UNIQUE,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- DSR items catalogue (populated by import-dsr tool)
CREATE TABLE IF NOT EXISTS dsr_items (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    category    TEXT NOT NULL,
    code        TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    unit        TEXT NOT NULL DEFAULT 'CUM',
    rate        REAL NOT NULL DEFAULT 0
);

-- Project Sheets table
CREATE TABLE IF NOT EXISTS project_sheets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL DEFAULT 'Main',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- BOQ entries per project sheet
CREATE TABLE IF NOT EXISTS boq_entries (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id  INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sheet_id    INTEGER REFERENCES project_sheets(id) ON DELETE CASCADE,
    item_no     INTEGER NOT NULL DEFAULT 1,
    dsr_item_id INTEGER REFERENCES dsr_items(id),
    description TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    length      REAL NOT NULL DEFAULT 0,
    breadth     REAL NOT NULL DEFAULT 0,
    height      REAL NOT NULL DEFAULT 0,
    quantity    REAL NOT NULL DEFAULT 0,
    unit        TEXT NOT NULL DEFAULT 'CUM',
    rate        REAL NOT NULL DEFAULT 0,
    amount      REAL NOT NULL DEFAULT 0
);

-- Custom saved items library
CREATE TABLE IF NOT EXISTS custom_items (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category    TEXT NOT NULL,
    description TEXT NOT NULL,
    unit        TEXT NOT NULL DEFAULT 'CUM',
    rate        REAL NOT NULL DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
