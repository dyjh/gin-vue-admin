const dishes = [
  {
    id: "dish-1",
    name: "香菇鸡腿饭",
    category: "主食",
    meta: "主食 · 家常菜",
    description: "鸡腿与香菇焖进米饭，咸香温和，适合工作日晚餐。",
    tags: ["下饭", "快手", "少油"],
    serving: 2,
    status: "usable",
    discoverable: false,
    image: "/assets/images/dish-detail-cover-realistic.jpg",
    ingredients: [
      { id: "i-1", name: "鸡腿肉", amount: "220 克" },
      { id: "i-2", name: "香菇", amount: "120 克" },
      { id: "i-3", name: "大米", amount: "160 克" }
    ],
    steps: [
      { id: "s-1", text: "鸡腿切块，香菇切片，先把食材处理好。", image: "/assets/images/dish-detail-step-1-realistic.jpg" },
      { id: "s-2", text: "锅中少油煎香鸡腿，加入香菇翻炒。", image: "" },
      { id: "s-3", text: "与大米一起焖熟，出锅前撒少许葱花。", image: "" }
    ]
  },
  {
    id: "dish-2",
    name: "番茄炒蛋",
    category: "家常菜",
    meta: "家常菜 · 快手",
    description: "酸甜开胃，十几分钟就能端上桌。",
    tags: ["下饭", "快手"],
    serving: 2,
    status: "usable",
    discoverable: true,
    image: "/assets/images/recommended-tomato-egg.jpg",
    ingredients: [
      { id: "i-4", name: "番茄", amount: "2 个" },
      { id: "i-5", name: "鸡蛋", amount: "3 个" }
    ],
    steps: [
      { id: "s-4", text: "番茄切块，鸡蛋加少许盐打散。", image: "" },
      { id: "s-5", text: "先炒鸡蛋盛出，再炒番茄出汁后合炒。", image: "" }
    ]
  },
  {
    id: "dish-3",
    name: "蒜蓉西兰花",
    category: "素菜",
    meta: "素菜 · 少油",
    description: "清爽脆嫩，适合作为一餐里的绿色蔬菜。",
    tags: ["清淡", "快手"],
    serving: 2,
    status: "usable",
    discoverable: false,
    image: "/assets/images/recommended-garlic-broccoli.jpg",
    ingredients: [
      { id: "i-6", name: "西兰花", amount: "1 颗" },
      { id: "i-7", name: "蒜", amount: "3 瓣" }
    ],
    steps: [
      { id: "s-6", text: "西兰花切小朵焯水，蒜切末。", image: "" },
      { id: "s-7", text: "蒜末爆香后放入西兰花快速翻炒。", image: "" }
    ]
  },
  {
    id: "dish-4",
    name: "冬瓜丸子汤",
    category: "汤菜",
    meta: "汤菜 · 适合多人",
    description: "清淡暖胃，能提前处理食材。",
    tags: ["清淡", "可提前备"],
    serving: 3,
    status: "usable",
    discoverable: false,
    image: "/assets/images/recommended-winter-melon-soup.jpg",
    ingredients: [
      { id: "i-8", name: "冬瓜", amount: "400 克" },
      { id: "i-9", name: "猪肉馅", amount: "220 克" }
    ],
    steps: [
      { id: "s-8", text: "冬瓜切片，肉馅调味后挤成丸子。", image: "" },
      { id: "s-9", text: "水开后下丸子，定型后放冬瓜煮熟。", image: "" }
    ]
  },
  {
    id: "dish-5",
    name: "青椒牛柳",
    category: "家常菜",
    meta: "家常菜 · 20 分钟",
    description: "牛肉嫩滑，青椒清香，是稳定的下饭菜。",
    tags: ["下饭", "少油"],
    serving: 2,
    status: "usable",
    discoverable: true,
    image: "/assets/images/recommended-green-pepper-beef.jpg",
    ingredients: [
      { id: "i-10", name: "牛里脊", amount: "250 克" },
      { id: "i-11", name: "青椒", amount: "2 个" }
    ],
    steps: [
      { id: "s-10", text: "牛肉逆纹切条并简单腌制，青椒切丝。", image: "/assets/images/recommended-green-pepper-beef-step.jpg" },
      { id: "s-11", text: "牛肉滑炒变色后盛出，再与青椒快速合炒。", image: "" }
    ]
  }
];

const recommendations = dishes.map((dish, index) => ({
  ...dish,
  id: `recommend-${index + 1}`,
  sourceDishId: dish.id,
  sourceType: index === 0 ? "official" : "creator",
  author: index === 0 ? "来干饭官方" : ["晚饭研究所", "小满厨房", "认真吃饭", "家常味道"][index - 1],
  copied: false
}));

const recipes = [
  {
    id: "recipe-1",
    name: "工作日晚餐",
    note: "快手、少油，30 分钟内",
    dishIds: ["dish-5", "dish-2", "dish-3", "dish-4"],
    coverImages: [
      "/assets/images/recommended-green-pepper-beef.jpg",
      "/assets/images/recommended-tomato-egg.jpg",
      "/assets/images/recommended-garlic-broccoli.jpg"
    ]
  },
  {
    id: "recipe-2",
    name: "周末慢慢做",
    note: "有汤有菜，适合一家人",
    dishIds: ["dish-1", "dish-4", "dish-2"],
    coverImages: [
      "/assets/images/dish-detail-cover-realistic.jpg",
      "/assets/images/recommended-winter-melon-soup.jpg",
      "/assets/images/recommended-tomato-egg.jpg"
    ]
  }
];

const checkins = [
  { id: "checkin-1", dishId: "dish-2", dishName: "番茄炒蛋", date: "2026-07-18", image: "/assets/images/recommended-tomato-egg.jpg", note: "今天番茄多炒了一会，更入味。", rewarded: true },
  { id: "checkin-2", dishId: "dish-4", dishName: "冬瓜丸子汤", date: "2026-07-15", image: "/assets/images/recommended-winter-melon-soup.jpg", note: "汤很清爽。", rewarded: true }
];

const shoppingItems = [
  { id: "shop-1", name: "土豆", amount: "4 个", note: "来自土豆丝", completed: false },
  { id: "shop-2", name: "青椒", amount: "3 个", note: "来自青椒牛柳", completed: false },
  { id: "shop-3", name: "西兰花", amount: "2 颗", note: "来自蒜蓉西兰花", completed: false },
  { id: "shop-4", name: "鸡腿肉", amount: "600 克", note: "来自香菇鸡腿饭", completed: false },
  { id: "shop-5", name: "鸡蛋", amount: "6 个", note: "来自番茄炒蛋", completed: true },
  { id: "shop-6", name: "大米", amount: "400 克", note: "来自香菇鸡腿饭", completed: true }
];

const pointEntries = Array.from({ length: 46 }, (_, index) => {
  const kinds = [
    { type: "earned", title: "每日首次做菜打卡", description: "完成做菜打卡", amount: 5 },
    { type: "spent", title: "AI 不知道吃什么", description: "生成个性化推荐", amount: -8 },
    { type: "earned", title: "系统发放", description: "内容贡献奖励", amount: 10 },
    { type: "refund", title: "AI 失败退还", description: "关联原消费记录", amount: 8 }
  ];
  const item = kinds[index % kinds.length];
  return {
    id: `point-${index + 1}`,
    ...item,
    createdAt: `2026-07-${String(18 - (index % 15)).padStart(2, "0")} ${String(20 - (index % 10)).padStart(2, "0")}:30`
  };
});

const notifications = Array.from({ length: 33 }, (_, index) => {
  const items = [
    { type: "governance", title: "菜品已取消推荐", content: "“番茄炒蛋”已从菜品推荐中移除，仍保留在你的菜品库。" },
    { type: "discover", title: "允许被发现已关闭", content: "“香菇鸡腿饭”当前为未公开状态。" },
    { type: "points", title: "积分到账", content: "每日首次做菜打卡获得 5 积分。" },
    { type: "refund", title: "AI 失败积分已退还", content: "本次生成未完成，8 积分已退回。" },
    { type: "meal", title: "饭局状态有变化", content: "“周六家庭聚餐”已确认菜单，可以查看采购清单。" }
  ];
  return {
    id: `notice-${index + 1}`,
    ...items[index % items.length],
    read: index > 4,
    createdAt: index < 5 ? "今天 10:20" : "07-17 18:40"
  };
});

const seed = {
  profile: {
    nickname: "厨房小记",
    avatarUrl: "",
    points: 128,
    checkinDays: 18
  },
  dishes,
  recommendations,
  recipes,
  checkins,
  meal: {
    id: "meal-1",
    name: "周六家庭聚餐",
    code: "638214",
    status: "collecting",
    deadline: "今天 21:00",
    participantCount: 4,
    candidateIds: ["dish-1", "dish-2", "dish-3", "dish-4"],
    selectedIds: ["dish-1", "dish-3"],
    votes: { "dish-1": 3, "dish-2": 2, "dish-3": 2, "dish-4": 1 }
  },
  shoppingItems,
  pointEntries,
  notifications,
  aiUsage: [
    { id: "ai-1", name: "不知道吃什么", cost: 8, createdAt: "今天 12:20", status: "success" },
    { id: "ai-2", name: "AI 备菜提醒", cost: 6, createdAt: "07-16 17:40", status: "success" }
  ]
};

module.exports = { seed };
