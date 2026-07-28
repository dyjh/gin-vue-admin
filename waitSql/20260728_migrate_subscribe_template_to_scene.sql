-- 订阅消息管理由自由模板 CRUD 收敛为固定场景配置。
-- 适用于已经存在 OrderFoodSubscribeTemplates 菜单的 MySQL 数据库。
-- 脚本会：
-- 1. 原位重命名菜单，合并新 seed 可能已经创建的重复菜单；
-- 2. 将旧模板按钮授权映射为固定场景 read/update/status 权限；
-- 3. 删除旧模板 API/Casbin 路由，并按原角色边界映射到四条新路由。
-- 不修改 of_sub_templates 业务数据。

START TRANSACTION;

SET @legacy_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodSubscribeTemplates'
    ORDER BY id DESC
    LIMIT 1
);

SET @seeded_scene_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodSubscribeScenes'
    ORDER BY id DESC
    LIMIT 1
);

SET @target_menu_id := COALESCE(@legacy_menu_id, @seeded_scene_menu_id);
SET @duplicate_menu_id := IF(
    @legacy_menu_id IS NOT NULL
    AND @seeded_scene_menu_id IS NOT NULL
    AND @legacy_menu_id <> @seeded_scene_menu_id,
    @seeded_scene_menu_id,
    NULL
);

-- 先合并重复新菜单的可见角色，保留现有自定义角色边界。
INSERT INTO sys_authority_menus (
    sys_base_menu_id,
    sys_authority_authority_id
)
SELECT
    CAST(@target_menu_id AS CHAR),
    source_authority_menu.sys_authority_authority_id
FROM sys_authority_menus AS source_authority_menu
WHERE @duplicate_menu_id IS NOT NULL
  AND CAST(source_authority_menu.sys_base_menu_id AS UNSIGNED) = @duplicate_menu_id
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_menus AS target_authority_menu
      WHERE CAST(target_authority_menu.sys_base_menu_id AS UNSIGNED) = @target_menu_id
        AND target_authority_menu.sys_authority_authority_id =
            source_authority_menu.sys_authority_authority_id
  );

-- 确保目标菜单只保留三个新权限按钮。
INSERT INTO sys_base_menu_btns (
    created_at,
    updated_at,
    deleted_at,
    name,
    `desc`,
    sys_base_menu_id
)
SELECT NOW(), NOW(), NULL, permission.name, permission.description, @target_menu_id
FROM (
    SELECT 'orderfood:subscribe-scene:read' AS name, '查看固定订阅场景' AS description
    UNION ALL
    SELECT 'orderfood:subscribe-scene:update', '配置订阅场景模板绑定'
    UNION ALL
    SELECT 'orderfood:subscribe-scene:status', '启用或停用订阅场景'
) AS permission
WHERE @target_menu_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_base_menu_btns AS existing
      WHERE existing.sys_base_menu_id = @target_menu_id
        AND existing.name = permission.name
  );

UPDATE sys_base_menu_btns
SET updated_at = NOW(),
    deleted_at = NULL,
    `desc` = CASE name
        WHEN 'orderfood:subscribe-scene:read' THEN '查看固定订阅场景'
        WHEN 'orderfood:subscribe-scene:update' THEN '配置订阅场景模板绑定'
        WHEN 'orderfood:subscribe-scene:status' THEN '启用或停用订阅场景'
    END
WHERE sys_base_menu_id = @target_menu_id
  AND name IN (
      'orderfood:subscribe-scene:read',
      'orderfood:subscribe-scene:update',
      'orderfood:subscribe-scene:status'
  );

SET @scene_read_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @target_menu_id
      AND name = 'orderfood:subscribe-scene:read'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @scene_update_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @target_menu_id
      AND name = 'orderfood:subscribe-scene:update'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @scene_status_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @target_menu_id
      AND name = 'orderfood:subscribe-scene:status'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

-- 旧 read -> 新 read。
INSERT INTO sys_authority_btns (
    authority_id,
    sys_menu_id,
    sys_base_menu_btn_id
)
SELECT DISTINCT
    source_authority_btn.authority_id,
    @target_menu_id,
    @scene_read_btn_id
FROM sys_authority_btns AS source_authority_btn
INNER JOIN sys_base_menu_btns AS source_btn
    ON source_btn.id = source_authority_btn.sys_base_menu_btn_id
WHERE source_authority_btn.sys_menu_id IN (
        @target_menu_id,
        COALESCE(@duplicate_menu_id, @target_menu_id)
    )
  AND source_btn.name IN (
      'orderfood:subscribe-template:read',
      'orderfood:subscribe-scene:read'
  )
  AND @scene_read_btn_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_btns AS target_authority_btn
      WHERE target_authority_btn.authority_id = source_authority_btn.authority_id
        AND target_authority_btn.sys_menu_id = @target_menu_id
        AND target_authority_btn.sys_base_menu_btn_id = @scene_read_btn_id
  );

-- 旧 create/update -> 新 update。
INSERT INTO sys_authority_btns (
    authority_id,
    sys_menu_id,
    sys_base_menu_btn_id
)
SELECT DISTINCT
    source_authority_btn.authority_id,
    @target_menu_id,
    @scene_update_btn_id
FROM sys_authority_btns AS source_authority_btn
INNER JOIN sys_base_menu_btns AS source_btn
    ON source_btn.id = source_authority_btn.sys_base_menu_btn_id
WHERE source_authority_btn.sys_menu_id IN (
        @target_menu_id,
        COALESCE(@duplicate_menu_id, @target_menu_id)
    )
  AND source_btn.name IN (
      'orderfood:subscribe-template:create',
      'orderfood:subscribe-template:update',
      'orderfood:subscribe-scene:update'
  )
  AND @scene_update_btn_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_btns AS target_authority_btn
      WHERE target_authority_btn.authority_id = source_authority_btn.authority_id
        AND target_authority_btn.sys_menu_id = @target_menu_id
        AND target_authority_btn.sys_base_menu_btn_id = @scene_update_btn_id
  );

-- 旧 status -> 新 status。
INSERT INTO sys_authority_btns (
    authority_id,
    sys_menu_id,
    sys_base_menu_btn_id
)
SELECT DISTINCT
    source_authority_btn.authority_id,
    @target_menu_id,
    @scene_status_btn_id
FROM sys_authority_btns AS source_authority_btn
INNER JOIN sys_base_menu_btns AS source_btn
    ON source_btn.id = source_authority_btn.sys_base_menu_btn_id
WHERE source_authority_btn.sys_menu_id IN (
        @target_menu_id,
        COALESCE(@duplicate_menu_id, @target_menu_id)
    )
  AND source_btn.name IN (
      'orderfood:subscribe-template:status',
      'orderfood:subscribe-scene:status'
  )
  AND @scene_status_btn_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_btns AS target_authority_btn
      WHERE target_authority_btn.authority_id = source_authority_btn.authority_id
        AND target_authority_btn.sys_menu_id = @target_menu_id
        AND target_authority_btn.sys_base_menu_btn_id = @scene_status_btn_id
  );

-- 清理目标菜单上的五个旧按钮及其授权。
DELETE authority_btn
FROM sys_authority_btns AS authority_btn
INNER JOIN sys_base_menu_btns AS menu_btn
    ON menu_btn.id = authority_btn.sys_base_menu_btn_id
WHERE menu_btn.sys_base_menu_id = @target_menu_id
  AND menu_btn.name LIKE 'orderfood:subscribe-template:%';

DELETE FROM sys_base_menu_btns
WHERE sys_base_menu_id = @target_menu_id
  AND name LIKE 'orderfood:subscribe-template:%';

-- 清理并删除新 seed 可能生成的重复菜单；其授权已映射到目标菜单。
DELETE FROM sys_authority_btns
WHERE @duplicate_menu_id IS NOT NULL
  AND sys_menu_id = @duplicate_menu_id;

DELETE FROM sys_base_menu_btns
WHERE @duplicate_menu_id IS NOT NULL
  AND sys_base_menu_id = @duplicate_menu_id;

DELETE FROM sys_authority_menus
WHERE @duplicate_menu_id IS NOT NULL
  AND CAST(sys_base_menu_id AS UNSIGNED) = @duplicate_menu_id;

DELETE FROM sys_base_menu_parameters
WHERE @duplicate_menu_id IS NOT NULL
  AND sys_base_menu_id = @duplicate_menu_id;

DELETE FROM sys_base_menus
WHERE @duplicate_menu_id IS NOT NULL
  AND id = @duplicate_menu_id;

-- 原位更新菜单路由和标题。
UPDATE sys_base_menus
SET updated_at = NOW(),
    deleted_at = NULL,
    path = 'subscribe-scenes',
    name = 'OrderFoodSubscribeScenes',
    component = 'view/orderFood/subscribeTemplate/index.vue',
    title = '订阅场景配置',
    hidden = 0,
    keep_alive = 1
WHERE id = @target_menu_id;

-- 记录旧路由拥有者，按原 read/update/status 边界映射新 Casbin 路由。
CREATE TEMPORARY TABLE tmp_subscribe_scene_read_roles (
    authority_id varchar(64) PRIMARY KEY
);
CREATE TEMPORARY TABLE tmp_subscribe_scene_update_roles (
    authority_id varchar(64) PRIMARY KEY
);
CREATE TEMPORARY TABLE tmp_subscribe_scene_status_roles (
    authority_id varchar(64) PRIMARY KEY
);

INSERT IGNORE INTO tmp_subscribe_scene_read_roles (authority_id)
SELECT DISTINCT v0
FROM casbin_rule
WHERE ptype = 'p'
  AND v2 = 'GET'
  AND v1 IN (
      '/orderfood/subscribe-templates',
      '/orderfood/subscribe-templates/:templateId',
      '/orderfood/subscribe-scenes',
      '/orderfood/subscribe-scenes/:scene'
  );

INSERT IGNORE INTO tmp_subscribe_scene_update_roles (authority_id)
SELECT DISTINCT v0
FROM casbin_rule
WHERE ptype = 'p'
  AND (
      (v1 = '/orderfood/subscribe-templates' AND v2 = 'POST')
      OR (
          v1 IN (
              '/orderfood/subscribe-templates/:templateId',
              '/orderfood/subscribe-scenes/:scene'
          )
          AND v2 = 'PUT'
      )
  );

INSERT IGNORE INTO tmp_subscribe_scene_status_roles (authority_id)
SELECT DISTINCT v0
FROM casbin_rule
WHERE ptype = 'p'
  AND v2 = 'PUT'
  AND v1 IN (
      '/orderfood/subscribe-templates/:templateId/status',
      '/orderfood/subscribe-scenes/:scene/status'
  );

DELETE FROM casbin_rule
WHERE ptype = 'p'
  AND v1 LIKE '/orderfood/subscribe-templates%';

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', role.authority_id, route.path, 'GET', '', '', ''
FROM tmp_subscribe_scene_read_roles AS role
CROSS JOIN (
    SELECT '/orderfood/subscribe-scenes' AS path
    UNION ALL
    SELECT '/orderfood/subscribe-scenes/:scene'
) AS route
WHERE NOT EXISTS (
    SELECT 1
    FROM casbin_rule AS existing
    WHERE existing.ptype = 'p'
      AND existing.v0 = role.authority_id
      AND existing.v1 = route.path
      AND existing.v2 = 'GET'
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT
    'p',
    role.authority_id,
    '/orderfood/subscribe-scenes/:scene',
    'PUT',
    '',
    '',
    ''
FROM tmp_subscribe_scene_update_roles AS role
WHERE NOT EXISTS (
    SELECT 1
    FROM casbin_rule AS existing
    WHERE existing.ptype = 'p'
      AND existing.v0 = role.authority_id
      AND existing.v1 = '/orderfood/subscribe-scenes/:scene'
      AND existing.v2 = 'PUT'
);

INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
SELECT
    'p',
    role.authority_id,
    '/orderfood/subscribe-scenes/:scene/status',
    'PUT',
    '',
    '',
    ''
FROM tmp_subscribe_scene_status_roles AS role
WHERE NOT EXISTS (
    SELECT 1
    FROM casbin_rule AS existing
    WHERE existing.ptype = 'p'
      AND existing.v0 = role.authority_id
      AND existing.v1 = '/orderfood/subscribe-scenes/:scene/status'
      AND existing.v2 = 'PUT'
);

-- 替换 GVA API 元数据。
DELETE FROM sys_apis
WHERE path LIKE '/orderfood/subscribe-templates%';

INSERT INTO sys_apis (
    created_at,
    updated_at,
    deleted_at,
    path,
    description,
    api_group,
    method
)
SELECT NOW(), NOW(), NULL, api.path, api.description, '来干饭-消息中心', api.method
FROM (
    SELECT
        '/orderfood/subscribe-scenes' AS path,
        '查询固定订阅消息场景' AS description,
        'GET' AS method
    UNION ALL
    SELECT
        '/orderfood/subscribe-scenes/:scene',
        '获取固定订阅消息场景详情',
        'GET'
    UNION ALL
    SELECT
        '/orderfood/subscribe-scenes/:scene',
        '配置固定订阅场景模板绑定',
        'PUT'
    UNION ALL
    SELECT
        '/orderfood/subscribe-scenes/:scene/status',
        '启用或停用固定订阅场景',
        'PUT'
) AS api
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_apis AS existing
    WHERE existing.path = api.path
      AND existing.method = api.method
);

UPDATE sys_apis
SET updated_at = NOW(),
    deleted_at = NULL,
    api_group = '来干饭-消息中心',
    description = CASE
        WHEN path = '/orderfood/subscribe-scenes' AND method = 'GET'
            THEN '查询固定订阅消息场景'
        WHEN path = '/orderfood/subscribe-scenes/:scene' AND method = 'GET'
            THEN '获取固定订阅消息场景详情'
        WHEN path = '/orderfood/subscribe-scenes/:scene' AND method = 'PUT'
            THEN '配置固定订阅场景模板绑定'
        WHEN path = '/orderfood/subscribe-scenes/:scene/status' AND method = 'PUT'
            THEN '启用或停用固定订阅场景'
    END
WHERE path LIKE '/orderfood/subscribe-scenes%';

DROP TEMPORARY TABLE tmp_subscribe_scene_read_roles;
DROP TEMPORARY TABLE tmp_subscribe_scene_update_roles;
DROP TEMPORARY TABLE tmp_subscribe_scene_status_roles;

COMMIT;
