-- 1. Создание таблицы пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Создание таблицы директорий
CREATE TABLE directories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES directories(id) ON DELETE CASCADE,
    size BIGINT DEFAULT 0, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Создание таблицы файлов
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    path VARCHAR(512) NOT NULL, 
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    directory_id INTEGER REFERENCES directories(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =========================================================================
-- АВТОМАТИЧЕСКИЙ СЧЕТ РАЗМЕРА ДИРЕКТОРИЙ 
-- =========================================================================
CREATE OR REPLACE FUNCTION update_directory_size()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        IF NEW.directory_id IS NOT NULL THEN
            UPDATE directories 
            SET size = size + COALESCE(NEW.size, 0) - CASE WHEN TG_OP = 'UPDATE' THEN COALESCE(OLD.size, 0) ELSE 0 END
            WHERE id = NEW.directory_id;
        END IF;
    END IF;

    IF (TG_OP = 'DELETE' OR TG_OP = 'UPDATE') THEN
        IF OLD.directory_id IS NOT NULL THEN
            UPDATE directories 
            SET size = size - COALESCE(OLD.size, 0)
            WHERE id = OLD.directory_id;
        END IF;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_dir_size
AFTER INSERT OR UPDATE OR DELETE ON files
FOR EACH ROW
EXECUTE FUNCTION update_directory_size();