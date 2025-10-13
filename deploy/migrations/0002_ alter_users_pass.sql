-- +goose Up
-- Добавляем колонку с временным значением по умолчанию
ALTER TABLE users 
ADD COLUMN password_hash TEXT NOT NULL DEFAULT 'temp_password_need_to_be_updated';

-- Убираем значение по умолчанию после добавления колонки
ALTER TABLE users 
ALTER COLUMN password_hash DROP DEFAULT;

-- +goose Down
ALTER TABLE users
DROP password_hash;

