-- 积分调整改为弹框后，清理原独立菜单。
-- 仅删除 OrderFoodPointAdjustments 菜单及其关联按钮授权；
-- 不删除 orderfood:points:adjust 权限，也不会修改积分业务数据。

START TRANSACTION;

SET @point_adjustment_menu_id := (
    SELECT id
    FROM sys_base_menus
    WHERE name = 'OrderFoodPointAdjustments'
    ORDER BY id DESC
    LIMIT 1
);

DELETE FROM sys_authority_btns
WHERE sys_menu_id = @point_adjustment_menu_id;

DELETE authority_btn
FROM sys_authority_btns AS authority_btn
INNER JOIN sys_base_menu_btns AS menu_btn
    ON menu_btn.id = authority_btn.sys_base_menu_btn_id
WHERE menu_btn.sys_base_menu_id = @point_adjustment_menu_id;

DELETE FROM sys_base_menu_btns
WHERE sys_base_menu_id = @point_adjustment_menu_id;

DELETE FROM sys_authority_menus
WHERE sys_base_menu_id = CAST(@point_adjustment_menu_id AS CHAR);

DELETE FROM sys_base_menu_parameters
WHERE sys_base_menu_id = @point_adjustment_menu_id;

DELETE FROM sys_base_menus
WHERE id = @point_adjustment_menu_id;

UPDATE sys_base_menus
SET sort = 3
WHERE name = 'OrderFoodPointRules'
  AND deleted_at IS NULL;

UPDATE sys_base_menu_btns AS menu_btn
INNER JOIN sys_base_menus AS menu
    ON menu.id = menu_btn.sys_base_menu_id
SET menu_btn.`desc` = '调整用户积分'
WHERE menu_btn.name = 'orderfood:points:adjust'
  AND menu.name IN ('OrderFoodUsers', 'OrderFoodPointEntries');

COMMIT;
