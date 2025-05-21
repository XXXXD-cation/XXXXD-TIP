 -- 启用UUID生成模块
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    role_id UUID NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMPTZ
);

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    permission_key VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    module VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 创建情报源配置表
CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    url_or_endpoint TEXT,
    api_key_encrypted TEXT,
    pull_frequency_seconds INTEGER DEFAULT 3600,
    reliability_score SMALLINT DEFAULT 50 CHECK (reliability_score >= 0 AND reliability_score <= 100),
    parser_plugin_name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_polled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建标签表
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    color_hex VARCHAR(7) DEFAULT '#CCCCCC',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建IOC标签关联表
CREATE TABLE IF NOT EXISTS ioc_tag_relations (
    ioc_value TEXT NOT NULL,
    ioc_type VARCHAR(50) NOT NULL,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (ioc_value, tag_id)
);

-- 创建攻击组织表
CREATE TABLE IF NOT EXISTS threat_actors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    aliases TEXT[],
    description TEXT,
    target_sectors TEXT[],
    origin_country VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建恶意软件表
CREATE TABLE IF NOT EXISTS malware (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    family VARCHAR(255),
    aliases TEXT[],
    type VARCHAR(100),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建漏洞信息表
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cve_id VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    cvss_score_v3 DECIMAL(3,1),
    affected_products TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建IOC与实体关联表
CREATE TABLE IF NOT EXISTS ioc_entity_relations (
    ioc_value TEXT NOT NULL,
    ioc_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    relation_type VARCHAR(100),
    description TEXT,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    source_ref TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ioc_value, entity_id, entity_type, relation_type)
);

-- 创建情报案例表
CREATE TABLE IF NOT EXISTS cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    assignee_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    creator_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建案例IOC关联表
CREATE TABLE IF NOT EXISTS case_ioc_relations (
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    ioc_value TEXT NOT NULL,
    ioc_type VARCHAR(50) NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (case_id, ioc_value)
);

-- 创建案例笔记表
CREATE TABLE IF NOT EXISTS case_notes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    note_content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建案例附件表
CREATE TABLE IF NOT EXISTS case_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path_or_id TEXT NOT NULL,
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    uploader_user_id UUID NOT NULL REFERENCES users(id),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建情报报告元数据表
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    format VARCHAR(20) NOT NULL,
    content_stix JSONB,
    content_html TEXT,
    file_path_or_id TEXT,
    creator_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统配置表
CREATE TABLE IF NOT EXISTS system_configurations (
    config_key VARCHAR(255) PRIMARY KEY,
    config_value TEXT NOT NULL,
    description TEXT,
    is_encrypted BOOLEAN DEFAULT FALSE,
    last_updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles (name);

CREATE INDEX IF NOT EXISTS idx_permissions_permission_key ON permissions (permission_key);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions (role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions (permission_id);

CREATE INDEX IF NOT EXISTS idx_sources_name ON sources (name);
CREATE INDEX IF NOT EXISTS idx_sources_status ON sources (status);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags (name);

CREATE INDEX IF NOT EXISTS idx_ioc_tag_relations_ioc_value ON ioc_tag_relations (ioc_value);
CREATE INDEX IF NOT EXISTS idx_ioc_tag_relations_tag_id ON ioc_tag_relations (tag_id);

CREATE INDEX IF NOT EXISTS idx_threat_actors_name ON threat_actors (name);
CREATE INDEX IF NOT EXISTS gin_idx_threat_actors_aliases ON threat_actors USING GIN (aliases);
CREATE INDEX IF NOT EXISTS gin_idx_threat_actors_target_sectors ON threat_actors USING GIN (target_sectors);

CREATE INDEX IF NOT EXISTS idx_malware_name ON malware (name);
CREATE INDEX IF NOT EXISTS idx_malware_family ON malware (family);
CREATE INDEX IF NOT EXISTS gin_idx_malware_aliases ON malware USING GIN (aliases);

CREATE INDEX IF NOT EXISTS idx_vulnerabilities_cve_id ON vulnerabilities (cve_id);

CREATE INDEX IF NOT EXISTS idx_ioc_entity_relations_ioc ON ioc_entity_relations (ioc_value, ioc_type);
CREATE INDEX IF NOT EXISTS idx_ioc_entity_relations_entity ON ioc_entity_relations (entity_id, entity_type);

CREATE INDEX IF NOT EXISTS idx_cases_title ON cases (title);
CREATE INDEX IF NOT EXISTS idx_cases_status ON cases (status);
CREATE INDEX IF NOT EXISTS idx_cases_priority ON cases (priority);
CREATE INDEX IF NOT EXISTS idx_cases_assignee_user_id ON cases (assignee_user_id);
CREATE INDEX IF NOT EXISTS idx_cases_creator_user_id ON cases (creator_user_id);

CREATE INDEX IF NOT EXISTS idx_case_ioc_relations_case_id ON case_ioc_relations (case_id);
CREATE INDEX IF NOT EXISTS idx_case_ioc_relations_ioc_value ON case_ioc_relations (ioc_value);

CREATE INDEX IF NOT EXISTS idx_case_notes_case_id ON case_notes (case_id);
CREATE INDEX IF NOT EXISTS idx_case_notes_user_id ON case_notes (user_id);

CREATE INDEX IF NOT EXISTS idx_case_attachments_case_id ON case_attachments (case_id);

CREATE INDEX IF NOT EXISTS idx_reports_title ON reports (title);
CREATE INDEX IF NOT EXISTS idx_reports_creator_user_id ON reports (creator_user_id);