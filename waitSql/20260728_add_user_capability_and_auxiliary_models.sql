-- 用户单独关闭 AI 能力、图片能力主/辅助模型拆分，以及打卡智能分析合并配置。
-- 新数据库由 server/model 建表，并由 server/source/orderfood_domain_defaults.go 写入业务默认数据。
-- 本脚本仅用于已有数据库，可重复执行；请由管理员手动执行。
-- 说明：菜谱长截图与打卡图片原主模型先保留为辅助模型，执行后请在“AI 能力配置”中为它们选择文字主模型。

-- 请在 autocommit=1 的独立会话中执行；脚本包含 DDL，MySQL 会隐式提交。
-- 不额外包裹长事务，避免中途失败后把能力配置行持续锁住。

-- 1. 用户级覆盖：只允许单独关闭，恢复后继续跟随平台总开关。
SET @ddl := IF(
    EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name = 'of_users'
          AND column_name = 'capability_disabled'
    ),
    'SELECT 1',
    'ALTER TABLE `of_users` ADD COLUMN `capability_disabled` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT ''是否单独关闭AI能力'' AFTER `checkin_count`'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 能力定义与当前配置增加辅助模型。
SET @ddl := IF(
    EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name = 'of_ai_capabilities'
          AND column_name = 'required_auxiliary_model_capability'
    ),
    'SELECT 1',
    'ALTER TABLE `of_ai_capabilities` ADD COLUMN `required_auxiliary_model_capability` varchar(32) NULL DEFAULT NULL COMMENT ''辅助模型所需能力'' AFTER `required_model_capability`'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF(
    EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = DATABASE()
          AND table_name = 'of_ai_cap_config'
          AND column_name = 'auxiliary_model_id'
    ),
    'SELECT 1',
    'ALTER TABLE `of_ai_cap_config` ADD COLUMN `auxiliary_model_id` varchar(64) NULL DEFAULT NULL COMMENT ''辅助模型ID'' AFTER `primary_model_id`'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `of_ai_capabilities`
SET `required_model_capability` = 'text',
    `required_auxiliary_model_capability` = 'vision',
    `updated_at` = NOW()
WHERE `code` IN ('recipe_image_extract', 'checkin_image_analyze');

UPDATE `of_ai_capabilities`
SET `client_feature_code` = 'checkin_image_analyze',
    `name` = '打卡智能分析',
    `sort_order` = 4,
    `updated_at` = NOW()
WHERE `code` = 'checkin_image_analyze';

UPDATE `of_ai_capabilities`
SET `sort_order` = 6, `updated_at` = NOW()
WHERE `code` = 'meal_suggest';

UPDATE `of_ai_capabilities`
SET `sort_order` = 7, `updated_at` = NOW()
WHERE `code` = 'prep_sequence';

INSERT INTO `of_ai_capabilities` (
    `code`, `name`, `client_feature_code`, `required_model_capability`,
    `required_auxiliary_model_capability`, `sort_order`, `created_at`, `updated_at`
)
SELECT
    'preference_profile_summarize', '打卡智能分析', 'checkin_image_analyze',
    'text', NULL, 5, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM `of_ai_capabilities`
    WHERE `code` = 'preference_profile_summarize'
);

UPDATE `of_ai_capabilities`
SET `name` = '打卡智能分析',
    `client_feature_code` = 'checkin_image_analyze',
    `required_model_capability` = 'text',
    `required_auxiliary_model_capability` = NULL,
    `sort_order` = 5,
    `updated_at` = NOW()
WHERE `code` = 'preference_profile_summarize';

-- 原图片模型先迁到辅助模型槽位；主模型由管理员按实际供应商能力重新选择文字模型。
UPDATE `of_ai_cap_config`
SET `auxiliary_model_id` = COALESCE(`auxiliary_model_id`, `primary_model_id`)
WHERE `capability_code` IN ('recipe_image_extract', 'checkin_image_analyze');

-- 3. 打卡智能分析是后台自动能力，从客户端入口文案中移除；其两阶段均不计费。
UPDATE `of_ai_policy`
SET `feature_labels_json` = JSON_REMOVE(
        `feature_labels_json`,
        REPLACE(
            JSON_UNQUOTE(JSON_SEARCH(`feature_labels_json`, 'one', 'taste_profile', NULL, '$[*].code')),
            '.code',
            ''
        )
    )
WHERE JSON_SEARCH(`feature_labels_json`, 'one', 'taste_profile', NULL, '$[*].code') IS NOT NULL;

UPDATE `of_ai_policy`
SET `feature_labels_json` = JSON_REMOVE(
        `feature_labels_json`,
        REPLACE(
            JSON_UNQUOTE(JSON_SEARCH(`feature_labels_json`, 'one', 'checkin_image_analyze', NULL, '$[*].code')),
            '.code',
            ''
        )
    )
WHERE JSON_SEARCH(`feature_labels_json`, 'one', 'checkin_image_analyze', NULL, '$[*].code') IS NOT NULL;

UPDATE `of_ai_cap_config`
SET `point_cost` = 0,
    `free_quota_per_day` = 0
WHERE `capability_code` = 'checkin_image_analyze';

UPDATE `of_ai_cap_config`
SET `point_cost` = 0,
    `free_quota_per_day` = 0,
    `daily_limit_per_user` = 0
WHERE `capability_code` = 'preference_profile_summarize';

-- 合并后的能力首次保存时会自动同步偏好整理阶段；先补齐该阶段的默认提示词。
INSERT INTO `of_ai_prompt_defaults` (
    `capability_code`, `version`, `system_prompt`, `user_prompt_template`,
    `allowed_variables_json`, `required_variables_json`, `output_schema_version`,
    `content_hash`, `updated_at`
)
SELECT
    'preference_profile_summarize', 1,
    '你负责根据多次结构化使用证据整理用户的长期饮食偏好。不得从单次记录推断过敏、疾病、身份或其他敏感属性；用户明确设置与系统归纳必须分开。只输出 JSON：tastePreferenceSummary 和 avoidanceOrPreferenceSummary 均为简短字符串，无法确认时为空字符串。',
    '结构化偏好证据：{{preference_evidence}}',
    JSON_ARRAY('preference_evidence'), JSON_ARRAY('preference_evidence'), 'v1',
    '6b573a0cb5f1ebcf9a04c0ccbab75ef89c7388d6357c4cad2d851a9dd887e076', NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM `of_ai_prompt_defaults`
    WHERE `capability_code` = 'preference_profile_summarize'
);

-- 4. 管理端按钮与 API 权限。沿用“查看偏好画像”的现有授权边界，不扩大到其他角色。
SET @users_menu_id := (
    SELECT `id` FROM `sys_base_menus`
    WHERE `name` = 'OrderFoodUsers' AND `deleted_at` IS NULL
    ORDER BY `id` DESC LIMIT 1
);

SET @source_button_id := (
    SELECT `id` FROM `sys_base_menu_btns`
    WHERE `sys_base_menu_id` = @users_menu_id
      AND `name` = 'orderfood:user:preference:read'
      AND `deleted_at` IS NULL
    ORDER BY `id` DESC LIMIT 1
);

INSERT INTO `sys_base_menu_btns` (
    `created_at`, `updated_at`, `name`, `desc`, `sys_base_menu_id`
)
SELECT
    NOW(), NOW(), 'orderfood:user:capability:update',
    '单独关闭或恢复用户AI能力', @users_menu_id
WHERE @users_menu_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM `sys_base_menu_btns`
      WHERE `sys_base_menu_id` = @users_menu_id
        AND `name` = 'orderfood:user:capability:update'
        AND `deleted_at` IS NULL
  );

UPDATE `sys_base_menu_btns`
SET `desc` = '单独关闭或恢复用户AI能力', `updated_at` = NOW()
WHERE `sys_base_menu_id` = @users_menu_id
  AND `name` = 'orderfood:user:capability:update'
  AND `deleted_at` IS NULL;

SET @target_button_id := (
    SELECT `id` FROM `sys_base_menu_btns`
    WHERE `sys_base_menu_id` = @users_menu_id
      AND `name` = 'orderfood:user:capability:update'
      AND `deleted_at` IS NULL
    ORDER BY `id` DESC LIMIT 1
);

INSERT INTO `sys_authority_btns` (`authority_id`, `sys_menu_id`, `sys_base_menu_btn_id`)
SELECT DISTINCT source.`authority_id`, @users_menu_id, @target_button_id
FROM `sys_authority_btns` AS source
WHERE source.`sys_menu_id` = @users_menu_id
  AND source.`sys_base_menu_btn_id` = @source_button_id
  AND @source_button_id IS NOT NULL
  AND @target_button_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM `sys_authority_btns` AS existing
      WHERE existing.`authority_id` = source.`authority_id`
        AND existing.`sys_menu_id` = @users_menu_id
        AND existing.`sys_base_menu_btn_id` = @target_button_id
  );

INSERT INTO `sys_apis` (
    `created_at`, `updated_at`, `deleted_at`, `path`, `description`, `api_group`, `method`
)
SELECT
    NOW(), NOW(), NULL, '/orderfood/users/:userId/capability',
    '单独关闭或恢复小程序用户AI能力', '来干饭-用户与积分', 'PUT'
WHERE NOT EXISTS (
    SELECT 1 FROM `sys_apis`
    WHERE `path` = '/orderfood/users/:userId/capability'
      AND `method` = 'PUT'
      AND `deleted_at` IS NULL
);

UPDATE `sys_apis`
SET `description` = '单独关闭或恢复小程序用户AI能力',
    `api_group` = '来干饭-用户与积分',
    `updated_at` = NOW()
WHERE `path` = '/orderfood/users/:userId/capability'
  AND `method` = 'PUT'
  AND `deleted_at` IS NULL;

INSERT INTO `casbin_rule` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
SELECT DISTINCT
    'p', source.`v0`, '/orderfood/users/:userId/capability', 'PUT', '', '', ''
FROM `casbin_rule` AS source
WHERE source.`ptype` = 'p'
  AND source.`v1` = '/orderfood/users/:userId/preference-profile'
  AND source.`v2` = 'GET'
  AND NOT EXISTS (
      SELECT 1 FROM `casbin_rule` AS existing
      WHERE existing.`ptype` = 'p'
        AND existing.`v0` = source.`v0`
        AND existing.`v1` = '/orderfood/users/:userId/capability'
        AND existing.`v2` = 'PUT'
  );


-- 执行后核对。管理员需要为“菜谱长截图解析”选择文字主模型；
-- “打卡智能分析”只在能力配置页出现一次，选择文字主模型和图片辅助模型后，
-- 服务端会自动同步内部偏好整理阶段，且只按每日打卡分析次数限流，不扣积分。
SELECT
    `code`, `name`, `client_feature_code`, `required_model_capability`,
    `required_auxiliary_model_capability`, `sort_order`
FROM `of_ai_capabilities`
ORDER BY `sort_order`, `code`;

SELECT
    `capability_code`, `primary_model_id`, `auxiliary_model_id`,
    `point_cost`, `free_quota_per_day`, `daily_limit_per_user`
FROM `of_ai_cap_config`
ORDER BY `capability_code`;
