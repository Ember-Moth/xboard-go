-- 创建用户组表
CREATE TABLE IF NOT EXISTS v2_user_group (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL COMMENT '组名称',
    description TEXT COMMENT '组描述',
    server_ids JSON COMMENT '该组可访问的节点ID列表',
    plan_ids JSON COMMENT '该组可购买的套餐ID列表',
    default_transfer_enable BIGINT DEFAULT 0 COMMENT '默认流量（字节）',
    default_speed_limit INT COMMENT '默认速度限制（Mbps）',
    default_device_limit INT COMMENT '默认设备数限制',
    sort INT DEFAULT 0 COMMENT '排序',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
) COMMENT='用户权限组（核心表）';

-- 插入默认组
INSERT INTO v2_user_group (id, name, description, server_ids, plan_ids, default_transfer_enable, sort, created_at, updated_at) VALUES
(1, '试用用户', '新注册用户默认组，流量较少', '[]', '[]', 1073741824, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, '普通用户', '购买基础套餐的用户', '[]', '[]', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 'VIP用户', '购买高级套餐的用户', '[]', '[]', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 修改用户表：确保 group_id 有默认值
ALTER TABLE v2_user MODIFY COLUMN group_id BIGINT DEFAULT 1 COMMENT '用户所属权限组ID（关联v2_user_group）';

-- 为现有用户设置默认组
UPDATE v2_user SET group_id = 1 WHERE group_id IS NULL OR group_id = 0;

-- 修改套餐表：添加 upgrade_group_id
ALTER TABLE v2_plan ADD COLUMN IF NOT EXISTS upgrade_group_id BIGINT COMMENT '购买后升级到的用户组ID（可选）';
