-- 删除短信验证码表
DROP TABLE IF EXISTS sms_codes;

-- 删除手机号字段和索引
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone;

