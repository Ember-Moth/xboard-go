-- 回滚用户组表创建
DROP TABLE IF EXISTS v2_user_group;

-- 回滚套餐表修改
ALTER TABLE v2_plan DROP COLUMN IF EXISTS upgrade_group_id;

-- 回滚用户表修改（恢复为可空）
ALTER TABLE v2_user MODIFY COLUMN group_id BIGINT COMMENT '用户所属权限组ID';
