-- 添加用户签到字段
ALTER TABLE v2_user ADD COLUMN IF NOT EXISTS last_checkin_at BIGINT DEFAULT NULL;

-- 添加用户注册 IP 字段
ALTER TABLE v2_user ADD COLUMN IF NOT EXISTS register_ip VARCHAR(45) DEFAULT NULL;

-- 添加站点设置
INSERT INTO v2_setting (`key`, `value`) VALUES 
('site_name', 'XBoard'),
('site_logo', ''),
('site_description', ''),
('site_keywords', ''),
('site_theme', 'default'),
('site_primary_color', '#6366f1'),
('site_favicon', ''),
('site_footer', ''),
('site_tos', ''),
('site_privacy', ''),
('payment_currency', 'CNY'),
('payment_symbol', '¥'),
('register_enable', '1'),
('register_invite_only', '0'),
('register_trial', '0'),
('register_trial_days', '1'),
('register_trial_traffic', '10'),
('register_ip_limit', '0'),
('mail_verify', '0'),
('telegram_enable', '0'),
('telegram_bot_token', ''),
('telegram_chat_id', '')
ON DUPLICATE KEY UPDATE `key` = `key`;
