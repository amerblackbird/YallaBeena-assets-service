-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_set_public_url ON assets;
DROP TRIGGER IF EXISTS trigger_update_updated_at ON assets;

-- Drop functions
DROP FUNCTION IF EXISTS set_public_url();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table
DROP TABLE IF EXISTS assets;