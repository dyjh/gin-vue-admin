-- 将 AI 供应商凭据从环境变量引用切换为数据库明文 API Key。
-- 重要：旧 credential_ref 只记录 env://变量名，数据库无法据此取得真实密钥。
-- 执行后原供应商会暂时停用并清空连接测试结果；请在后台重新录入 API Key，
-- 完成连接测试后再启用。credential_ref 会保留为可空的旧配置核对/回退字段，新代码不再读取。
--
-- MySQL 的 ALTER TABLE 会隐式提交，因此本脚本不使用事务包装 DDL。

ALTER TABLE `of_ai_providers`
    MODIFY COLUMN `credential_ref` TEXT NULL
        COMMENT '旧环境变量凭据引用（迁移核对/回退用）',
    ADD COLUMN `api_key` TEXT NULL
        COMMENT 'API密钥（明文）'
        AFTER `base_url`;

UPDATE `of_ai_providers`
SET
    `api_key` = '',
    `enabled` = 0,
    `last_test_success` = NULL,
    `last_test_category` = NULL,
    `last_test_duration_ms` = NULL,
    `last_test_safe_message` = NULL,
    `last_tested_at` = NULL,
    `last_change_reason` = '凭据存储已切换为数据库 API Key，等待重新录入',
    `version` = `version` + 1,
    `updated_at` = NOW();

ALTER TABLE `of_ai_providers`
    MODIFY COLUMN `api_key` TEXT NOT NULL
        COMMENT 'API密钥（明文）';

UPDATE `sys_base_menu_btns`
SET `desc` = '更新供应商 API Key'
WHERE `name` = 'orderfood:provider:credential-write';