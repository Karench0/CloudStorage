DROP TRIGGER IF EXISTS trigger_update_dir_size ON files;
DROP FUNCTION IF EXISTS update_directory_size();
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS directories;
DROP TABLE IF EXISTS users;