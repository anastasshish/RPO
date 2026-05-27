-- +goose Up
CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, login TEXT NOT NULL UNIQUE, name TEXT, password_hash TEXT NOT NULL, is_admin BOOLEAN DEFAULT 0, created_at DATETIME, updated_at DATETIME);
CREATE TABLE IF NOT EXISTS terminals (id INTEGER PRIMARY KEY AUTOINCREMENT, serial_number TEXT NOT NULL UNIQUE, name TEXT, address TEXT, description TEXT, created_at DATETIME, updated_at DATETIME);
CREATE TABLE IF NOT EXISTS card_keys (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, key_value TEXT NOT NULL, description TEXT, created_at DATETIME, updated_at DATETIME);
CREATE TABLE IF NOT EXISTS cards (id INTEGER PRIMARY KEY AUTOINCREMENT, number TEXT NOT NULL UNIQUE, balance INTEGER DEFAULT 0, blocked BOOLEAN DEFAULT 0, owner_name TEXT, user_id INTEGER, key_id INTEGER, created_at DATETIME, updated_at DATETIME, FOREIGN KEY(key_id) REFERENCES card_keys(id), FOREIGN KEY(user_id) REFERENCES users(id));
CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);
CREATE TABLE IF NOT EXISTS transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, amount INTEGER NOT NULL, card_id INTEGER, terminal_id INTEGER, status TEXT, message TEXT, created_at DATETIME, FOREIGN KEY(card_id) REFERENCES cards(id), FOREIGN KEY(terminal_id) REFERENCES terminals(id));
-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS card_keys;
DROP TABLE IF EXISTS terminals;
DROP TABLE IF EXISTS users;
