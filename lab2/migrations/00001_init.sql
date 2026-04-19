CREATE TABLE IF NOT EXISTS terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    serial_number TEXT NOT NULL UNIQUE,
    address TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS card_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_name TEXT NOT NULL,
    key_value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_admin INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_number TEXT NOT NULL UNIQUE,
    balance REAL NOT NULL DEFAULT 0,
    blocked INTEGER NOT NULL DEFAULT 0,
    owner_name TEXT NOT NULL,
    key_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    FOREIGN KEY(key_id) REFERENCES card_keys(id)
);

CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    card_number TEXT NOT NULL,
    terminal_id INTEGER NOT NULL,
    approved INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(card_number) REFERENCES cards(card_number),
    FOREIGN KEY(terminal_id) REFERENCES terminals(id)
);

INSERT OR IGNORE INTO users(login, password_hash, is_admin)
VALUES ('admin', '$2a$10$FqMweFNuPDE0cePwW7JgtuQnHfT8f0VAAxV2n8BmlqK4wRklS1z/u', 1);

-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS terminals;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS card_keys;
