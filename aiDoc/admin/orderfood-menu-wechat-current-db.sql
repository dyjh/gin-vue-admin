-- 来干饭管理端现有 MySQL 数据库调整脚本
-- 适用范围：
-- 1. 去掉“来干饭”和“工作台”包装层；
-- 2. 使用“数据概览”替代 GVA 原仪表盘，并提升为一级菜单；
-- 3. 将六个业务分组提升为一级菜单；
-- 4. 新增一级“微信配置”菜单、按钮权限、API 和 Casbin 权限；
-- 5. 创建当前微信配置表。
--
-- 执行前请备份数据库。脚本不写入 AppSecret；部署新代码后由超级管理员在
-- “微信配置”页面首次保存，服务端会使用 config.yaml 的
-- orderfood.identity-key 加密保存。

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `of_wx_config` (
  `singleton_key` varchar(32) NOT NULL COMMENT '单例记录键',
  `version` bigint NOT NULL DEFAULT 1 COMMENT '配置版本',
  `app_id` varchar(64) NOT NULL COMMENT '微信小程序AppID',
  `app_secret_encrypted` text NOT NULL COMMENT '加密后的微信小程序AppSecret',
  `secret_updated_at` datetime DEFAULT NULL COMMENT 'AppSecret更新时间',
  `applied_by_id` bigint unsigned NOT NULL COMMENT '应用管理员ID',
  `applied_by_username` varchar(80) NOT NULL COMMENT '应用管理员用户名',
  `applied_by_nickname` varchar(80) DEFAULT NULL COMMENT '应用管理员昵称',
  `applied_at` datetime NOT NULL COMMENT '应用时间',
  `reason` varchar(200) NOT NULL COMMENT '修改原因',
  PRIMARY KEY (`singleton_key`),
  KEY `idx_of_wx_config_applied_at` (`applied_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信小程序当前配置';

-- 数据概览不存在时先补建，再统一写入最终一级菜单属性。
INSERT INTO `sys_base_menus`
  (`created_at`, `updated_at`, `deleted_at`, `menu_level`, `parent_id`, `path`, `name`,
   `hidden`, `component`, `sort`, `active_name`, `keep_alive`, `default_menu`, `title`,
   `icon`, `close_tab`, `transition_type`)
SELECT
  NOW(), NOW(), NULL, 0, 0, 'dashboard', 'OrderFoodDashboard',
  0, 'view/orderFood/dashboard/index.vue', 1, '', 1, 0, '数据概览',
  'data-analysis', 0, ''
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_base_menus` WHERE `name` = 'OrderFoodDashboard'
);

UPDATE `sys_base_menus`
SET
  `updated_at` = NOW(),
  `deleted_at` = NULL,
  `menu_level` = 0,
  `parent_id` = 0,
  `path` = 'dashboard',
  `hidden` = 0,
  `component` = 'view/orderFood/dashboard/index.vue',
  `sort` = 1,
  `keep_alive` = 1,
  `title` = '数据概览',
  `icon` = 'data-analysis'
WHERE `name` = 'OrderFoodDashboard';

-- 六个业务分组直接提升到后台一级。
UPDATE `sys_base_menus`
SET
  `updated_at` = NOW(),
  `deleted_at` = NULL,
  `menu_level` = 0,
  `parent_id` = 0,
  `path` = CASE `name`
    WHEN 'OrderFoodUsersAndPoints' THEN 'users-and-points'
    WHEN 'OrderFoodDishOperations' THEN 'dish-operations'
    WHEN 'OrderFoodContentSafety' THEN 'content-safety'
    WHEN 'OrderFoodMealManagement' THEN 'meal-management'
    WHEN 'OrderFoodAiCapabilities' THEN 'ai-capabilities'
    WHEN 'OrderFoodMessageCenter' THEN 'message-center'
  END,
  `hidden` = 0,
  `component` = 'view/routerHolder.vue',
  `sort` = CASE `name`
    WHEN 'OrderFoodUsersAndPoints' THEN 2
    WHEN 'OrderFoodDishOperations' THEN 3
    WHEN 'OrderFoodContentSafety' THEN 4
    WHEN 'OrderFoodMealManagement' THEN 5
    WHEN 'OrderFoodAiCapabilities' THEN 6
    WHEN 'OrderFoodMessageCenter' THEN 7
  END,
  `keep_alive` = 1,
  `title` = CASE `name`
    WHEN 'OrderFoodUsersAndPoints' THEN '用户与积分'
    WHEN 'OrderFoodDishOperations' THEN '菜品运营'
    WHEN 'OrderFoodContentSafety' THEN '内容安全'
    WHEN 'OrderFoodMealManagement' THEN '饭局管理'
    WHEN 'OrderFoodAiCapabilities' THEN 'AI 能力'
    WHEN 'OrderFoodMessageCenter' THEN '消息中心'
  END,
  `icon` = CASE `name`
    WHEN 'OrderFoodUsersAndPoints' THEN 'user'
    WHEN 'OrderFoodDishOperations' THEN 'dish'
    WHEN 'OrderFoodContentSafety' THEN 'lock'
    WHEN 'OrderFoodMealManagement' THEN 'calendar'
    WHEN 'OrderFoodAiCapabilities' THEN 'cpu'
    WHEN 'OrderFoodMessageCenter' THEN 'message'
  END
WHERE `name` IN (
  'OrderFoodUsersAndPoints',
  'OrderFoodDishOperations',
  'OrderFoodContentSafety',
  'OrderFoodMealManagement',
  'OrderFoodAiCapabilities',
  'OrderFoodMessageCenter'
);

-- 分组下的业务页面相应调整为第二层。
UPDATE `sys_base_menus` AS child
INNER JOIN `sys_base_menus` AS parent ON parent.`id` = child.`parent_id`
SET child.`menu_level` = 1, child.`updated_at` = NOW()
WHERE parent.`name` IN (
  'OrderFoodUsersAndPoints',
  'OrderFoodDishOperations',
  'OrderFoodContentSafety',
  'OrderFoodMealManagement',
  'OrderFoodAiCapabilities',
  'OrderFoodMessageCenter'
);

-- 新增一级微信配置菜单。
INSERT INTO `sys_base_menus`
  (`created_at`, `updated_at`, `deleted_at`, `menu_level`, `parent_id`, `path`, `name`,
   `hidden`, `component`, `sort`, `active_name`, `keep_alive`, `default_menu`, `title`,
   `icon`, `close_tab`, `transition_type`)
SELECT
  NOW(), NOW(), NULL, 0, 0, 'wechat-config', 'OrderFoodWeChatConfig',
  0, 'view/orderFood/wechatConfig/index.vue', 8, '', 1, 0, '微信配置',
  'setting', 0, ''
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_base_menus` WHERE `name` = 'OrderFoodWeChatConfig'
);

UPDATE `sys_base_menus`
SET
  `updated_at` = NOW(),
  `deleted_at` = NULL,
  `menu_level` = 0,
  `parent_id` = 0,
  `path` = 'wechat-config',
  `hidden` = 0,
  `component` = 'view/orderFood/wechatConfig/index.vue',
  `sort` = 8,
  `keep_alive` = 1,
  `title` = '微信配置',
  `icon` = 'setting'
WHERE `name` = 'OrderFoodWeChatConfig';

SET @orderfood_dashboard_menu_id := (
  SELECT `id` FROM `sys_base_menus` WHERE `name` = 'OrderFoodDashboard' LIMIT 1
);
SET @orderfood_wechat_menu_id := (
  SELECT `id` FROM `sys_base_menus` WHERE `name` = 'OrderFoodWeChatConfig' LIMIT 1
);

-- 首个管理员、来干饭超管和运营管理员均可看到数据概览。
INSERT INTO `sys_authority_menus` (`sys_base_menu_id`, `sys_authority_authority_id`)
SELECT CAST(@orderfood_dashboard_menu_id AS CHAR), CAST(role_ids.`authority_id` AS CHAR)
FROM (
  SELECT 888 AS `authority_id`
  UNION ALL SELECT 9901
  UNION ALL SELECT 9902
) AS role_ids
INNER JOIN `sys_authorities` AS authority
  ON authority.`authority_id` = role_ids.`authority_id`
WHERE @orderfood_dashboard_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM `sys_authority_menus` AS existing
    WHERE CAST(existing.`sys_base_menu_id` AS UNSIGNED) = @orderfood_dashboard_menu_id
      AND CAST(existing.`sys_authority_authority_id` AS UNSIGNED) = role_ids.`authority_id`
  );

-- 微信配置默认只开放给平台首个管理员和来干饭超级管理员。
INSERT INTO `sys_authority_menus` (`sys_base_menu_id`, `sys_authority_authority_id`)
SELECT CAST(@orderfood_wechat_menu_id AS CHAR), CAST(role_ids.`authority_id` AS CHAR)
FROM (
  SELECT 888 AS `authority_id`
  UNION ALL SELECT 9901
) AS role_ids
INNER JOIN `sys_authorities` AS authority
  ON authority.`authority_id` = role_ids.`authority_id`
WHERE @orderfood_wechat_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM `sys_authority_menus` AS existing
    WHERE CAST(existing.`sys_base_menu_id` AS UNSIGNED) = @orderfood_wechat_menu_id
      AND CAST(existing.`sys_authority_authority_id` AS UNSIGNED) = role_ids.`authority_id`
  );

-- 为微信配置页创建读取和保存按钮权限。
INSERT INTO `sys_base_menu_btns`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `desc`, `sys_base_menu_id`)
SELECT
  NOW(), NOW(), NULL, 'orderfood:wechat-config:read', '查看权限', @orderfood_wechat_menu_id
WHERE @orderfood_wechat_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_base_menu_btns`
    WHERE `sys_base_menu_id` = @orderfood_wechat_menu_id
      AND `name` = 'orderfood:wechat-config:read'
  );

INSERT INTO `sys_base_menu_btns`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `desc`, `sys_base_menu_id`)
SELECT
  NOW(), NOW(), NULL, 'orderfood:wechat-config:update', '保存微信小程序配置并立即生效', @orderfood_wechat_menu_id
WHERE @orderfood_wechat_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_base_menu_btns`
    WHERE `sys_base_menu_id` = @orderfood_wechat_menu_id
      AND `name` = 'orderfood:wechat-config:update'
  );

UPDATE `sys_base_menu_btns`
SET `updated_at` = NOW(), `deleted_at` = NULL, `desc` = '查看权限'
WHERE `sys_base_menu_id` = @orderfood_wechat_menu_id
  AND `name` = 'orderfood:wechat-config:read';

UPDATE `sys_base_menu_btns`
SET `updated_at` = NOW(), `deleted_at` = NULL, `desc` = '保存微信小程序配置并立即生效'
WHERE `sys_base_menu_id` = @orderfood_wechat_menu_id
  AND `name` = 'orderfood:wechat-config:update';

INSERT INTO `sys_authority_btns`
  (`authority_id`, `sys_menu_id`, `sys_base_menu_btn_id`)
SELECT role_ids.`authority_id`, @orderfood_wechat_menu_id, button.`id`
FROM (
  SELECT 888 AS `authority_id`
  UNION ALL SELECT 9901
) AS role_ids
INNER JOIN `sys_authorities` AS authority
  ON authority.`authority_id` = role_ids.`authority_id`
INNER JOIN `sys_base_menu_btns` AS button
  ON button.`sys_base_menu_id` = @orderfood_wechat_menu_id
 AND button.`name` IN (
   'orderfood:wechat-config:read',
   'orderfood:wechat-config:update'
 )
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_authority_btns` AS existing
  WHERE existing.`authority_id` = role_ids.`authority_id`
    AND existing.`sys_menu_id` = @orderfood_wechat_menu_id
    AND existing.`sys_base_menu_btn_id` = button.`id`
);

-- 注入 GVA API 记录。
INSERT INTO `sys_apis`
  (`created_at`, `updated_at`, `deleted_at`, `path`, `description`, `api_group`, `method`)
SELECT
  NOW(), NOW(), NULL, '/orderfood/wechat-config', '获取微信小程序配置', '来干饭-微信配置', 'GET'
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_apis`
  WHERE `path` = '/orderfood/wechat-config' AND `method` = 'GET'
);

INSERT INTO `sys_apis`
  (`created_at`, `updated_at`, `deleted_at`, `path`, `description`, `api_group`, `method`)
SELECT
  NOW(), NOW(), NULL, '/orderfood/wechat-config', '保存微信小程序配置并立即生效', '来干饭-微信配置', 'PUT'
WHERE NOT EXISTS (
  SELECT 1 FROM `sys_apis`
  WHERE `path` = '/orderfood/wechat-config' AND `method` = 'PUT'
);

UPDATE `sys_apis`
SET
  `updated_at` = NOW(),
  `deleted_at` = NULL,
  `description` = CASE `method`
    WHEN 'GET' THEN '获取微信小程序配置'
    WHEN 'PUT' THEN '保存微信小程序配置并立即生效'
  END,
  `api_group` = '来干饭-微信配置'
WHERE `path` = '/orderfood/wechat-config' AND `method` IN ('GET', 'PUT');

-- 注入平台首个管理员和来干饭超级管理员的 Casbin 路由权限。
INSERT INTO `casbin_rule` (`ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`)
SELECT 'p', CAST(role_ids.`authority_id` AS CHAR), '/orderfood/wechat-config', methods.`method`, '', '', ''
FROM (
  SELECT 888 AS `authority_id`
  UNION ALL SELECT 9901
) AS role_ids
CROSS JOIN (
  SELECT 'GET' AS `method`
  UNION ALL SELECT 'PUT'
) AS methods
INNER JOIN `sys_authorities` AS authority
  ON authority.`authority_id` = role_ids.`authority_id`
WHERE NOT EXISTS (
  SELECT 1 FROM `casbin_rule` AS existing
  WHERE existing.`ptype` = 'p'
    AND CAST(existing.`v0` AS UNSIGNED) = role_ids.`authority_id`
    AND existing.`v1` = '/orderfood/wechat-config'
    AND existing.`v2` = methods.`method`
);

-- 原 GVA 仪表盘不再初始化或授权。
DELETE authority_btn
FROM `sys_authority_btns` AS authority_btn
INNER JOIN `sys_base_menu_btns` AS button
  ON button.`id` = authority_btn.`sys_base_menu_btn_id`
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = button.`sys_base_menu_id`
WHERE menu.`name` = 'dashboard';

DELETE parameter
FROM `sys_base_menu_parameters` AS parameter
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = parameter.`sys_base_menu_id`
WHERE menu.`name` = 'dashboard';

DELETE button
FROM `sys_base_menu_btns` AS button
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = button.`sys_base_menu_id`
WHERE menu.`name` = 'dashboard';

DELETE authority_menu
FROM `sys_authority_menus` AS authority_menu
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = CAST(authority_menu.`sys_base_menu_id` AS UNSIGNED)
WHERE menu.`name` = 'dashboard';

DELETE FROM `sys_base_menus` WHERE `name` = 'dashboard';

-- 所有业务子菜单已重挂载，随后清理旧的包装菜单及其授权。
DELETE authority_btn
FROM `sys_authority_btns` AS authority_btn
INNER JOIN `sys_base_menu_btns` AS button
  ON button.`id` = authority_btn.`sys_base_menu_btn_id`
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = button.`sys_base_menu_id`
WHERE menu.`name` IN ('OrderFood', 'OrderFoodWorkspace');

DELETE parameter
FROM `sys_base_menu_parameters` AS parameter
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = parameter.`sys_base_menu_id`
WHERE menu.`name` IN ('OrderFood', 'OrderFoodWorkspace');

DELETE button
FROM `sys_base_menu_btns` AS button
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = button.`sys_base_menu_id`
WHERE menu.`name` IN ('OrderFood', 'OrderFoodWorkspace');

DELETE authority_menu
FROM `sys_authority_menus` AS authority_menu
INNER JOIN `sys_base_menus` AS menu
  ON menu.`id` = CAST(authority_menu.`sys_base_menu_id` AS UNSIGNED)
WHERE menu.`name` IN ('OrderFood', 'OrderFoodWorkspace');

DELETE FROM `sys_base_menus`
WHERE `name` IN ('OrderFood', 'OrderFoodWorkspace');

-- 业务菜单占用 1—8；GVA 自带菜单统一后置。
UPDATE `sys_base_menus`
SET
  `updated_at` = NOW(),
  `sort` = CASE `name`
    WHEN 'superAdmin' THEN 90
    WHEN 'person' THEN 91
    WHEN 'systemTools' THEN 92
    WHEN 'plugin' THEN 93
    WHEN 'example' THEN 94
    WHEN 'state' THEN 95
    WHEN 'about' THEN 99
    WHEN 'https://www.gin-vue-admin.com' THEN 100
  END
WHERE `parent_id` = 0
  AND `name` IN (
    'superAdmin',
    'person',
    'systemTools',
    'plugin',
    'example',
    'state',
    'about',
    'https://www.gin-vue-admin.com'
  );

-- 取消对原 GVA 仪表盘路由名的依赖。
ALTER TABLE `sys_authorities`
  ALTER COLUMN `default_router` SET DEFAULT 'OrderFoodDashboard';

UPDATE `sys_authorities`
SET `updated_at` = NOW(), `default_router` = 'OrderFoodDashboard'
WHERE `authority_id` IN (888, 9901, 9902);

UPDATE `sys_authorities`
SET `updated_at` = NOW(), `default_router` = 'about'
WHERE `authority_id` IN (9528, 8881)
   OR (`default_router` = 'dashboard' AND `authority_id` NOT IN (888, 9901, 9902));

COMMIT;
