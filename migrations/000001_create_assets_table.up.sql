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
    content_type VARCHAR(100) NOT NULL, -- Moved to the top for better visibility, ex image/png, application/pdf, etc.
    user_id VARCHAR(255),
    access_level VARCHAR(50) DEFAULT 'public',
    allowed_roles TEXT[],
    is_encrypted BOOLEAN DEFAULT FALSE,
    encryption_key VARCHAR(255),
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    tags TEXT[],
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

-- Create composite index for user queries
CREATE INDEX IF NOT EXISTS idx_assets_user_active ON assets(user_id, active);