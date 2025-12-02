-- 创建图片表
CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    width INTEGER DEFAULT 0,
    height INTEGER DEFAULT 0,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT,
    tags JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_images_uploader_id ON images(uploader_id);
CREATE INDEX idx_images_created_at ON images(created_at);
CREATE INDEX idx_images_deleted_at ON images(deleted_at);
CREATE INDEX idx_images_tags ON images USING GIN(tags);

-- 创建全文搜索索引（用于搜索文件名和描述）
CREATE INDEX idx_images_search ON images USING GIN(to_tsvector('english', coalesce(filename, '') || ' ' || coalesce(description, '')));

