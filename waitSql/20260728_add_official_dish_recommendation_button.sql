-- 官方菜品列表需要在当前菜单读取“创建推荐草稿”按钮权限。
-- 新数据库由 server/source/orderfood_admin_seed.go 自动注入。
-- 本脚本只为现有数据库补齐菜单按钮，并复制“推荐精选”菜单现有授权；
-- 不新增 API 权限、不修改业务数据，也不覆盖管理员的自定义角色边界。

START TRANSACTION;

SET @recommendations_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodRecommendations'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @official_dishes_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodOfficialDishes'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @recommendation_create_source_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @recommendations_menu_id
      AND name = 'orderfood:recommendation:create'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

INSERT INTO sys_base_menu_btns (
    created_at,
    updated_at,
    name,
    `desc`,
    sys_base_menu_id
)
SELECT
    NOW(),
    NOW(),
    'orderfood:recommendation:create',
    '将官方菜品加入推荐草稿',
    @official_dishes_menu_id
WHERE @official_dishes_menu_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_base_menu_btns
      WHERE sys_base_menu_id = @official_dishes_menu_id
        AND name = 'orderfood:recommendation:create'
        AND deleted_at IS NULL
  );

UPDATE sys_base_menu_btns
SET `desc` = '将官方菜品加入推荐草稿',
    updated_at = NOW()
WHERE sys_base_menu_id = @official_dishes_menu_id
  AND name = 'orderfood:recommendation:create'
  AND deleted_at IS NULL;

SET @official_dish_recommendation_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @official_dishes_menu_id
      AND name = 'orderfood:recommendation:create'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

INSERT INTO sys_authority_btns (
    authority_id,
    sys_menu_id,
    sys_base_menu_btn_id
)
SELECT DISTINCT
    source_authority_btn.authority_id,
    @official_dishes_menu_id,
    @official_dish_recommendation_btn_id
FROM sys_authority_btns AS source_authority_btn
INNER JOIN sys_authorities AS authority
    ON authority.authority_id = source_authority_btn.authority_id
   AND authority.deleted_at IS NULL
WHERE source_authority_btn.sys_menu_id = @recommendations_menu_id
  AND source_authority_btn.sys_base_menu_btn_id = @recommendation_create_source_btn_id
  AND @recommendations_menu_id IS NOT NULL
  AND @official_dishes_menu_id IS NOT NULL
  AND @recommendation_create_source_btn_id IS NOT NULL
  AND @official_dish_recommendation_btn_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_btns AS target_authority_btn
      WHERE target_authority_btn.authority_id = source_authority_btn.authority_id
        AND target_authority_btn.sys_menu_id = @official_dishes_menu_id
        AND target_authority_btn.sys_base_menu_btn_id = @official_dish_recommendation_btn_id
  );

COMMIT;

SELECT
    menu.name AS menu_name,
    button.name AS button_name,
    authority.authority_id,
    authority.authority_name
FROM sys_authority_btns AS authority_button
INNER JOIN sys_base_menus AS menu
    ON menu.id = authority_button.sys_menu_id
INNER JOIN sys_base_menu_btns AS button
    ON button.id = authority_button.sys_base_menu_btn_id
INNER JOIN sys_authorities AS authority
    ON authority.authority_id = authority_button.authority_id
WHERE menu.name = 'OrderFoodOfficialDishes'
  AND menu.deleted_at IS NULL
  AND button.name = 'orderfood:recommendation:create'
  AND button.deleted_at IS NULL
  AND authority.deleted_at IS NULL
ORDER BY authority.authority_id;