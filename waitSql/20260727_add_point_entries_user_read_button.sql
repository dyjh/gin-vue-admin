-- 积分调整弹框位于积分流水页时，需要在当前菜单读取用户列表。
-- 本脚本只复制现有“小程序用户”菜单的 user:read 按钮授权；
-- 不新增 API 权限、不修改业务数据，也不覆盖管理员的自定义角色边界。

START TRANSACTION;

SET @users_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodUsers'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @point_entries_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodPointEntries'
      AND deleted_at IS NULL
    ORDER BY id DESC
    LIMIT 1
);

SET @users_user_read_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @users_menu_id
      AND name = 'orderfood:user:read'
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
    'orderfood:user:read',
    '查看用户',
    @point_entries_menu_id
WHERE @point_entries_menu_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_base_menu_btns
      WHERE sys_base_menu_id = @point_entries_menu_id
        AND name = 'orderfood:user:read'
        AND deleted_at IS NULL
  );

UPDATE sys_base_menu_btns
SET `desc` = '查看用户',
    updated_at = NOW()
WHERE sys_base_menu_id = @point_entries_menu_id
  AND name = 'orderfood:user:read'
  AND deleted_at IS NULL;

SET @point_entries_user_read_btn_id := (
    SELECT id
    FROM sys_base_menu_btns
    WHERE sys_base_menu_id = @point_entries_menu_id
      AND name = 'orderfood:user:read'
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
    @point_entries_menu_id,
    @point_entries_user_read_btn_id
FROM sys_authority_btns AS source_authority_btn
INNER JOIN sys_authorities AS authority
    ON authority.authority_id = source_authority_btn.authority_id
   AND authority.deleted_at IS NULL
WHERE source_authority_btn.sys_menu_id = @users_menu_id
  AND source_authority_btn.sys_base_menu_btn_id = @users_user_read_btn_id
  AND @users_menu_id IS NOT NULL
  AND @point_entries_menu_id IS NOT NULL
  AND @users_user_read_btn_id IS NOT NULL
  AND @point_entries_user_read_btn_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM sys_authority_btns AS target_authority_btn
      WHERE target_authority_btn.authority_id = source_authority_btn.authority_id
        AND target_authority_btn.sys_menu_id = @point_entries_menu_id
        AND target_authority_btn.sys_base_menu_btn_id = @point_entries_user_read_btn_id
  );

COMMIT;
