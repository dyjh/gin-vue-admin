-- 将后台菜单展示名称由“全部用户菜品 / 全部用户菜谱”
-- 收敛为“用户菜品 / 用户菜谱”。
-- 仅修改菜单标题，不改变路由、权限或业务数据。

START TRANSACTION;

UPDATE sys_base_menus
SET title = '用户菜品',
    updated_at = NOW()
WHERE name = 'OrderFoodUserDishes'
  AND deleted_at IS NULL;

UPDATE sys_base_menus
SET title = '用户菜谱',
    updated_at = NOW()
WHERE name = 'OrderFoodUserRecipes'
  AND deleted_at IS NULL;

COMMIT;