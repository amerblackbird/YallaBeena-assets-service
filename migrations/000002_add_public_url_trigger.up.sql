-- Function to set public_url based on id
CREATE OR REPLACE FUNCTION set_public_url()
RETURNS TRIGGER AS $$
BEGIN
    NEW.public_url := '/assets/' || NEW.id::text;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically set public_url on insert and update
CREATE TRIGGER trigger_set_public_url
    BEFORE INSERT OR UPDATE ON assets
    FOR EACH ROW
    EXECUTE FUNCTION set_public_url();