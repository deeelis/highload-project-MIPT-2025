CREATE TABLE content
(
    id         VARCHAR(36) PRIMARY KEY,
    type       VARCHAR(10) NOT NULL,
    status     VARCHAR(20) NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMP   NOT NULL,
    updated_at TIMESTAMP   NOT NULL
);

CREATE TABLE text_content
(
    content_id    VARCHAR(36) PRIMARY KEY REFERENCES content (id),
    original_text TEXT NOT NULL
);

CREATE TABLE image_content
(
    content_id VARCHAR(36) PRIMARY KEY REFERENCES content (id),
    s3_key     VARCHAR(255) NOT NULL
);

-- CREATE INDEX idx_content_status ON content (status);
-- CREATE INDEX idx_content_type ON content (type);
-- CREATE INDEX idx_content_created_at ON content (created_at);