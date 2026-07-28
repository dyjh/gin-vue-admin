-- 为现有数据库补充 20 道可用状态的官方菜品，以及菜品引用的本地封面媒体记录。
-- 前置条件：
--   1. 已执行 waitSql/20260727_seed_common_catalog_data.sql；
--   2. 将 server/uploads/file/cover-official-*.jpg 原样同步到测试服相同目录；
--   3. sys_users 中至少存在一个 enable = 1 且未软删除的管理员。
-- 脚本按 public_id / 媒体 ID 幂等插入，不覆盖已有官方菜品和媒体。
-- 执行结束后请检查最后的 prerequisite_ready 与四项 actual 计数。

SET NAMES utf8mb4;
START TRANSACTION;

DROP TEMPORARY TABLE IF EXISTS tmp_official_media_seed;
CREATE TEMPORARY TABLE tmp_official_media_seed (
    id VARCHAR(64) PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,
    storage_path VARCHAR(1024) NOT NULL,
    width INT NOT NULL,
    height INT NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL
);
INSERT INTO tmp_official_media_seed (
    id, file_name, url, storage_path, width, height, size_bytes, checksum
) VALUES
    ('cover-official-tomato-egg', '番茄炒蛋.jpg', '/uploads/file/cover-official-tomato-egg.jpg', 'uploads/file/cover-official-tomato-egg.jpg', 1254, 1254, 262141, '67f7c96dcac04d9fa1c001c4b281c4e371138c46343ecc8d76b21a635fcced5a'),
    ('cover-official-pepper-pork', '青椒肉丝.jpg', '/uploads/file/cover-official-pepper-pork.jpg', 'uploads/file/cover-official-pepper-pork.jpg', 1254, 1254, 257320, '69984bb42201453698531a9bc9ffd401f3250bbede0251180ff5dc898e7a3e55'),
    ('cover-official-kung-pao-chicken', '宫保鸡丁.jpg', '/uploads/file/cover-official-kung-pao-chicken.jpg', 'uploads/file/cover-official-kung-pao-chicken.jpg', 1254, 1254, 256004, '1f5135252f6f19d6796096fa5df541f98a322c1e489f5a1df44ac562075fc924'),
    ('cover-official-garlic-broccoli', '蒜蓉西兰花.jpg', '/uploads/file/cover-official-garlic-broccoli.jpg', 'uploads/file/cover-official-garlic-broccoli.jpg', 1254, 1254, 226996, '333f17b2d7178b10470947d41db933d8b5dfb42629c33fba7b9abe0b081298f6'),
    ('cover-official-steamed-sea-bass', '清蒸鲈鱼.jpg', '/uploads/file/cover-official-steamed-sea-bass.jpg', 'uploads/file/cover-official-steamed-sea-bass.jpg', 1254, 1254, 261144, '6619bf2abee1c99fb491552b567e81fc979fe1fddb6308e06cf559ee7f1f3018'),
    ('cover-official-garlic-shrimp', '蒜蓉粉丝蒸虾.jpg', '/uploads/file/cover-official-garlic-shrimp.jpg', 'uploads/file/cover-official-garlic-shrimp.jpg', 1254, 1254, 249805, 'b715936a0d8957cc1821c3b0d24112eed14578ea7df94a4e6604a64b215d0d72'),
    ('cover-official-cucumber-salad', '凉拌黄瓜.jpg', '/uploads/file/cover-official-cucumber-salad.jpg', 'uploads/file/cover-official-cucumber-salad.jpg', 1254, 1254, 246929, '15fed1f3aa46cc8397ef6c37cba2e219b8070a145348979d624c33e83b6eeb21'),
    ('cover-official-white-cut-chicken', '白切鸡.jpg', '/uploads/file/cover-official-white-cut-chicken.jpg', 'uploads/file/cover-official-white-cut-chicken.jpg', 1254, 1254, 235683, '6555a6cb7e1bb558d46f1306629aa0f975e1b8f5f302a19ad09ead09845731db'),
    ('cover-official-beef-potato-stew', '土豆炖牛腩.jpg', '/uploads/file/cover-official-beef-potato-stew.jpg', 'uploads/file/cover-official-beef-potato-stew.jpg', 1254, 1254, 198950, 'add9fb4fc49cf852dbff638617803dfaac9a6a734e9780d90c9395399dad1392'),
    ('cover-official-northeast-stew', '东北乱炖.jpg', '/uploads/file/cover-official-northeast-stew.jpg', 'uploads/file/cover-official-northeast-stew.jpg', 1254, 1254, 275554, '05b3b5d04e8997a780c1a886788133e0751bbbe52d91c2cdb44ac4d3ad581aa1'),
    ('cover-official-mapo-tofu', '麻婆豆腐.jpg', '/uploads/file/cover-official-mapo-tofu.jpg', 'uploads/file/cover-official-mapo-tofu.jpg', 1254, 1254, 279693, '0e8bdf32b2a8e3e4a6e9c4bbe67ef28240206a374d4c836d197eaf3ab69fde96'),
    ('cover-official-braised-pork', '红烧肉.jpg', '/uploads/file/cover-official-braised-pork.jpg', 'uploads/file/cover-official-braised-pork.jpg', 1254, 1254, 258730, '9d03d087e2ecb16853fae296f589397c9619c3897bf4972bd099b125cd1439d2'),
    ('cover-official-cola-wings', '可乐鸡翅.jpg', '/uploads/file/cover-official-cola-wings.jpg', 'uploads/file/cover-official-cola-wings.jpg', 1254, 1254, 228817, 'd53526fb36ad81696478df9a5025442faa627c567509449c9795f84dac3b4e0c'),
    ('cover-official-sweet-sour-pork', '糖醋里脊.jpg', '/uploads/file/cover-official-sweet-sour-pork.jpg', 'uploads/file/cover-official-sweet-sour-pork.jpg', 1254, 1254, 252777, '114533100a56abca2d5259346d2d277d7fe554847c7166031a3c40f6d65ae5f6'),
    ('cover-official-pan-fried-hairtail', '香煎带鱼.jpg', '/uploads/file/cover-official-pan-fried-hairtail.jpg', 'uploads/file/cover-official-pan-fried-hairtail.jpg', 1254, 1254, 190017, '22efd7a3096c0760d639c362d81378b21a1307bc452386b45f4247e21403077a'),
    ('cover-official-honey-roast-ribs', '蜜汁烤排骨.jpg', '/uploads/file/cover-official-honey-roast-ribs.jpg', 'uploads/file/cover-official-honey-roast-ribs.jpg', 1254, 1254, 219507, '214b9dd918e27b38510e13526a7c71294a9039e3a0ae27c26b23d35f58d86fd9'),
    ('cover-official-seaweed-egg-soup', '紫菜蛋花汤.jpg', '/uploads/file/cover-official-seaweed-egg-soup.jpg', 'uploads/file/cover-official-seaweed-egg-soup.jpg', 1254, 1254, 244523, '693c693ec03e77af40ee5f309110d28a4972b4e5a59ebdcab11fe1168d22edb2'),
    ('cover-official-corn-rib-soup', '玉米排骨汤.jpg', '/uploads/file/cover-official-corn-rib-soup.jpg', 'uploads/file/cover-official-corn-rib-soup.jpg', 1254, 1254, 221918, '80dbd0f74273752daa4821193490daab303997697c424e4e9b0e7d3bb70d8b84'),
    ('cover-official-yangzhou-fried-rice', '扬州炒饭.jpg', '/uploads/file/cover-official-yangzhou-fried-rice.jpg', 'uploads/file/cover-official-yangzhou-fried-rice.jpg', 1254, 1254, 206942, 'cef2e19a864e1a6e59fc6dfee2c235f7b4b575b1d3becf4ab7c6d8179a039e98'),
    ('cover-official-brown-sugar-ciba', '红糖糍粑.jpg', '/uploads/file/cover-official-brown-sugar-ciba.jpg', 'uploads/file/cover-official-brown-sugar-ciba.jpg', 1254, 1254, 286655, 'b1c456c08a46788ce0b4f50b77f1b783f4adb41dced0d7ee4d575e9518b50871');

DROP TEMPORARY TABLE IF EXISTS tmp_official_dish_seed;
CREATE TEMPORARY TABLE tmp_official_dish_seed (
    public_id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(40) NOT NULL,
    cover_file_id VARCHAR(64) NOT NULL,
    category_public_id VARCHAR(64) NOT NULL,
    serving INT NOT NULL,
    description VARCHAR(180) NOT NULL
);
INSERT INTO tmp_official_dish_seed (
    public_id, name, cover_file_id, category_public_id, serving, description
) VALUES
    ('official-tomato-egg', '番茄炒蛋', 'cover-official-tomato-egg', 'category-stir-fry', 2, '酸甜开胃的经典家常菜，鸡蛋蓬松、番茄多汁，适合日常快手下饭。'),
    ('official-pepper-pork', '青椒肉丝', 'cover-official-pepper-pork', 'category-stir-fry', 2, '青椒清香、肉丝滑嫩，咸鲜微辣，是简单耐吃的家常下饭菜。'),
    ('official-kung-pao-chicken', '宫保鸡丁', 'cover-official-kung-pao-chicken', 'category-stir-fry', 3, '鸡丁嫩滑、花生酥香，酸甜咸辣平衡，具有代表性的川味下饭菜。'),
    ('official-garlic-broccoli', '蒜蓉西兰花', 'cover-official-garlic-broccoli', 'category-stir-fry', 2, '清爽脆嫩的家常素菜，蒜香自然，调味简洁，适合搭配荤菜。'),
    ('official-steamed-sea-bass', '清蒸鲈鱼', 'cover-official-steamed-sea-bass', 'category-steamed', 3, '突出鲈鱼本味的清蒸做法，鱼肉细嫩，葱姜与豉油清鲜不抢味。'),
    ('official-garlic-shrimp', '蒜蓉粉丝蒸虾', 'cover-official-garlic-shrimp', 'category-steamed', 3, '鲜虾与粉丝吸收蒜蓉豉汁，鲜香入味，适合家庭聚餐。'),
    ('official-cucumber-salad', '凉拌黄瓜', 'cover-official-cucumber-salad', 'category-cold-dish', 2, '爽脆开胃的家常凉菜，蒜香、醋香与轻微辣味清爽平衡。'),
    ('official-white-cut-chicken', '白切鸡', 'cover-official-white-cut-chicken', 'category-cold-dish', 4, '皮爽肉嫩的粤式经典，保留鸡肉本味，搭配姜葱蘸料食用。'),
    ('official-beef-potato-stew', '土豆炖牛腩', 'cover-official-beef-potato-stew', 'category-stewed', 4, '牛腩软烂、土豆绵软，汤汁浓郁，是适合家庭分享的暖胃炖菜。'),
    ('official-northeast-stew', '东北乱炖', 'cover-official-northeast-stew', 'category-stewed', 4, '排骨与多种时蔬一锅慢炖，食材丰富、汤味厚实，适合家庭共享。'),
    ('official-mapo-tofu', '麻婆豆腐', 'cover-official-mapo-tofu', 'category-braised', 3, '豆腐嫩滑，麻、辣、鲜、香层次清楚，是经典川味下饭菜。'),
    ('official-braised-pork', '红烧肉', 'cover-official-braised-pork', 'category-braised', 4, '五花肉酥而不散、肥而不腻，酱色红亮，咸甜适口。'),
    ('official-cola-wings', '可乐鸡翅', 'cover-official-cola-wings', 'category-braised', 3, '鸡翅软嫩入味，咸甜平衡，做法简单，适合家庭日常制作。'),
    ('official-sweet-sour-pork', '糖醋里脊', 'cover-official-sweet-sour-pork', 'category-fried', 3, '外酥里嫩，糖醋汁酸甜明亮，是老少皆宜的经典菜。'),
    ('official-pan-fried-hairtail', '香煎带鱼', 'cover-official-pan-fried-hairtail', 'category-fried', 3, '带鱼外层焦香、内部细嫩，咸鲜适口，适合家常配饭。'),
    ('official-honey-roast-ribs', '蜜汁烤排骨', 'cover-official-honey-roast-ribs', 'category-roasted', 4, '排骨外层焦香、肉质多汁，蜜汁咸甜适中，适合聚餐分享。'),
    ('official-seaweed-egg-soup', '紫菜蛋花汤', 'cover-official-seaweed-egg-soup', 'category-soup', 3, '清鲜快手的家常汤，蛋花轻盈，紫菜鲜香，几分钟即可完成。'),
    ('official-corn-rib-soup', '玉米排骨汤', 'cover-official-corn-rib-soup', 'category-soup', 4, '排骨汤清甜温润，搭配玉米和胡萝卜，适合家庭日常饮用。'),
    ('official-yangzhou-fried-rice', '扬州炒饭', 'cover-official-yangzhou-fried-rice', 'category-staple', 3, '米粒松散，鸡蛋、虾仁与时蔬分布均匀，咸鲜丰富的经典主食。'),
    ('official-brown-sugar-ciba', '红糖糍粑', 'cover-official-brown-sugar-ciba', 'category-dessert', 3, '外层微酥、内部软糯，红糖汁香甜，搭配熟黄豆粉风味浓郁。');

DROP TEMPORARY TABLE IF EXISTS tmp_official_ingredient_seed;
CREATE TEMPORARY TABLE tmp_official_ingredient_seed (
    dish_public_id VARCHAR(64) NOT NULL,
    name VARCHAR(120) NOT NULL,
    quantity VARCHAR(40) NULL,
    unit_public_id VARCHAR(64) NULL,
    note VARCHAR(300) NULL,
    sort_order INT NOT NULL,
    INDEX idx_tmp_official_ingredient_dish (dish_public_id)
);
INSERT INTO tmp_official_ingredient_seed (
    dish_public_id, name, quantity, unit_public_id, note, sort_order
) VALUES
    ('official-tomato-egg', '番茄', '3', 'unit-piece', NULL, 1),
    ('official-tomato-egg', '鸡蛋', '4', 'unit-egg', NULL, 2),
    ('official-tomato-egg', '小葱', '1', 'unit-root', NULL, 3),
    ('official-tomato-egg', '食用油', '2', 'unit-tablespoon', NULL, 4),
    ('official-tomato-egg', '盐', '1/2', 'unit-teaspoon', NULL, 5),
    ('official-tomato-egg', '白糖', '1', 'unit-teaspoon', NULL, 6),
    ('official-pepper-pork', '猪里脊', '250', 'unit-gram', NULL, 1),
    ('official-pepper-pork', '青椒', '3', 'unit-piece', NULL, 2),
    ('official-pepper-pork', '姜', '5', 'unit-gram', NULL, 3),
    ('official-pepper-pork', '生抽', '1', 'unit-tablespoon', NULL, 4),
    ('official-pepper-pork', '料酒', '1', 'unit-tablespoon', NULL, 5),
    ('official-pepper-pork', '淀粉', '1', 'unit-teaspoon', NULL, 6),
    ('official-pepper-pork', '食用油', '2', 'unit-tablespoon', NULL, 7),
    ('official-pepper-pork', '盐', '1/2', 'unit-teaspoon', NULL, 8),
    ('official-kung-pao-chicken', '鸡腿肉', '350', 'unit-gram', '去骨', 1),
    ('official-kung-pao-chicken', '熟花生米', '60', 'unit-gram', NULL, 2),
    ('official-kung-pao-chicken', '大葱', '2', 'unit-root', NULL, 3),
    ('official-kung-pao-chicken', '干辣椒', '8', 'unit-piece', NULL, 4),
    ('official-kung-pao-chicken', '花椒', '1', 'unit-teaspoon', NULL, 5),
    ('official-kung-pao-chicken', '生抽', '1', 'unit-tablespoon', NULL, 6),
    ('official-kung-pao-chicken', '香醋', '2', 'unit-tablespoon', NULL, 7),
    ('official-kung-pao-chicken', '白糖', '1', 'unit-tablespoon', NULL, 8),
    ('official-kung-pao-chicken', '淀粉', '1', 'unit-teaspoon', NULL, 9),
    ('official-garlic-broccoli', '西兰花', '1', 'unit-plant', NULL, 1),
    ('official-garlic-broccoli', '大蒜', '4', 'unit-clove', NULL, 2),
    ('official-garlic-broccoli', '食用油', '1', 'unit-tablespoon', NULL, 3),
    ('official-garlic-broccoli', '盐', '1/2', 'unit-teaspoon', NULL, 4),
    ('official-garlic-broccoli', '清水', '2', 'unit-tablespoon', NULL, 5),
    ('official-steamed-sea-bass', '鲈鱼', '1', 'unit-animal', '约 600 克', 1),
    ('official-steamed-sea-bass', '姜', '20', 'unit-gram', NULL, 2),
    ('official-steamed-sea-bass', '小葱', '2', 'unit-root', NULL, 3),
    ('official-steamed-sea-bass', '蒸鱼豉油', '2', 'unit-tablespoon', NULL, 4),
    ('official-steamed-sea-bass', '料酒', '1', 'unit-tablespoon', NULL, 5),
    ('official-steamed-sea-bass', '食用油', '1', 'unit-tablespoon', NULL, 6),
    ('official-steamed-sea-bass', '盐', '1', 'unit-pinch', NULL, 7),
    ('official-garlic-shrimp', '鲜虾', '12', 'unit-animal', NULL, 1),
    ('official-garlic-shrimp', '粉丝', '100', 'unit-gram', NULL, 2),
    ('official-garlic-shrimp', '大蒜', '8', 'unit-clove', NULL, 3),
    ('official-garlic-shrimp', '小葱', '1', 'unit-root', NULL, 4),
    ('official-garlic-shrimp', '生抽', '1.5', 'unit-tablespoon', NULL, 5),
    ('official-garlic-shrimp', '蚝油', '1', 'unit-teaspoon', NULL, 6),
    ('official-garlic-shrimp', '白糖', '1/2', 'unit-teaspoon', NULL, 7),
    ('official-garlic-shrimp', '食用油', '2', 'unit-tablespoon', NULL, 8),
    ('official-cucumber-salad', '黄瓜', '2', 'unit-root', NULL, 1),
    ('official-cucumber-salad', '大蒜', '4', 'unit-clove', NULL, 2),
    ('official-cucumber-salad', '小米椒', '2', 'unit-piece', NULL, 3),
    ('official-cucumber-salad', '香醋', '2', 'unit-tablespoon', NULL, 4),
    ('official-cucumber-salad', '生抽', '1', 'unit-tablespoon', NULL, 5),
    ('official-cucumber-salad', '香油', '1', 'unit-teaspoon', NULL, 6),
    ('official-cucumber-salad', '白糖', '1', 'unit-teaspoon', NULL, 7),
    ('official-cucumber-salad', '盐', '1/2', 'unit-teaspoon', NULL, 8),
    ('official-white-cut-chicken', '三黄鸡', '1', 'unit-animal', '约 1000 克', 1),
    ('official-white-cut-chicken', '姜', '30', 'unit-gram', NULL, 2),
    ('official-white-cut-chicken', '小葱', '4', 'unit-root', NULL, 3),
    ('official-white-cut-chicken', '盐', '1', 'unit-teaspoon', NULL, 4),
    ('official-white-cut-chicken', '香油', '1', 'unit-teaspoon', NULL, 5),
    ('official-white-cut-chicken', '食用油', '2', 'unit-tablespoon', NULL, 6),
    ('official-beef-potato-stew', '牛腩', '600', 'unit-gram', NULL, 1),
    ('official-beef-potato-stew', '土豆', '3', 'unit-piece', NULL, 2),
    ('official-beef-potato-stew', '胡萝卜', '1', 'unit-root', NULL, 3),
    ('official-beef-potato-stew', '姜', '20', 'unit-gram', NULL, 4),
    ('official-beef-potato-stew', '大葱', '2', 'unit-root', NULL, 5),
    ('official-beef-potato-stew', '八角', '2', 'unit-grain', NULL, 6),
    ('official-beef-potato-stew', '生抽', '2', 'unit-tablespoon', NULL, 7),
    ('official-beef-potato-stew', '老抽', '1', 'unit-teaspoon', NULL, 8),
    ('official-beef-potato-stew', '料酒', '2', 'unit-tablespoon', NULL, 9),
    ('official-beef-potato-stew', '清水', NULL, 'unit-as-needed', NULL, 10),
    ('official-northeast-stew', '猪肋排', '500', 'unit-gram', NULL, 1),
    ('official-northeast-stew', '豆角', '300', 'unit-gram', NULL, 2),
    ('official-northeast-stew', '土豆', '2', 'unit-piece', NULL, 3),
    ('official-northeast-stew', '玉米', '1', 'unit-root', NULL, 4),
    ('official-northeast-stew', '南瓜', '300', 'unit-gram', NULL, 5),
    ('official-northeast-stew', '番茄', '2', 'unit-piece', NULL, 6),
    ('official-northeast-stew', '姜', '15', 'unit-gram', NULL, 7),
    ('official-northeast-stew', '生抽', '2', 'unit-tablespoon', NULL, 8),
    ('official-northeast-stew', '清水', NULL, 'unit-as-needed', NULL, 9),
    ('official-mapo-tofu', '嫩豆腐', '500', 'unit-gram', NULL, 1),
    ('official-mapo-tofu', '牛肉末', '120', 'unit-gram', NULL, 2),
    ('official-mapo-tofu', '郫县豆瓣酱', '2', 'unit-tablespoon', NULL, 3),
    ('official-mapo-tofu', '花椒', '1', 'unit-teaspoon', NULL, 4),
    ('official-mapo-tofu', '大蒜', '3', 'unit-clove', NULL, 5),
    ('official-mapo-tofu', '小葱', '2', 'unit-root', NULL, 6),
    ('official-mapo-tofu', '生抽', '1', 'unit-tablespoon', NULL, 7),
    ('official-mapo-tofu', '淀粉', '1', 'unit-teaspoon', NULL, 8),
    ('official-mapo-tofu', '清水', '150', 'unit-milliliter', NULL, 9),
    ('official-braised-pork', '五花肉', '600', 'unit-gram', NULL, 1),
    ('official-braised-pork', '冰糖', '30', 'unit-gram', NULL, 2),
    ('official-braised-pork', '姜', '20', 'unit-gram', NULL, 3),
    ('official-braised-pork', '大葱', '2', 'unit-root', NULL, 4),
    ('official-braised-pork', '八角', '2', 'unit-grain', NULL, 5),
    ('official-braised-pork', '生抽', '2', 'unit-tablespoon', NULL, 6),
    ('official-braised-pork', '老抽', '1', 'unit-tablespoon', NULL, 7),
    ('official-braised-pork', '料酒', '2', 'unit-tablespoon', NULL, 8),
    ('official-braised-pork', '清水', NULL, 'unit-as-needed', NULL, 9),
    ('official-cola-wings', '鸡中翅', '10', 'unit-piece', NULL, 1),
    ('official-cola-wings', '可乐', '500', 'unit-milliliter', NULL, 2),
    ('official-cola-wings', '姜', '15', 'unit-gram', NULL, 3),
    ('official-cola-wings', '小葱', '2', 'unit-root', NULL, 4),
    ('official-cola-wings', '生抽', '2', 'unit-tablespoon', NULL, 5),
    ('official-cola-wings', '料酒', '1', 'unit-tablespoon', NULL, 6),
    ('official-cola-wings', '盐', '1/2', 'unit-teaspoon', NULL, 7),
    ('official-sweet-sour-pork', '猪里脊', '400', 'unit-gram', NULL, 1),
    ('official-sweet-sour-pork', '鸡蛋', '1', 'unit-egg', NULL, 2),
    ('official-sweet-sour-pork', '淀粉', '100', 'unit-gram', NULL, 3),
    ('official-sweet-sour-pork', '面粉', '50', 'unit-gram', NULL, 4),
    ('official-sweet-sour-pork', '番茄酱', '3', 'unit-tablespoon', NULL, 5),
    ('official-sweet-sour-pork', '白醋', '2', 'unit-tablespoon', NULL, 6),
    ('official-sweet-sour-pork', '白糖', '2', 'unit-tablespoon', NULL, 7),
    ('official-sweet-sour-pork', '食用油', NULL, 'unit-as-needed', '用于炸制', 8),
    ('official-pan-fried-hairtail', '带鱼', '600', 'unit-gram', NULL, 1),
    ('official-pan-fried-hairtail', '姜', '15', 'unit-gram', NULL, 2),
    ('official-pan-fried-hairtail', '小葱', '2', 'unit-root', NULL, 3),
    ('official-pan-fried-hairtail', '料酒', '2', 'unit-tablespoon', NULL, 4),
    ('official-pan-fried-hairtail', '盐', '1', 'unit-teaspoon', NULL, 5),
    ('official-pan-fried-hairtail', '淀粉', '60', 'unit-gram', NULL, 6),
    ('official-pan-fried-hairtail', '食用油', '3', 'unit-tablespoon', NULL, 7),
    ('official-honey-roast-ribs', '猪肋排', '700', 'unit-gram', NULL, 1),
    ('official-honey-roast-ribs', '蜂蜜', '2', 'unit-tablespoon', NULL, 2),
    ('official-honey-roast-ribs', '叉烧酱', '3', 'unit-tablespoon', NULL, 3),
    ('official-honey-roast-ribs', '生抽', '1', 'unit-tablespoon', NULL, 4),
    ('official-honey-roast-ribs', '蚝油', '1', 'unit-tablespoon', NULL, 5),
    ('official-honey-roast-ribs', '大蒜', '4', 'unit-clove', NULL, 6),
    ('official-honey-roast-ribs', '料酒', '1', 'unit-tablespoon', NULL, 7),
    ('official-seaweed-egg-soup', '干紫菜', '10', 'unit-gram', NULL, 1),
    ('official-seaweed-egg-soup', '鸡蛋', '2', 'unit-egg', NULL, 2),
    ('official-seaweed-egg-soup', '虾皮', '10', 'unit-gram', NULL, 3),
    ('official-seaweed-egg-soup', '小葱', '1', 'unit-root', NULL, 4),
    ('official-seaweed-egg-soup', '盐', '1/2', 'unit-teaspoon', NULL, 5),
    ('official-seaweed-egg-soup', '香油', '1', 'unit-teaspoon', NULL, 6),
    ('official-seaweed-egg-soup', '清水', '800', 'unit-milliliter', NULL, 7),
    ('official-corn-rib-soup', '猪肋排', '500', 'unit-gram', NULL, 1),
    ('official-corn-rib-soup', '甜玉米', '2', 'unit-root', NULL, 2),
    ('official-corn-rib-soup', '胡萝卜', '1', 'unit-root', NULL, 3),
    ('official-corn-rib-soup', '姜', '15', 'unit-gram', NULL, 4),
    ('official-corn-rib-soup', '小葱', '2', 'unit-root', NULL, 5),
    ('official-corn-rib-soup', '料酒', '1', 'unit-tablespoon', NULL, 6),
    ('official-corn-rib-soup', '盐', '1', 'unit-teaspoon', NULL, 7),
    ('official-corn-rib-soup', '清水', '1500', 'unit-milliliter', NULL, 8),
    ('official-yangzhou-fried-rice', '隔夜米饭', '3', 'unit-bowl', NULL, 1),
    ('official-yangzhou-fried-rice', '鸡蛋', '3', 'unit-egg', NULL, 2),
    ('official-yangzhou-fried-rice', '虾仁', '100', 'unit-gram', NULL, 3),
    ('official-yangzhou-fried-rice', '火腿', '100', 'unit-gram', NULL, 4),
    ('official-yangzhou-fried-rice', '青豆', '80', 'unit-gram', NULL, 5),
    ('official-yangzhou-fried-rice', '胡萝卜', '80', 'unit-gram', NULL, 6),
    ('official-yangzhou-fried-rice', '小葱', '2', 'unit-root', NULL, 7),
    ('official-yangzhou-fried-rice', '食用油', '2', 'unit-tablespoon', NULL, 8),
    ('official-yangzhou-fried-rice', '盐', '1', 'unit-teaspoon', NULL, 9),
    ('official-brown-sugar-ciba', '糍粑', '400', 'unit-gram', NULL, 1),
    ('official-brown-sugar-ciba', '红糖', '80', 'unit-gram', NULL, 2),
    ('official-brown-sugar-ciba', '清水', '100', 'unit-milliliter', NULL, 3),
    ('official-brown-sugar-ciba', '熟黄豆粉', '50', 'unit-gram', NULL, 4),
    ('official-brown-sugar-ciba', '食用油', '1', 'unit-tablespoon', NULL, 5);

DROP TEMPORARY TABLE IF EXISTS tmp_official_step_seed;
CREATE TEMPORARY TABLE tmp_official_step_seed (
    dish_public_id VARCHAR(64) NOT NULL,
    description VARCHAR(2000) NOT NULL,
    image_url VARCHAR(512) NULL,
    sort_order INT NOT NULL,
    INDEX idx_tmp_official_step_dish (dish_public_id)
);
INSERT INTO tmp_official_step_seed (
    dish_public_id, description, image_url, sort_order
) VALUES
    ('official-tomato-egg', '番茄洗净切块，小葱切葱花；鸡蛋加少许盐充分打散。', NULL, 1),
    ('official-tomato-egg', '锅烧热后放一半食用油，倒入蛋液，炒至七八成熟后盛出。', NULL, 2),
    ('official-tomato-egg', '锅中补入剩余食用油，下番茄、盐和白糖，中火炒至番茄出汁。', NULL, 3),
    ('official-tomato-egg', '倒回鸡蛋快速翻匀，让蛋块吸收汤汁，撒葱花后立即出锅。', NULL, 4),
    ('official-pepper-pork', '猪里脊顺纹切丝，加入料酒、生抽和淀粉抓匀，腌制十分钟。', NULL, 1),
    ('official-pepper-pork', '青椒去籽切丝，姜切细丝备用。', NULL, 2),
    ('official-pepper-pork', '热锅放油，下肉丝快速滑散，肉丝变色后先盛出。', NULL, 3),
    ('official-pepper-pork', '锅中下姜丝和青椒炒至断生，倒回肉丝，加盐，大火翻匀出锅。', NULL, 4),
    ('official-kung-pao-chicken', '鸡腿肉切丁，加少量生抽和淀粉抓匀；大葱切段，干辣椒剪段。', NULL, 1),
    ('official-kung-pao-chicken', '将剩余生抽、香醋、白糖和少量清水调成碗汁。', NULL, 2),
    ('official-kung-pao-chicken', '热锅放油，下花椒和干辣椒小火煸香，再放鸡丁大火炒至变色。', NULL, 3),
    ('official-kung-pao-chicken', '加入葱段翻炒，沿锅边倒入碗汁，快速收汁。', NULL, 4),
    ('official-kung-pao-chicken', '关火后拌入熟花生米，保持花生酥脆，立即装盘。', NULL, 5),
    ('official-garlic-broccoli', '西兰花切成大小接近的小朵，用淡盐水浸泡后冲洗干净。', NULL, 1),
    ('official-garlic-broccoli', '沸水中加少许盐和油，放西兰花焯约一分钟，捞出沥水。', NULL, 2),
    ('official-garlic-broccoli', '热锅放油，下蒜末小火炒香，注意不要炒焦。', NULL, 3),
    ('official-garlic-broccoli', '倒入西兰花和清水，大火翻炒，加盐调味后出锅。', NULL, 4),
    ('official-steamed-sea-bass', '鲈鱼处理干净，两面划浅刀，抹少许盐和料酒，腌制十分钟。', NULL, 1),
    ('official-steamed-sea-bass', '盘底铺姜片和葱段，放上鲈鱼，水沸后入锅大火蒸八至十分钟。', NULL, 2),
    ('official-steamed-sea-bass', '取出后倒掉盘中多余汤汁，去掉旧葱姜，铺上新鲜姜丝和葱丝。', NULL, 3),
    ('official-steamed-sea-bass', '淋蒸鱼豉油，再浇一勺烧热的食用油激出葱姜香味。', NULL, 4),
    ('official-garlic-shrimp', '粉丝用温水泡软后沥干，铺在盘底；鲜虾剪须、开背并去虾线。', NULL, 1),
    ('official-garlic-shrimp', '蒜切末，一半用小火炸至浅金色，与生蒜、生抽、蚝油和白糖拌匀。', NULL, 2),
    ('official-garlic-shrimp', '将鲜虾摆在粉丝上，每只虾背铺入蒜蓉，剩余料汁淋在粉丝上。', NULL, 3),
    ('official-garlic-shrimp', '水沸后入锅大火蒸六至八分钟，取出撒葱花，淋少量热油。', NULL, 4),
    ('official-cucumber-salad', '黄瓜洗净后拍裂，切成容易入口的小段。', NULL, 1),
    ('official-cucumber-salad', '加盐拌匀，静置十分钟后倒掉析出的水分。', NULL, 2),
    ('official-cucumber-salad', '大蒜切末，小米椒切圈，与香醋、生抽、香油和白糖调匀。', NULL, 3),
    ('official-cucumber-salad', '将料汁倒入黄瓜中充分拌匀，冷藏十分钟后食用风味更佳。', NULL, 4),
    ('official-white-cut-chicken', '整鸡处理干净；锅中加足量水、姜片和葱结，烧至将沸未沸。', NULL, 1),
    ('official-white-cut-chicken', '提鸡浸入热水再提起，重复三次后完全放入锅中，小火浸煮。', NULL, 2),
    ('official-white-cut-chicken', '约二十五分钟后关火焖十分钟，捞出立即放入冰水降温。', NULL, 3),
    ('official-white-cut-chicken', '鸡皮擦干后抹薄薄一层香油，斩件装盘。', NULL, 4),
    ('official-white-cut-chicken', '姜和葱切末，加盐，浇入热油调成姜葱蘸料。', NULL, 5),
    ('official-beef-potato-stew', '牛腩切块后冷水下锅，加姜和料酒焯去血沫，捞出洗净。', NULL, 1),
    ('official-beef-potato-stew', '锅中少油炒香葱姜和八角，下牛腩翻炒，加入生抽和老抽上色。', NULL, 2),
    ('official-beef-potato-stew', '加足量热水没过牛腩，小火炖约九十分钟至基本软烂。', NULL, 3),
    ('official-beef-potato-stew', '土豆和胡萝卜切块加入锅中，再炖二十分钟。', NULL, 4),
    ('official-beef-potato-stew', '根据咸度补盐，开盖收至汤汁浓度适中后出锅。', NULL, 5),
    ('official-northeast-stew', '排骨冷水下锅焯去血沫，捞出洗净；所有蔬菜切成较大的块。', NULL, 1),
    ('official-northeast-stew', '锅中放少量油炒香姜片，下排骨和番茄翻炒，加入生抽。', NULL, 2),
    ('official-northeast-stew', '加入热水没过排骨，小火先炖四十分钟。', NULL, 3),
    ('official-northeast-stew', '放入豆角、土豆、玉米和南瓜，再炖二十五分钟。', NULL, 4),
    ('official-northeast-stew', '最后按口味加盐，轻轻翻动，避免南瓜和土豆碎散。', NULL, 5),
    ('official-mapo-tofu', '豆腐切两厘米见方的小块，放入淡盐热水中浸泡五分钟后沥干。', NULL, 1),
    ('official-mapo-tofu', '花椒干锅焙香后碾碎；蒜切末，小葱切葱花。', NULL, 2),
    ('official-mapo-tofu', '锅中放油，下牛肉末炒酥，加入豆瓣酱和蒜末炒出红油。', NULL, 3),
    ('official-mapo-tofu', '加入清水和生抽，轻轻放入豆腐，小火烧五分钟。', NULL, 4),
    ('official-mapo-tofu', '分两次淋入水淀粉收汁，撒花椒粉和葱花后出锅。', NULL, 5),
    ('official-braised-pork', '五花肉切方块，冷水下锅焯去血沫，捞出擦干。', NULL, 1),
    ('official-braised-pork', '锅中不放油，小火煸出五花肉部分油脂，盛出备用。', NULL, 2),
    ('official-braised-pork', '留少量底油放冰糖，小火炒至琥珀色，倒入五花肉翻匀上色。', NULL, 3),
    ('official-braised-pork', '加入葱姜、八角、生抽、老抽和料酒，再加热水没过肉块。', NULL, 4),
    ('official-braised-pork', '小火炖约六十分钟，最后开盖收汁至红亮浓稠。', NULL, 5),
    ('official-cola-wings', '鸡翅两面各划两刀，冷水下锅焯去浮沫，捞出擦干。', NULL, 1),
    ('official-cola-wings', '平底锅放少量油，将鸡翅两面煎至微黄。', NULL, 2),
    ('official-cola-wings', '加入姜片、葱段、生抽和料酒，倒入可乐没过鸡翅大半。', NULL, 3),
    ('official-cola-wings', '中小火烧二十分钟，去掉葱姜，按口味加盐。', NULL, 4),
    ('official-cola-wings', '转大火不断翻动收汁，待酱汁能薄薄裹住鸡翅即可。', NULL, 5),
    ('official-sweet-sour-pork', '猪里脊切粗条，加少许盐和鸡蛋抓匀，腌制十分钟。', NULL, 1),
    ('official-sweet-sour-pork', '淀粉与面粉混合，逐条裹住肉条并抖去多余干粉。', NULL, 2),
    ('official-sweet-sour-pork', '油温约六成热时分批炸至定型捞出，升高油温后复炸至金黄酥脆。', NULL, 3),
    ('official-sweet-sour-pork', '锅中留少量底油，下番茄酱、白醋、白糖和少量清水熬至浓稠。', NULL, 4),
    ('official-sweet-sour-pork', '倒入炸好的里脊快速翻匀，让糖醋汁薄薄挂在表面后出锅。', NULL, 5),
    ('official-pan-fried-hairtail', '带鱼清理干净后切段，加入姜丝、葱段、料酒和盐腌制二十分钟。', NULL, 1),
    ('official-pan-fried-hairtail', '挑去葱姜，用厨房纸吸干带鱼表面水分，均匀拍一层薄淀粉。', NULL, 2),
    ('official-pan-fried-hairtail', '平底锅烧热后放油，依次放入带鱼，中小火煎至底面定型。', NULL, 3),
    ('official-pan-fried-hairtail', '翻面继续煎至两面金黄，出锅前用大火逼出多余油脂。', NULL, 4),
    ('official-honey-roast-ribs', '肋排切段后泡去血水，擦干表面；大蒜压成蒜末。', NULL, 1),
    ('official-honey-roast-ribs', '叉烧酱、生抽、蚝油、蒜末和料酒调匀，抹在排骨上冷藏腌制四小时。', NULL, 2),
    ('official-honey-roast-ribs', '排骨放在铺锡纸的烤盘上，盖锡纸，以一百八十度烤三十分钟。', NULL, 3),
    ('official-honey-roast-ribs', '去掉上层锡纸，刷一层蜂蜜和腌料，再烤十五分钟。', NULL, 4),
    ('official-honey-roast-ribs', '中途翻面并再次刷汁，烤至边缘微焦、内部熟透后取出。', NULL, 5),
    ('official-seaweed-egg-soup', '紫菜撕成小片，小葱切葱花，鸡蛋充分打散。', NULL, 1),
    ('official-seaweed-egg-soup', '锅中加清水和虾皮烧开，放入紫菜煮约半分钟。', NULL, 2),
    ('official-seaweed-egg-soup', '保持汤面微沸，沿锅边缓慢淋入蛋液，静置数秒后轻轻推动形成蛋花。', NULL, 3),
    ('official-seaweed-egg-soup', '加盐调味，关火后淋香油并撒葱花。', NULL, 4),
    ('official-corn-rib-soup', '排骨冷水下锅，加料酒和几片姜，煮开后撇去浮沫，捞出洗净。', NULL, 1),
    ('official-corn-rib-soup', '玉米切段，胡萝卜切滚刀块，小葱打结。', NULL, 2),
    ('official-corn-rib-soup', '排骨、姜片、葱结和清水放入汤锅，大火烧开后转小火炖四十分钟。', NULL, 3),
    ('official-corn-rib-soup', '加入玉米和胡萝卜继续炖三十分钟。', NULL, 4),
    ('official-corn-rib-soup', '捞出葱结，按口味加盐，静置片刻后盛出。', NULL, 5),
    ('official-yangzhou-fried-rice', '米饭提前打散；火腿和胡萝卜切小丁，虾仁去虾线，小葱切花。', NULL, 1),
    ('official-yangzhou-fried-rice', '青豆和胡萝卜焯水后沥干；鸡蛋打散。', NULL, 2),
    ('official-yangzhou-fried-rice', '热锅放油，先炒散鸡蛋，再放虾仁和火腿丁炒香。', NULL, 3),
    ('official-yangzhou-fried-rice', '加入米饭大火翻炒至米粒松散，再放青豆和胡萝卜。', NULL, 4),
    ('official-yangzhou-fried-rice', '加盐调味，持续翻炒至锅气充足，最后撒葱花出锅。', NULL, 5),
    ('official-brown-sugar-ciba', '糍粑切成大小均匀的长条或方块，表面擦干。', NULL, 1),
    ('official-brown-sugar-ciba', '平底锅刷一层薄油，小火将糍粑各面煎至微黄、内部变软。', NULL, 2),
    ('official-brown-sugar-ciba', '红糖和清水放入小锅，小火熬至红糖完全融化、糖汁略微浓稠。', NULL, 3),
    ('official-brown-sugar-ciba', '煎好的糍粑装盘，淋红糖汁，最后撒熟黄豆粉。', NULL, 4);

DROP TEMPORARY TABLE IF EXISTS tmp_official_tag_seed;
CREATE TEMPORARY TABLE tmp_official_tag_seed (
    dish_public_id VARCHAR(64) NOT NULL,
    tag_public_id VARCHAR(64) NOT NULL,
    PRIMARY KEY (dish_public_id, tag_public_id)
);
INSERT INTO tmp_official_tag_seed (dish_public_id, tag_public_id) VALUES
    ('official-tomato-egg', 'tag-cuisine-home-style'),
    ('official-tomato-egg', 'tag-feature-quick'),
    ('official-tomato-egg', 'tag-spice-none'),
    ('official-pepper-pork', 'tag-cuisine-home-style'),
    ('official-pepper-pork', 'tag-feature-rice-friendly'),
    ('official-pepper-pork', 'tag-spice-mild'),
    ('official-kung-pao-chicken', 'tag-cuisine-sichuan'),
    ('official-kung-pao-chicken', 'tag-flavor-numbing-spicy'),
    ('official-kung-pao-chicken', 'tag-feature-rice-friendly'),
    ('official-garlic-broccoli', 'tag-cuisine-home-style'),
    ('official-garlic-broccoli', 'tag-feature-vegetarian'),
    ('official-garlic-broccoli', 'tag-spice-none'),
    ('official-steamed-sea-bass', 'tag-cuisine-cantonese'),
    ('official-steamed-sea-bass', 'tag-flavor-light'),
    ('official-steamed-sea-bass', 'tag-spice-none'),
    ('official-garlic-shrimp', 'tag-cuisine-cantonese'),
    ('official-garlic-shrimp', 'tag-flavor-savory'),
    ('official-garlic-shrimp', 'tag-spice-none'),
    ('official-cucumber-salad', 'tag-cuisine-home-style'),
    ('official-cucumber-salad', 'tag-feature-quick'),
    ('official-cucumber-salad', 'tag-feature-vegetarian'),
    ('official-white-cut-chicken', 'tag-cuisine-cantonese'),
    ('official-white-cut-chicken', 'tag-flavor-light'),
    ('official-white-cut-chicken', 'tag-spice-none'),
    ('official-beef-potato-stew', 'tag-cuisine-home-style'),
    ('official-beef-potato-stew', 'tag-feature-rice-friendly'),
    ('official-beef-potato-stew', 'tag-spice-none'),
    ('official-northeast-stew', 'tag-cuisine-northeast'),
    ('official-northeast-stew', 'tag-cuisine-home-style'),
    ('official-northeast-stew', 'tag-feature-rice-friendly'),
    ('official-mapo-tofu', 'tag-cuisine-sichuan'),
    ('official-mapo-tofu', 'tag-flavor-numbing-spicy'),
    ('official-mapo-tofu', 'tag-feature-rice-friendly'),
    ('official-braised-pork', 'tag-cuisine-home-style'),
    ('official-braised-pork', 'tag-flavor-savory'),
    ('official-braised-pork', 'tag-feature-rice-friendly'),
    ('official-cola-wings', 'tag-cuisine-home-style'),
    ('official-cola-wings', 'tag-feature-quick'),
    ('official-cola-wings', 'tag-spice-none'),
    ('official-sweet-sour-pork', 'tag-cuisine-shandong'),
    ('official-sweet-sour-pork', 'tag-flavor-sweet-sour'),
    ('official-sweet-sour-pork', 'tag-spice-none'),
    ('official-pan-fried-hairtail', 'tag-cuisine-home-style'),
    ('official-pan-fried-hairtail', 'tag-flavor-savory'),
    ('official-pan-fried-hairtail', 'tag-spice-none'),
    ('official-honey-roast-ribs', 'tag-cuisine-cantonese'),
    ('official-honey-roast-ribs', 'tag-flavor-savory'),
    ('official-honey-roast-ribs', 'tag-spice-none'),
    ('official-seaweed-egg-soup', 'tag-cuisine-home-style'),
    ('official-seaweed-egg-soup', 'tag-feature-quick'),
    ('official-seaweed-egg-soup', 'tag-flavor-light'),
    ('official-corn-rib-soup', 'tag-cuisine-home-style'),
    ('official-corn-rib-soup', 'tag-flavor-light'),
    ('official-corn-rib-soup', 'tag-spice-none'),
    ('official-yangzhou-fried-rice', 'tag-cuisine-jiangsu'),
    ('official-yangzhou-fried-rice', 'tag-flavor-savory'),
    ('official-yangzhou-fried-rice', 'tag-feature-quick'),
    ('official-brown-sugar-ciba', 'tag-cuisine-sichuan'),
    ('official-brown-sugar-ciba', 'tag-cuisine-home-style'),
    ('official-brown-sugar-ciba', 'tag-spice-none');

SET @seed_admin_id := (
    SELECT id
    FROM sys_users
    WHERE deleted_at IS NULL AND enable = 1
    ORDER BY id ASC
    LIMIT 1
);
SET @seed_admin_username := (
    SELECT username FROM sys_users WHERE id = @seed_admin_id LIMIT 1
);
SET @seed_admin_nickname := (
    SELECT nick_name FROM sys_users WHERE id = @seed_admin_id LIMIT 1
);

SET @seed_category_expected := (
    SELECT COUNT(DISTINCT category_public_id) FROM tmp_official_dish_seed
);
SET @seed_category_actual := (
    SELECT COUNT(DISTINCT category.public_id)
    FROM of_categories AS category
    INNER JOIN tmp_official_dish_seed AS seed
        ON seed.category_public_id = category.public_id
    WHERE category.deleted_at IS NULL AND category.enabled = 1
);
SET @seed_tag_expected := (
    SELECT COUNT(DISTINCT tag_public_id) FROM tmp_official_tag_seed
);
SET @seed_tag_actual := (
    SELECT COUNT(DISTINCT tag.public_id)
    FROM of_tags AS tag
    INNER JOIN tmp_official_tag_seed AS seed
        ON seed.tag_public_id = tag.public_id
    WHERE tag.deleted_at IS NULL AND tag.enabled = 1
);
SET @seed_unit_expected := (
    SELECT COUNT(DISTINCT unit_public_id)
    FROM tmp_official_ingredient_seed
    WHERE unit_public_id IS NOT NULL
);
SET @seed_unit_actual := (
    SELECT COUNT(DISTINCT unit.public_id)
    FROM of_units AS unit
    INNER JOIN tmp_official_ingredient_seed AS seed
        ON seed.unit_public_id = unit.public_id
    WHERE unit.deleted_at IS NULL AND unit.enabled = 1
);
SET @seed_ready := (
    @seed_admin_id IS NOT NULL
    AND @seed_category_actual = @seed_category_expected
    AND @seed_tag_actual = @seed_tag_expected
    AND @seed_unit_actual = @seed_unit_expected
);

INSERT INTO of_media (
    id, user_id,
    administrator_id, administrator_username, administrator_nickname,
    file_name, upload_source, scene, resource_status,
    url, storage_path, content_type,
    width, height, size_bytes, checksum,
    review_status, reject_reason,
    deleted_at, delete_reason, created_at
)
SELECT
    seed.id, '',
    @seed_admin_id, @seed_admin_username, NULLIF(@seed_admin_nickname, ''),
    seed.file_name, 'admin_upload', 'official_dish_cover', 'active',
    seed.url, seed.storage_path, 'image/jpeg',
    seed.width, seed.height, seed.size_bytes, seed.checksum,
    'not_required', NULL,
    NULL, NULL, NOW()
FROM tmp_official_media_seed AS seed
WHERE @seed_ready = 1
  AND NOT EXISTS (
      SELECT 1 FROM of_media AS existing WHERE existing.id = seed.id
  );

INSERT INTO of_off_dishes (
    created_at, updated_at, deleted_at,
    public_id, name, cover_file_id, cover_url,
    category_id, serving, description, status,
    recommendation_count, version,
    created_by_id, created_by_username, created_by_nickname,
    updated_by_id, updated_by_username, updated_by_nickname,
    deleted_reason, deleted_by_id
)
SELECT
    NOW(), NOW(), NULL,
    seed.public_id, seed.name, seed.cover_file_id, media.url,
    category.id, seed.serving, seed.description, 'usable',
    0, 1,
    @seed_admin_id, @seed_admin_username, NULLIF(@seed_admin_nickname, ''),
    @seed_admin_id, @seed_admin_username, NULLIF(@seed_admin_nickname, ''),
    NULL, NULL
FROM tmp_official_dish_seed AS seed
INNER JOIN of_categories AS category
    ON category.public_id = seed.category_public_id
   AND category.deleted_at IS NULL
   AND category.enabled = 1
INNER JOIN of_media AS media
    ON media.id = seed.cover_file_id
   AND media.deleted_at IS NULL
   AND media.upload_source = 'admin_upload'
   AND media.scene = 'official_dish_cover'
   AND media.resource_status = 'active'
   AND media.review_status = 'not_required'
WHERE @seed_ready = 1
  AND NOT EXISTS (
      SELECT 1
      FROM of_off_dishes AS existing
      WHERE existing.public_id = seed.public_id
         OR (existing.deleted_at IS NULL AND existing.name = seed.name)
  );

INSERT INTO of_off_ingredients (
    created_at, updated_at, deleted_at,
    dish_id, name, quantity, unit_id, note, sort_order
)
SELECT
    NOW(), NOW(), NULL,
    dish.id, seed.name, seed.quantity, unit.id, seed.note, seed.sort_order
FROM tmp_official_ingredient_seed AS seed
INNER JOIN of_off_dishes AS dish
    ON dish.public_id = seed.dish_public_id
   AND dish.deleted_at IS NULL
LEFT JOIN of_units AS unit
    ON unit.public_id = seed.unit_public_id
   AND unit.deleted_at IS NULL
   AND unit.enabled = 1
WHERE @seed_ready = 1
  AND (seed.unit_public_id IS NULL OR unit.id IS NOT NULL)
  AND NOT EXISTS (
      SELECT 1
      FROM of_off_ingredients AS existing
      WHERE existing.dish_id = dish.id
  );

INSERT INTO of_off_steps (
    created_at, updated_at, deleted_at,
    dish_id, description, image_url, sort_order
)
SELECT
    NOW(), NOW(), NULL,
    dish.id, seed.description, seed.image_url, seed.sort_order
FROM tmp_official_step_seed AS seed
INNER JOIN of_off_dishes AS dish
    ON dish.public_id = seed.dish_public_id
   AND dish.deleted_at IS NULL
WHERE @seed_ready = 1
  AND NOT EXISTS (
      SELECT 1
      FROM of_off_steps AS existing
      WHERE existing.dish_id = dish.id
  );

INSERT INTO of_off_tags (dish_id, tag_id, created_at)
SELECT dish.id, tag.id, NOW()
FROM tmp_official_tag_seed AS seed
INNER JOIN of_off_dishes AS dish
    ON dish.public_id = seed.dish_public_id
   AND dish.deleted_at IS NULL
INNER JOIN of_tags AS tag
    ON tag.public_id = seed.tag_public_id
   AND tag.deleted_at IS NULL
   AND tag.enabled = 1
WHERE @seed_ready = 1
  AND NOT EXISTS (
      SELECT 1
      FROM of_off_tags AS existing
      WHERE existing.dish_id = dish.id
        AND existing.tag_id = tag.id
  );

COMMIT;

SELECT
    @seed_ready AS prerequisite_ready,
    @seed_admin_id AS administrator_id,
    @seed_category_expected AS expected_categories,
    @seed_category_actual AS actual_categories,
    @seed_tag_expected AS expected_tags,
    @seed_tag_actual AS actual_tags,
    @seed_unit_expected AS expected_units,
    @seed_unit_actual AS actual_units;

SELECT
    20 AS expected,
    COUNT(*) AS actual,
    'of_media' AS item
FROM of_media
WHERE id IN (SELECT id FROM tmp_official_media_seed)
UNION ALL
SELECT
    20 AS expected,
    COUNT(*) AS actual,
    'of_off_dishes' AS item
FROM of_off_dishes
WHERE deleted_at IS NULL
  AND public_id IN (SELECT public_id FROM tmp_official_dish_seed)
UNION ALL
SELECT
    (SELECT COUNT(*) FROM tmp_official_ingredient_seed) AS expected,
    COUNT(*) AS actual,
    'of_off_ingredients' AS item
FROM of_off_ingredients
WHERE deleted_at IS NULL
  AND dish_id IN (
      SELECT id FROM of_off_dishes
      WHERE public_id IN (SELECT public_id FROM tmp_official_dish_seed)
  )
UNION ALL
SELECT
    (SELECT COUNT(*) FROM tmp_official_step_seed) AS expected,
    COUNT(*) AS actual,
    'of_off_steps' AS item
FROM of_off_steps
WHERE deleted_at IS NULL
  AND dish_id IN (
      SELECT id FROM of_off_dishes
      WHERE public_id IN (SELECT public_id FROM tmp_official_dish_seed)
  )
UNION ALL
SELECT
    (SELECT COUNT(*) FROM tmp_official_tag_seed) AS expected,
    COUNT(*) AS actual,
    'of_off_tags' AS item
FROM of_off_tags
WHERE dish_id IN (
    SELECT id FROM of_off_dishes
    WHERE public_id IN (SELECT public_id FROM tmp_official_dish_seed)
);