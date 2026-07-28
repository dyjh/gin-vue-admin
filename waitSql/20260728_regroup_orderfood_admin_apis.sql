-- 来干饭管理端 API 按一级功能域重新分组。
-- 适用范围：已经存在 sys_apis 数据的数据库。
-- 新数据库由 server/source/orderfood_admin_seed.go 自动写入正确分组。
-- 本脚本可重复执行。

START TRANSACTION;

UPDATE `sys_apis`
SET
  `api_group` = CASE SUBSTRING_INDEX(
    SUBSTRING(`path`, CHAR_LENGTH('/orderfood/') + 1),
    '/',
    1
  )
    WHEN 'dashboard' THEN '来干饭-数据概览'

    WHEN 'users' THEN '来干饭-用户与积分'
    WHEN 'point-entries' THEN '来干饭-用户与积分'
    WHEN 'point-adjustments' THEN '来干饭-用户与积分'
    WHEN 'point-rules' THEN '来干饭-用户与积分'

    WHEN 'user-dishes' THEN '来干饭-菜品运营'
    WHEN 'user-recipes' THEN '来干饭-菜品运营'
    WHEN 'suggestion-catalog' THEN '来干饭-菜品运营'
    WHEN 'discoverable-dishes' THEN '来干饭-菜品运营'
    WHEN 'recommendations' THEN '来干饭-菜品运营'
    WHEN 'official-dishes' THEN '来干饭-菜品运营'
    WHEN 'official-dish-covers' THEN '来干饭-菜品运营'
    WHEN 'categories' THEN '来干饭-菜品运营'
    WHEN 'tags' THEN '来干饭-菜品运营'
    WHEN 'units' THEN '来干饭-菜品运营'

    WHEN 'media' THEN '来干饭-内容安全'
    WHEN 'moderation-records' THEN '来干饭-内容安全'
    WHEN 'moderation-config' THEN '来干饭-内容安全'
    WHEN 'governance-records' THEN '来干饭-内容安全'
    WHEN 'governance-jobs' THEN '来干饭-内容安全'
    WHEN 'governance-actions' THEN '来干饭-内容安全'
    WHEN 'audit-logs' THEN '来干饭-内容安全'

    WHEN 'meals' THEN '来干饭-饭局管理'
    WHEN 'shopping-lists' THEN '来干饭-饭局管理'

    WHEN 'platform-capability-policy' THEN '来干饭-AI 能力'
    WHEN 'ai-providers' THEN '来干饭-AI 能力'
    WHEN 'ai-model-provider-options' THEN '来干饭-AI 能力'
    WHEN 'ai-models' THEN '来干饭-AI 能力'
    WHEN 'ai-capabilities' THEN '来干饭-AI 能力'
    WHEN 'ai-usages' THEN '来干饭-AI 能力'

    WHEN 'notifications' THEN '来干饭-消息中心'
    WHEN 'subscribe-scenes' THEN '来干饭-消息中心'
    WHEN 'subscribe-templates' THEN '来干饭-消息中心'
    WHEN 'subscribe-logs' THEN '来干饭-消息中心'

    WHEN 'wechat-config' THEN '来干饭-微信配置'
    ELSE `api_group`
  END,
  `updated_at` = NOW()
WHERE `deleted_at` IS NULL
  AND `path` LIKE '/orderfood/%';

COMMIT;

-- 核对结果：应只有上面的 8 个来干饭功能分组。
SELECT `api_group`, COUNT(*) AS `api_count`
FROM `sys_apis`
WHERE `deleted_at` IS NULL
  AND `path` LIKE '/orderfood/%'
GROUP BY `api_group`
ORDER BY `api_group`;
