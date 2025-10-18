-- Create assets table with all fields and enhancements
CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url VARCHAR(500) NOT NULL,
    public_url VARCHAR(500),
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    metadata JSONB DEFAULT '{}',
    secure BOOLEAN DEFAULT FALSE,
    storage_key VARCHAR(500),
    storage_provider VARCHAR(100),
    resource_id VARCHAR(255),
    resource_type VARCHAR(100),
    content_type VARCHAR(100) NOT NULL,
    user_id VARCHAR(255),
    access_level VARCHAR(50) DEFAULT 'public',
    allowed_roles TEXT[],
    is_encrypted BOOLEAN DEFAULT FALSE,
    encryption_key VARCHAR(255),
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    tags TEXT[],
    file_hash VARCHAR(255),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active BOOLEAN DEFAULT TRUE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_assets_user_id ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_content_type ON assets(content_type);
CREATE INDEX IF NOT EXISTS idx_assets_created_at ON assets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assets_active ON assets(active);
CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_assets_resource ON assets(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_assets_file_hash ON assets(file_hash);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_assets_user_active ON assets(user_id, active);
CREATE INDEX IF NOT EXISTS idx_assets_user_active_deleted ON assets(user_id, active, deleted_at);

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

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on record updates
CREATE TRIGGER trigger_update_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();