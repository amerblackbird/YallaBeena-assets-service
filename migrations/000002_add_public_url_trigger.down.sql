-- Drop the trigger
DROP TRIGGER IF EXISTS trigger_set_public_url ON assets;

-- Drop the function
DROP FUNCTION IF EXISTS set_public_url();