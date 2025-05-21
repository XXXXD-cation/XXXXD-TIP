 -- 添加默认角色
INSERT INTO roles (id, name, description) VALUES
    (uuid_generate_v4(), 'Admin', '系统管理员，拥有所有权限'),
    (uuid_generate_v4(), 'Analyst', '情报分析师，负责分析和处理情报'),
    (uuid_generate_v4(), 'Viewer', '查看者，只有查看权限，无操作权限')
ON CONFLICT (name) DO NOTHING;

-- 添加默认权限
INSERT INTO permissions (id, permission_key, description, module) VALUES
    (uuid_generate_v4(), 'user.read', '查看用户信息', 'user'),
    (uuid_generate_v4(), 'user.create', '创建用户', 'user'),
    (uuid_generate_v4(), 'user.update', '更新用户信息', 'user'),
    (uuid_generate_v4(), 'user.delete', '删除用户', 'user'),
    
    (uuid_generate_v4(), 'role.read', '查看角色信息', 'role'),
    (uuid_generate_v4(), 'role.create', '创建角色', 'role'),
    (uuid_generate_v4(), 'role.update', '更新角色信息', 'role'),
    (uuid_generate_v4(), 'role.delete', '删除角色', 'role'),
    
    (uuid_generate_v4(), 'ioc.read', '查看IOC信息', 'ioc'),
    (uuid_generate_v4(), 'ioc.create', '创建IOC', 'ioc'),
    (uuid_generate_v4(), 'ioc.update', '更新IOC信息', 'ioc'),
    (uuid_generate_v4(), 'ioc.delete', '删除IOC', 'ioc'),
    
    (uuid_generate_v4(), 'case.read', '查看案例信息', 'case'),
    (uuid_generate_v4(), 'case.create', '创建案例', 'case'),
    (uuid_generate_v4(), 'case.update', '更新案例信息', 'case'),
    (uuid_generate_v4(), 'case.delete', '删除案例', 'case'),
    
    (uuid_generate_v4(), 'source.read', '查看情报源信息', 'source'),
    (uuid_generate_v4(), 'source.create', '创建情报源', 'source'),
    (uuid_generate_v4(), 'source.update', '更新情报源信息', 'source'),
    (uuid_generate_v4(), 'source.delete', '删除情报源', 'source'),
    
    (uuid_generate_v4(), 'tag.read', '查看标签信息', 'tag'),
    (uuid_generate_v4(), 'tag.create', '创建标签', 'tag'),
    (uuid_generate_v4(), 'tag.update', '更新标签信息', 'tag'),
    (uuid_generate_v4(), 'tag.delete', '删除标签', 'tag'),
    
    (uuid_generate_v4(), 'report.read', '查看报告信息', 'report'),
    (uuid_generate_v4(), 'report.create', '创建报告', 'report'),
    (uuid_generate_v4(), 'report.update', '更新报告信息', 'report'),
    (uuid_generate_v4(), 'report.delete', '删除报告', 'report'),
    
    (uuid_generate_v4(), 'config.read', '查看系统配置', 'config'),
    (uuid_generate_v4(), 'config.update', '更新系统配置', 'config')
ON CONFLICT (permission_key) DO NOTHING;

-- 为角色分配权限 (Admin角色拥有所有权限)
DO $$
DECLARE
    admin_role_id UUID;
    analyst_role_id UUID;
    viewer_role_id UUID;
    perm_id UUID;
BEGIN
    -- 获取角色ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'Admin';
    SELECT id INTO analyst_role_id FROM roles WHERE name = 'Analyst';
    SELECT id INTO viewer_role_id FROM roles WHERE name = 'Viewer';
    
    -- Admin角色分配所有权限
    FOR perm_id IN SELECT id FROM permissions LOOP
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (admin_role_id, perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END LOOP;
    
    -- Analyst角色分配分析相关权限 (不包括用户管理、角色管理、系统配置)
    FOR perm_id IN SELECT id FROM permissions WHERE module IN ('ioc', 'case', 'source', 'tag', 'report') AND permission_key NOT LIKE '%.delete' LOOP
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (analyst_role_id, perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END LOOP;
    
    -- Viewer角色只分配读取权限
    FOR perm_id IN SELECT id FROM permissions WHERE permission_key LIKE '%.read' LOOP
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (viewer_role_id, perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END LOOP;
END $$;

-- 添加默认管理员用户 (密码: admin)
INSERT INTO users (id, username, password_hash, email, full_name, role_id, is_active)
VALUES (
    uuid_generate_v4(),
    'admin',
    '$2a$10$3UvCOisgdKsmq3XdZn6XBOWnbQ8ULd5lRmfV2Lh8mPceMbcfokwYa', -- 'admin'的bcrypt哈希
    'admin@tip.local',
    '系统管理员',
    (SELECT id FROM roles WHERE name = 'Admin'),
    TRUE
)
ON CONFLICT (username) DO NOTHING;

-- 添加默认标签
INSERT INTO tags (id, name, color_hex, description) VALUES
    (uuid_generate_v4(), 'High', '#FF4136', '高风险'),
    (uuid_generate_v4(), 'Medium', '#FF851B', '中等风险'),
    (uuid_generate_v4(), 'Low', '#FFDC00', '低风险'),
    (uuid_generate_v4(), 'False Positive', '#2ECC40', '误报'),
    (uuid_generate_v4(), 'Verified', '#0074D9', '已验证'),
    (uuid_generate_v4(), 'Critical', '#B10DC9', '关键资产相关'),
    (uuid_generate_v4(), 'Investigation Needed', '#AAAAAA', '需要进一步调查')
ON CONFLICT (name) DO NOTHING;

-- 添加默认系统配置
INSERT INTO system_configurations (config_key, config_value, description, is_encrypted)
VALUES
    ('default_threat_score_threshold', '65', '默认威胁评分阈值，高于此值的IOC被视为高风险', FALSE),
    ('default_ioc_expiry_days', '90', '默认IOC过期天数', FALSE),
    ('enable_external_api', 'true', '是否启用外部API', FALSE),
    ('enable_auto_enrichment', 'true', '是否启用自动富化', FALSE),
    ('notification_email_sender', 'noreply@tip.local', '发送通知邮件的地址', FALSE)
ON CONFLICT (config_key) DO NOTHING;