-- 为现有数据库补充常用菜品分类、菜品标签和食材单位。
-- 只插入 public_id 与名称均不存在的数据，不覆盖管理员已有配置，
-- 不恢复已软删除的数据，可重复执行。

START TRANSACTION;

INSERT INTO of_categories (
    created_at, updated_at, deleted_at,
    public_id, name, sort_order, enabled, version
)
SELECT
    NOW(), NOW(), NULL,
    seed.public_id, seed.name, seed.sort_order, 1, 1
FROM (
    SELECT 'category-stir-fry' AS public_id, '炒菜' AS name, 10 AS sort_order
    UNION ALL SELECT 'category-steamed', '蒸菜', 20
    UNION ALL SELECT 'category-cold-dish', '凉菜', 30
    UNION ALL SELECT 'category-stewed', '炖菜', 40
    UNION ALL SELECT 'category-braised', '烧菜', 50
    UNION ALL SELECT 'category-fried', '煎炸', 60
    UNION ALL SELECT 'category-roasted', '烤菜', 70
    UNION ALL SELECT 'category-soup', '汤羹', 80
    UNION ALL SELECT 'category-staple', '主食', 90
    UNION ALL SELECT 'category-dessert', '甜品', 100
) AS seed
WHERE NOT EXISTS (
    SELECT 1
    FROM of_categories AS existing
    WHERE existing.public_id = seed.public_id
       OR existing.name = seed.name
);

INSERT INTO of_tags (
    created_at, updated_at, deleted_at,
    public_id, name, sort_order, enabled, version
)
SELECT
    NOW(), NOW(), NULL,
    seed.public_id, seed.name, seed.sort_order, 1, 1
FROM (
    SELECT 'tag-cuisine-sichuan' AS public_id, '川菜' AS name, 10 AS sort_order
    UNION ALL SELECT 'tag-cuisine-hunan', '湘菜', 20
    UNION ALL SELECT 'tag-cuisine-cantonese', '粤菜', 30
    UNION ALL SELECT 'tag-cuisine-shandong', '鲁菜', 40
    UNION ALL SELECT 'tag-cuisine-jiangsu', '苏菜', 50
    UNION ALL SELECT 'tag-cuisine-zhejiang', '浙菜', 60
    UNION ALL SELECT 'tag-cuisine-fujian', '闽菜', 70
    UNION ALL SELECT 'tag-cuisine-anhui', '徽菜', 80
    UNION ALL SELECT 'tag-cuisine-northeast', '东北菜', 90
    UNION ALL SELECT 'tag-cuisine-northwest', '西北菜', 100
    UNION ALL SELECT 'tag-cuisine-yunnan-guizhou', '云贵菜', 110
    UNION ALL SELECT 'tag-cuisine-hakka', '客家菜', 120
    UNION ALL SELECT 'tag-cuisine-chaoshan', '潮汕菜', 130
    UNION ALL SELECT 'tag-cuisine-home-style', '家常菜', 140
    UNION ALL SELECT 'tag-flavor-light', '清淡', 150
    UNION ALL SELECT 'tag-flavor-savory', '咸鲜', 160
    UNION ALL SELECT 'tag-flavor-sweet-sour', '酸甜', 170
    UNION ALL SELECT 'tag-flavor-numbing-spicy', '麻辣', 180
    UNION ALL SELECT 'tag-spice-none', '不辣', 190
    UNION ALL SELECT 'tag-spice-mild', '微辣', 200
    UNION ALL SELECT 'tag-spice-medium', '中辣', 210
    UNION ALL SELECT 'tag-feature-quick', '快手菜', 220
    UNION ALL SELECT 'tag-feature-rice-friendly', '下饭菜', 230
    UNION ALL SELECT 'tag-feature-vegetarian', '素菜', 240
) AS seed
WHERE NOT EXISTS (
    SELECT 1
    FROM of_tags AS existing
    WHERE existing.public_id = seed.public_id
       OR existing.name = seed.name
);

INSERT INTO of_units (
    created_at, updated_at, deleted_at,
    public_id, name, sort_order, enabled, version
)
SELECT
    NOW(), NOW(), NULL,
    seed.public_id, seed.name, seed.sort_order, 1, 1
FROM (
    SELECT 'unit-gram' AS public_id, '克' AS name, 10 AS sort_order
    UNION ALL SELECT 'unit-kilogram', '千克', 20
    UNION ALL SELECT 'unit-milliliter', '毫升', 30
    UNION ALL SELECT 'unit-liter', '升', 40
    UNION ALL SELECT 'unit-piece', '个', 50
    UNION ALL SELECT 'unit-animal', '只', 60
    UNION ALL SELECT 'unit-egg', '枚', 70
    UNION ALL SELECT 'unit-slice', '片', 80
    UNION ALL SELECT 'unit-chunk', '块', 90
    UNION ALL SELECT 'unit-strip', '条', 100
    UNION ALL SELECT 'unit-root', '根', 110
    UNION ALL SELECT 'unit-grain', '颗', 120
    UNION ALL SELECT 'unit-clove', '瓣', 130
    UNION ALL SELECT 'unit-plant', '棵', 140
    UNION ALL SELECT 'unit-handful', '把', 150
    UNION ALL SELECT 'unit-serving', '份', 160
    UNION ALL SELECT 'unit-bowl', '碗', 170
    UNION ALL SELECT 'unit-cup', '杯', 180
    UNION ALL SELECT 'unit-tablespoon', '汤匙', 190
    UNION ALL SELECT 'unit-teaspoon', '茶匙', 200
    UNION ALL SELECT 'unit-pack', '包', 210
    UNION ALL SELECT 'unit-box', '盒', 220
    UNION ALL SELECT 'unit-pinch', '少许', 230
    UNION ALL SELECT 'unit-as-needed', '适量', 240
) AS seed
WHERE NOT EXISTS (
    SELECT 1
    FROM of_units AS existing
    WHERE existing.public_id = seed.public_id
       OR existing.name = seed.name
);

COMMIT;
