const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");

const pages = [
  ["home", "pages/home/home"],
  ["dishAdd", "pages/dish/add"],
  ["dishAiParse", "pages/dish/ai-parse"],
  ["dishImage", "pages/dish/image-generate"],
  ["dishDetail", "pages/dish/detail"],
  ["dishSearch", "pages/dish/search"],
  ["dishRecommendations", "pages/dish/recommendations"],
  ["recipeList", "pages/recipe/list"],
  ["recipeEdit", "pages/recipe/edit"],
  ["recipeDetail", "pages/recipe/detail"],
  ["profile", "pages/profile/profile"],
  ["points", "pages/profile/points"],
  ["aiRecords", "pages/profile/ai-records"],
  ["checkin", "pages/profile/checkin"],
  ["notifications", "pages/profile/notifications"],
  ["parties", "pages/profile/parties"],
  ["partyCreate", "pages/party/create"],
  ["partyCode", "pages/party/code"],
  ["partyJoin", "pages/party/join"],
  ["partyOrder", "pages/party/order"],
  ["partySummary", "pages/party/summary"],
  ["purchaseList", "pages/purchase/list"],
  ["aiPrepTips", "pages/ai/prep-tips"],
  ["whatToEat", "pages/ai/what-to-eat"],
  ["disabled", "pages/system/disabled"]
];

const routeOf = Object.fromEntries(pages.map(([key, route]) => [key, `/${route}`]));

function row(title, desc, badge, mark, tone = "green", badgeTone = "green", action = "") {
  return { title, desc, badge, mark, tone, badgeTone, action };
}

function recommend(title, desc, mark, tone = "green") {
  return { title, desc, mark, tone, action: "加入菜品库" };
}

function field(label, value, large = false) {
  return { label, value, large };
}

function stat(value, label) {
  return { value, label };
}

function tile(title, desc, chips = []) {
  return { title, desc, chips };
}

function section(id, type, title, data = {}) {
  return { id, type, title, ...data };
}

function chips(items) {
  return items.map((text, index) => ({
    text,
    tone: index === 1 ? "orange" : index === 2 ? "blue" : "green"
  }));
}

const myDishes = [
  row("清蒸鲈鱼", "清淡 · 适合晚饭", "已完善", "鱼", "blue"),
  row("菌菇鸡汤", "汤菜 · 可提前准备", "菜谱内", "汤", "orange", "blue"),
  row("蒜蓉西兰花", "快手 · 素菜", "常点", "菜", "green", "orange")
];

const recDishes = [
  recommend("虾仁蒸蛋", "公开精选 · 适合孩子", "虾", "coral"),
  recommend("南瓜小米粥", "清淡暖胃 · 早餐晚餐都合适", "粥", "yellow"),
  recommend("青椒牛柳", "下饭快手 · 家常热菜", "牛", "coral"),
  recommend("冬瓜丸子汤", "汤菜 · 适合多人", "汤", "blue"),
  recommend("凉拌黄瓜", "爽口素菜 · 可提前准备", "瓜", "green")
];

const screens = {
  home: {
    navTitle: "首页",
    tab: "home",
    hero: {
      title: "来干饭",
      desc: "把爱吃的菜，整理成自己的菜品库"
    },
    search: "搜索菜名、分类、食材",
    actions: [
      { text: "添加菜品", variant: "primary", to: routeOf.dishAdd },
      { text: "AI 识别", variant: "ghost", to: routeOf.dishAiParse }
    ],
    chips: chips(["手动录入", "生成菜品图", "分类筛选"]),
    sections: [
      section("home-dishes", "rows", "我的菜品", { items: myDishes }),
      section("home-rec", "recommend", "菜品推荐", {
        more: "查看更多",
        moreTo: routeOf.dishRecommendations,
        items: recDishes
      })
    ],
    bottomAI: true
  },
  dishAdd: {
    navTitle: "添加菜品",
    back: true,
    titleBlock: {
      title: "添加菜品",
      desc: "先把关键信息记下来，缺的部分之后再补。"
    },
    sections: [
      section("dish-add-form", "fields", "", {
        fields: [
          field("菜名", "清炒荷兰豆"),
          field("分类", "素菜"),
          field("基础份量", "2 人份"),
          field("标签", "快手 / 清淡 / 适合晚饭"),
          field("配料", "荷兰豆 300g，蒜 2 瓣，盐 少许", true),
          field("做法", "去筋洗净，热锅快炒，出锅前调味。", true),
          field("备注", "口感保持脆嫩，少放油。")
        ]
      })
    ],
    bottomActions: [
      { text: "保存草稿", variant: "ghost" },
      { text: "保存为可用", variant: "primary" }
    ]
  },
  dishAiParse: {
    navTitle: "AI 录入",
    back: true,
    titleBlock: {
      title: "AI 辅助录入",
      desc: "粘贴菜品文字，或上传长截图，让系统先生成草稿。"
    },
    sections: [
      section("ai-entry", "tiles", "", {
        items: [
          tile("粘贴文字", "适合从聊天、网页、备忘录复制来的做法。", ["文本解析"]),
          tile("上传长截图", "适合菜谱 App 或网页长图。", ["图片理解"])
        ]
      }),
      section("ai-preview", "fields", "解析预览", {
        more: "可手动修改",
        fields: [
          field("识别菜名", "番茄牛腩"),
          field("分类", "荤菜"),
          field("份量", "3 人份"),
          field("需要补全", "配料数量、收汁时间", true)
        ]
      })
    ],
    bottomActions: [
      { text: "重新解析", variant: "ghost" },
      { text: "填入草稿", variant: "primary" }
    ]
  },
  dishImage: {
    navTitle: "生成菜品图",
    back: true,
    titleBlock: {
      title: "生成菜品图",
      desc: "根据菜名和做法生成一张默认展示图。"
    },
    sections: [
      section("image-mock", "imageMock", ""),
      section("image-fields", "fields", "", {
        fields: [
          field("生成依据", "番茄牛腩，汤汁浓郁，家常晚饭风格", true),
          field("图片风格", "自然光 / 家常餐桌 / 浅色背景")
        ]
      })
    ],
    bottomActions: [
      { text: "换一张", variant: "ghost" },
      { text: "采用图片", variant: "primary" }
    ]
  },
  dishDetail: {
    navTitle: "菜品详情",
    back: true,
    detailHero: {
      title: "清蒸鲈鱼",
      desc: "清淡 · 2 人份 · 可用"
    },
    chips: chips(["清淡", "适合晚饭", "孩子爱吃"]),
    sections: [
      section("dish-ingredients", "purchaseRows", "配料", {
        items: [
          row("鲈鱼", "1 条，处理干净", "主料", "鱼", "blue", "gray"),
          row("姜葱", "适量，去腥增香", "辅料", "葱", "green", "gray")
        ]
      }),
      section("dish-method", "fields", "做法", {
        fields: [field("步骤", "鱼身划刀，铺姜葱，上锅蒸 8 分钟；出锅淋热油和蒸鱼豉油。", true)]
      })
    ],
    bottomActions: [
      { text: "加入菜谱", variant: "ghost" },
      { text: "编辑菜品", variant: "primary", to: routeOf.dishAdd }
    ]
  },
  dishSearch: {
    navTitle: "搜索菜品",
    back: true,
    search: "搜索菜名、分类、食材",
    chips: chips(["全部", "荤菜", "素菜", "汤菜", "快手"]),
    sections: [
      section("search-result", "rows", "筛选结果", {
        items: [
          row("蒜香鸡翅", "荤菜 · 下饭 · 常点", "查看", "翅", "coral", "gray"),
          row("山药排骨汤", "汤菜 · 适合周末", "查看", "汤", "orange", "gray"),
          row("清炒芦笋", "素菜 · 快手", "查看", "菜", "green", "gray"),
          row("虾皮蒸蛋", "蛋奶 · 适合孩子", "查看", "蛋", "yellow", "gray")
        ]
      })
    ]
  },
  dishRecommendations: {
    navTitle: "菜品推荐",
    back: true,
    titleBlock: {
      title: "菜品推荐",
      desc: "后台精选公开菜品，可复制到自己的菜品库。"
    },
    sections: [
      section("rec-all", "recommend", "", { items: recDishes })
    ],
    toast: "已加入菜品库"
  },
  recipeList: {
    navTitle: "菜谱",
    tab: "recipe",
    hero: {
      title: "我的菜谱",
      desc: "把自己的菜品整理成不同场景"
    },
    search: "搜索菜谱名称、备注",
    actions: [
      { text: "新建菜谱", variant: "primary", to: routeOf.recipeEdit },
      { text: "从菜品选", variant: "ghost", to: routeOf.dishSearch }
    ],
    sections: [
      section("recipe-tiles", "tiles", "", {
        items: [
          tile("家常晚饭", "6 道菜 · 适合工作日", ["清淡", "快手"]),
          tile("孩子爱吃", "4 道菜 · 少辣少油", ["蒸煮"]),
          tile("周末小聚", "8 道菜 · 有汤有硬菜", ["多人"]),
          tile("清淡少油", "5 道菜 · 适合长辈", ["软烂"])
        ]
      })
    ]
  },
  recipeEdit: {
    navTitle: "新建菜谱",
    back: true,
    titleBlock: {
      title: "新建菜谱",
      desc: "菜谱是自己的菜品合集，可以写一点口味备注。"
    },
    sections: [
      section("recipe-form", "fields", "", {
        fields: [
          field("菜谱名称", "家常晚饭"),
          field("适用场景", "工作日晚餐，2 到 3 人"),
          field("备注", "少油，优先快手菜，汤菜可提前准备。", true)
        ]
      }),
      section("recipe-dishes", "rows", "已选菜品", { more: "添加菜品", moreTo: routeOf.dishSearch, items: myDishes })
    ],
    bottomActions: [
      { text: "保存草稿", variant: "ghost" },
      { text: "保存菜谱", variant: "primary" }
    ]
  },
  recipeDetail: {
    navTitle: "菜谱详情",
    back: true,
    titleBlock: {
      title: "家常晚饭",
      desc: "6 道菜 · 备注：清淡、快手、适合 2 到 3 人。"
    },
    sections: [
      section("recipe-stats", "stats", "", { items: [stat("6", "菜品"), stat("2", "汤菜"), stat("25", "分钟")] }),
      section("recipe-list", "rows", "菜品合集", {
        items: [
          row("清蒸鲈鱼", "清淡 · 2 人份", "查看", "鱼", "blue", "gray"),
          row("菌菇鸡汤", "汤菜 · 可提前准备", "查看", "汤", "orange", "gray"),
          row("蒜蓉西兰花", "快手 · 素菜", "查看", "菜", "green", "gray")
        ]
      })
    ],
    bottomActions: [
      { text: "编辑菜谱", variant: "ghost", to: routeOf.recipeEdit },
      { text: "用于饭局", variant: "primary", to: routeOf.partyCreate }
    ]
  },
  profile: {
    navTitle: "我的",
    tab: "profile",
    profileCard: true,
    sections: [
      section("profile-stats", "stats", "", { items: [stat("18", "菜品"), stat("4", "菜谱"), stat("3", "饭局")] }),
      section("party-entry", "tiles", "饭局入口", {
        items: [
          tile("创建饭局", "生成点餐码，收集大家想吃什么。", []),
          tile("加入饭局", "输入点餐码，选择想吃的菜。", [])
        ]
      }),
      section("profile-list", "purchaseRows", "常用功能", {
        items: [
          row("做菜打卡", "记录做过的菜，积累偏好", "去看看", "打", "green", "gray"),
          row("采购清单", "查看已生成的采购任务", "打开", "购", "blue", "gray"),
          row("通知中心", "查看审核、积分和饭局提醒", "2 条", "信", "orange", "orange")
        ]
      })
    ]
  },
  points: {
    navTitle: "积分流水",
    back: true,
    titleBlock: {
      title: "当前积分 128",
      desc: "积分用于部分 AI 能力，失败会自动退还。"
    },
    sections: [
      section("points-list", "purchaseRows", "本月记录", {
        items: [
          row("做菜打卡奖励", "7 月 8 日 19:21", "+5", "+", "green"),
          row("生成菜品图", "7 月 7 日 20:04", "-8", "-", "orange", "orange"),
          row("AI 失败退还", "7 月 6 日 18:42", "+8", "+", "green")
        ]
      })
    ]
  },
  aiRecords: {
    navTitle: "AI 使用记录",
    back: true,
    titleBlock: {
      title: "AI 使用记录",
      desc: "查看解析、图片生成、备菜提醒和推荐结果。"
    },
    sections: [
      section("ai-record-list", "purchaseRows", "最近调用", {
        items: [
          row("不知道吃什么", "成功 · 免费次数 1/2", "成功", "AI", "green"),
          row("菜品文本解析", "成功 · 未扣积分", "成功", "文", "blue"),
          row("生成菜品图", "超时 · 已退还积分", "已退还", "图", "orange", "orange")
        ]
      })
    ]
  },
  checkin: {
    navTitle: "做菜打卡",
    back: true,
    titleBlock: {
      title: "做菜打卡",
      desc: "记录今天做过的菜，连续打卡会让推荐更懂你。"
    },
    sections: [
      section("calendar", "calendar", "", {
        days: Array.from({ length: 28 }, (_, index) => ({
          label: String(index + 1),
          done: [1, 2, 5, 6, 8, 12, 15].includes(index),
          today: index === 18
        }))
      }),
      section("today-checkin", "rows", "今日记录", {
        more: "新增",
        items: [
          row("清炒荷兰豆", "已获得今日首次打卡积分", "已记录", "菜", "green"),
          row("紫菜蛋花汤", "只记录，不重复奖励", "记录", "汤", "blue", "gray")
        ]
      })
    ],
    bottomActions: [{ text: "新增打卡", variant: "primary" }]
  },
  notifications: {
    navTitle: "通知中心",
    back: true,
    titleBlock: {
      title: "通知中心",
      desc: "站内通知会保留审核、积分和饭局状态变化。"
    },
    sections: [
      section("unread", "purchaseRows", "未读", {
        items: [
          row("菜品公开审核未通过", "原因：图片不够清晰，可修改后再提交。", "处理", "!", "orange", "orange"),
          row("饭局状态变化", "周末晚饭已关闭点菜，等待确认菜单。", "查看", "饭", "green")
        ]
      }),
      section("read", "purchaseRows", "已读", {
        items: [row("AI 失败退还积分", "生成菜品图超时，积分已返还。", "已读", "AI", "blue", "gray")]
      })
    ]
  },
  parties: {
    navTitle: "我的饭局",
    back: true,
    titleBlock: {
      title: "我的饭局",
      desc: "饭局入口放在这里，不抢首页菜品库的主路径。"
    },
    actions: [
      { text: "创建饭局", variant: "primary", to: routeOf.partyCreate },
      { text: "加入饭局", variant: "ghost", to: routeOf.partyJoin }
    ],
    sections: [
      section("active-party", "purchaseRows", "进行中", {
        items: [
          row("周末晚饭", "收集中 · 还有 2 小时过期", "点餐码", "饭", "green"),
          row("家庭小聚", "待确认菜单 · 5 人参与", "确认", "待", "orange", "orange")
        ]
      }),
      section("history-party", "purchaseRows", "历史饭局", {
        items: [row("端午午餐", "已生成采购清单 · 8 道菜", "查看", "✓", "blue", "gray")]
      })
    ]
  },
  partyCreate: {
    navTitle: "创建饭局",
    back: true,
    titleBlock: {
      title: "创建饭局",
      desc: "设置名称、有效期，再选择候选菜品。"
    },
    sections: [
      section("party-form", "fields", "", {
        fields: [
          field("饭局名称", "周末晚饭"),
          field("点餐码有效期", "今晚 22:00 前"),
          field("候选来源", "全部菜品 / 从菜谱选 / 常点菜品")
        ]
      }),
      section("candidate-dishes", "rows", "候选菜品", { more: "继续添加", moreTo: routeOf.dishSearch, items: myDishes })
    ],
    bottomActions: [
      { text: "保存稍后发", variant: "ghost" },
      { text: "生成点餐码", variant: "primary", to: routeOf.partyCode }
    ]
  },
  partyCode: {
    navTitle: "点餐码",
    back: true,
    titleBlock: {
      title: "周末晚饭",
      desc: "点餐码有效到今晚 22:00，可分享给家人朋友。"
    },
    sections: [
      section("code", "code", "", { code: "4826", desc: "输入点餐码加入饭局" }),
      section("code-stats", "stats", "", { items: [stat("5", "候选菜"), stat("3", "已参与"), stat("2h", "剩余")] }),
      section("code-notice", "notice", "当前状态", {
        text: "收集中。创建者可提前关闭点菜，关闭后参与者不能再修改选择。"
      })
    ],
    bottomActions: [
      { text: "分享点餐码", variant: "ghost" },
      { text: "提前关闭", variant: "primary", to: routeOf.partySummary }
    ]
  },
  partyJoin: {
    navTitle: "加入饭局",
    back: true,
    titleBlock: {
      title: "加入饭局",
      desc: "输入别人分享的点餐码，进入候选菜单。"
    },
    sections: [
      section("join-code", "fields", "", { fields: [field("点餐码", "4826")] }),
      section("join-notice", "notice", "", {
        text: "同一个微信用户不能重复加入同一个饭局；再次输入会回到原参与记录。"
      })
    ],
    bottomActions: [{ text: "进入点菜", variant: "primary", to: routeOf.partyOrder }]
  },
  partyOrder: {
    navTitle: "点菜",
    back: true,
    titleBlock: {
      title: "周末晚饭",
      desc: "点选想吃的菜，最终做几份由创建者确认。"
    },
    sections: [
      section("order-list", "rows", "", {
        items: [
          row("清蒸鲈鱼", "2 人想吃 · 清淡", "想吃", "鱼", "blue", "gray"),
          row("菌菇鸡汤", "3 人想吃 · 可提前准备", "已选", "汤", "orange"),
          row("蒜蓉西兰花", "1 人想吃 · 快手", "想吃", "菜", "green", "gray"),
          row("青椒牛柳", "4 人想吃 · 下饭", "想吃", "牛", "coral", "gray")
        ]
      })
    ],
    bottomActions: [{ text: "提交选择", variant: "primary" }],
    toast: "已保存你的选择"
  },
  partySummary: {
    navTitle: "确认菜单",
    back: true,
    titleBlock: {
      title: "确认最终菜单",
      desc: "根据想吃人数给出建议份数，可手动调整。"
    },
    sections: [
      section("summary-list", "purchaseRows", "", {
        items: [
          row("清蒸鲈鱼", "4 人想吃 · 建议 2 份", "2 份", "鱼", "blue"),
          row("菌菇鸡汤", "5 人想吃 · 建议 2 份", "2 份", "汤", "orange"),
          row("蒜蓉西兰花", "1 人想吃 · 可不做", "移除", "菜", "green", "gray")
        ]
      }),
      section("summary-notice", "notice", "", {
        text: "确认后会生成饭局快照和采购清单，历史记录不受菜品后续编辑影响。"
      })
    ],
    bottomActions: [
      { text: "继续调整", variant: "ghost" },
      { text: "生成采购清单", variant: "primary", to: routeOf.purchaseList }
    ]
  },
  purchaseList: {
    navTitle: "采购清单",
    back: true,
    titleBlock: {
      title: "周末晚饭采购",
      desc: "按最终制作份数汇总，可勾选、编辑、复制或分享。"
    },
    sections: [
      section("purchase-stats", "stats", "", { items: [stat("9", "总项"), stat("3", "已买"), stat("6", "待买")] }),
      section("purchase-items", "purchaseRows", "", {
        items: [
          row("鲈鱼", "2 条 · 来自清蒸鲈鱼", "已买", "✓", "green", "gray"),
          row("菌菇", "500g · 来自菌菇鸡汤", "待买", "菇", "orange"),
          row("青菜", "300g · 手动新增", "待买", "菜", "green"),
          row("姜葱", "适量 · 多菜合并", "已买", "✓", "blue", "gray")
        ]
      })
    ],
    bottomActions: [
      { text: "复制清单", variant: "ghost" },
      { text: "新增采购项", variant: "primary" }
    ]
  },
  aiPrepTips: {
    navTitle: "AI 备菜提醒",
    back: true,
    titleBlock: {
      title: "备菜提醒",
      desc: "给出处理优先级和并行建议，不强制变成时间表。"
    },
    sections: [
      section("prep-list", "purchaseRows", "", {
        items: [
          row("先处理菌菇鸡汤", "汤菜可提前炖煮，出餐前保温。", "优先", "高", "coral", "orange"),
          row("鲈鱼最后上锅蒸", "蒸好后口感最佳，临近开饭再做。", "中等", "中", "orange", "orange"),
          row("青菜提前洗切", "沥干水分，最后快炒。", "可提前", "备", "blue", "blue")
        ]
      }),
      section("prep-notice", "notice", "", {
        text: "提醒仅作参考，修改最终菜单或采购清单不会被这页限制。"
      })
    ],
    bottomActions: [
      { text: "重新生成", variant: "ghost" },
      { text: "保存提醒", variant: "primary" }
    ]
  },
  whatToEat: {
    navTitle: "不知道吃什么",
    back: true,
    titleBlock: {
      title: "不知道吃什么？",
      desc: "基于打卡、菜品标签、点菜记录和采购历史推荐。"
    },
    sections: [
      section("what-stats", "stats", "", { items: [stat("2", "今日免费"), stat("7", "打卡天数"), stat("18", "菜品样本")] }),
      section("what-fields", "fields", "", {
        fields: [
          field("今天偏好", "清淡 / 30 分钟内 / 有汤菜"),
          field("补充说明", "今晚 3 个人吃饭，想简单一点。", true)
        ]
      }),
      section("what-results", "rows", "推荐结果", {
        items: [
          row("菌菇鸡汤", "可提前准备 · 适合 3 人", "采用", "汤", "orange"),
          row("蒜蓉西兰花", "快手素菜 · 配汤刚好", "备选", "菜", "green", "gray")
        ]
      })
    ],
    bottomActions: [
      { text: "换一组", variant: "ghost" },
      { text: "加入今日菜单", variant: "primary" }
    ]
  },
  disabled: {
    navTitle: "账号状态",
    centerState: {
      title: "账号暂不可用",
      desc: "当前账号已被禁用，核心功能无法继续使用。请查看原因并按提示联系平台处理。",
      notice: "禁用原因：内容违规处理未完成。已有菜品、饭局和采购记录不会自动删除。"
    }
  }
};

const componentWxml = `<view class="screen">
  <view class="status-bar">
    <text>9:41</text>
    <view class="status-right"><text>5G</text><text>▮▮▮</text><text>84%</text></view>
  </view>
  <view class="nav-bar">
    <view class="nav-left" wx:if="{{screen.back}}" bindtap="goBack">‹</view>
    <view class="nav-left-placeholder" wx:else></view>
    <text class="nav-title">{{screen.navTitle}}</text>
    <view class="menu-capsule"><text></text><text></text><text></text></view>
  </view>

  <scroll-view scroll-y class="content {{screen.tab ? 'with-tab' : ''}} {{screen.bottomAI ? 'with-ai' : ''}} {{screen.bottomActions ? 'with-actions' : ''}}">
    <view wx:if="{{screen.hero}}" class="hero">
      <view class="hero-dot dot-a"></view>
      <view class="hero-dot dot-b"></view>
      <view class="hero-dot dot-c"></view>
      <view class="hero-copy">
        <text class="hero-title">{{screen.hero.title}}</text>
        <text class="hero-desc">{{screen.hero.desc}}</text>
      </view>
      <cooking-cat size="large"></cooking-cat>
    </view>

    <view wx:if="{{screen.titleBlock}}" class="title-block">
      <view class="title-copy">
        <text class="page-title">{{screen.titleBlock.title}}</text>
        <text class="page-desc">{{screen.titleBlock.desc}}</text>
      </view>
      <view class="mini-cat-card"><cooking-cat size="small"></cooking-cat></view>
    </view>

    <view wx:if="{{screen.detailHero}}" class="detail-hero">
      <cooking-cat size="small"></cooking-cat>
      <view class="detail-copy">
        <text class="detail-title">{{screen.detailHero.title}}</text>
        <text class="detail-desc">{{screen.detailHero.desc}}</text>
      </view>
    </view>

    <view wx:if="{{screen.profileCard}}" class="profile-card card">
      <view class="mini-cat-card"><cooking-cat size="small"></cooking-cat></view>
      <view class="profile-info">
        <text class="profile-name">微信用户</text>
        <text class="profile-meta">积分 128 · 已打卡 7 天</text>
      </view>
      <text class="badge">可推荐</text>
    </view>

    <view wx:if="{{screen.search}}" class="search-box">
      <view class="search-icon"></view>
      <text>{{screen.search}}</text>
    </view>

    <view wx:if="{{screen.actions}}" class="action-grid">
      <button wx:for="{{screen.actions}}" wx:for-item="action" wx:key="text" class="ui-btn {{action.variant == 'primary' ? 'primary' : 'ghost'}}" data-to="{{action.to}}" bindtap="handleNav">{{action.text}}</button>
    </view>

    <view wx:if="{{screen.chips}}" class="chip-row">
      <text wx:for="{{screen.chips}}" wx:for-item="chip" wx:key="text" class="chip chip-{{chip.tone}}">{{chip.text}}</text>
    </view>

    <block wx:for="{{screen.sections}}" wx:for-item="section" wx:key="id">
      <view wx:if="{{section.type == 'rows'}}" class="section">
        <view wx:if="{{section.title || section.more}}" class="section-head">
          <text class="section-title">{{section.title}}</text>
          <text wx:if="{{section.more}}" class="more" data-to="{{section.moreTo}}" bindtap="handleNav">{{section.more}}</text>
        </view>
        <view class="card list-card">
          <view wx:for="{{section.items}}" wx:for-item="row" wx:key="title" class="row" data-to="{{row.to}}" bindtap="handleNav">
            <view class="thumb tone-{{row.tone}}"><text>{{row.mark}}</text></view>
            <view class="row-main"><text class="row-title">{{row.title}}</text><text class="row-desc">{{row.desc}}</text></view>
            <text class="badge badge-{{row.badgeTone}}">{{row.badge}}</text>
          </view>
        </view>
      </view>

      <view wx:if="{{section.type == 'recommend'}}" class="section">
        <view wx:if="{{section.title || section.more}}" class="section-head">
          <text class="section-title">{{section.title}}</text>
          <text wx:if="{{section.more}}" class="more" data-to="{{section.moreTo}}" bindtap="handleNav">{{section.more}}</text>
        </view>
        <view class="recommend-stack">
          <view wx:for="{{section.items}}" wx:for-item="row" wx:key="title" class="card rec-card">
            <view class="thumb tone-{{row.tone}}"><text>{{row.mark}}</text></view>
            <view class="rec-main">
              <text class="row-title">{{row.title}}</text>
              <text class="row-desc">{{row.desc}}</text>
              <button class="mini-btn" catchtap="noop">{{row.action}}</button>
            </view>
          </view>
        </view>
      </view>

      <view wx:if="{{section.type == 'fields'}}" class="section">
        <view wx:if="{{section.title || section.more}}" class="section-head">
          <text class="section-title small">{{section.title}}</text>
          <text wx:if="{{section.more}}" class="more">{{section.more}}</text>
        </view>
        <view class="card form-card">
          <view wx:for="{{section.fields}}" wx:for-item="field" wx:key="label" class="field">
            <text class="field-label">{{field.label}}</text>
            <view class="input {{field.large ? 'textarea' : ''}}"><text>{{field.value}}</text></view>
          </view>
        </view>
      </view>

      <view wx:if="{{section.type == 'stats'}}" class="stats-grid">
        <view wx:for="{{section.items}}" wx:for-item="stat" wx:key="label" class="card stat-card">
          <text class="stat-value">{{stat.value}}</text>
          <text class="stat-label">{{stat.label}}</text>
        </view>
      </view>

      <view wx:if="{{section.type == 'tiles'}}" class="section">
        <view wx:if="{{section.title}}" class="section-head"><text class="section-title">{{section.title}}</text></view>
        <view class="tile-grid">
          <view wx:for="{{section.items}}" wx:for-item="tile" wx:key="title" class="card tile-card">
            <text class="tile-title">{{tile.title}}</text>
            <text class="tile-desc">{{tile.desc}}</text>
            <view wx:if="{{tile.chips.length}}" class="chip-row tight">
              <text wx:for="{{tile.chips}}" wx:for-item="chip" wx:key="*this" class="chip chip-green">{{chip}}</text>
            </view>
          </view>
        </view>
      </view>

      <view wx:if="{{section.type == 'purchaseRows'}}" class="section">
        <view wx:if="{{section.title}}" class="section-head"><text class="section-title">{{section.title}}</text></view>
        <view class="card list-card">
          <view wx:for="{{section.items}}" wx:for-item="row" wx:key="title" class="purchase-row">
            <view class="round-mark tone-{{row.tone}}"><text>{{row.mark}}</text></view>
            <view class="row-main"><text class="row-title compact">{{row.title}}</text><text class="row-desc">{{row.desc}}</text></view>
            <text class="badge badge-{{row.badgeTone}}">{{row.badge}}</text>
          </view>
        </view>
      </view>

      <view wx:if="{{section.type == 'calendar'}}" class="card calendar">
        <view wx:for="{{section.days}}" wx:for-item="day" wx:key="label" class="day {{day.done ? 'done' : ''}} {{day.today ? 'today' : ''}}">{{day.label}}</view>
      </view>

      <view wx:if="{{section.type == 'code'}}" class="card code-card">
        <text class="code-text">{{section.code}}</text>
        <text class="code-desc">{{section.desc}}</text>
      </view>

      <view wx:if="{{section.type == 'notice'}}" class="section">
        <view wx:if="{{section.title}}" class="section-head"><text class="section-title small">{{section.title}}</text></view>
        <view class="notice">{{section.text}}</view>
      </view>

      <view wx:if="{{section.type == 'imageMock'}}" class="card image-mock">
        <view class="plate"></view>
        <view class="food food-a"></view>
        <view class="food food-b"></view>
        <view class="food food-c"></view>
      </view>
    </block>

    <view wx:if="{{screen.centerState}}" class="center-state">
      <view class="center-cat"><cooking-cat size="large"></cooking-cat></view>
      <text class="center-title">{{screen.centerState.title}}</text>
      <text class="center-desc">{{screen.centerState.desc}}</text>
      <view class="notice">{{screen.centerState.notice}}</view>
    </view>
  </scroll-view>

  <view wx:if="{{screen.bottomAI}}" class="bottom-ai" data-to="/pages/ai/what-to-eat" bindtap="handleNav">
    <view><text class="bottom-ai-title">不知道吃什么？</text><text class="bottom-ai-desc">按口味和记录帮你推荐</text></view>
    <button class="mini-btn">试试推荐</button>
  </view>

  <view wx:if="{{screen.bottomActions}}" class="bottom-actions {{screen.bottomActions.length == 1 ? 'single' : ''}}">
    <button wx:for="{{screen.bottomActions}}" wx:for-item="action" wx:key="text" class="ui-btn {{action.variant == 'primary' ? 'primary' : 'ghost'}}" data-to="{{action.to}}" bindtap="handleNav">{{action.text}}</button>
  </view>

  <view wx:if="{{screen.toast}}" class="toast">{{screen.toast}}</view>

  <view wx:if="{{screen.tab}}" class="tabbar">
    <view class="tab-item {{screen.tab == 'home' ? 'active' : ''}}" data-to="/pages/home/home" bindtap="handleRedirect"><view class="tab-icon">⌂</view><text>首页</text></view>
    <view class="tab-item {{screen.tab == 'recipe' ? 'active' : ''}}" data-to="/pages/recipe/list" bindtap="handleRedirect"><view class="tab-icon">食</view><text>菜谱</text></view>
    <view class="tab-item {{screen.tab == 'profile' ? 'active' : ''}}" data-to="/pages/profile/profile" bindtap="handleRedirect"><view class="tab-icon">人</view><text>我的</text></view>
  </view>
</view>`;

const componentJs = `Component({
  properties: {
    screen: {
      type: Object,
      value: {}
    }
  },
  methods: {
    goBack() {
      const pages = getCurrentPages();
      if (pages.length > 1) {
        wx.navigateBack();
        return;
      }
      wx.redirectTo({ url: "/pages/home/home" });
    },
    handleNav(event) {
      const to = event.currentTarget.dataset.to;
      if (!to) return;
      wx.navigateTo({ url: to });
    },
    handleRedirect(event) {
      const to = event.currentTarget.dataset.to;
      if (!to) return;
      wx.redirectTo({ url: to });
    },
    noop() {}
  }
});`;

const componentJson = `{
  "component": true,
  "usingComponents": {
    "cooking-cat": "/components/cooking-cat/index"
  }
}
`;

const componentWxss = `:host {
  display: block;
}

button {
  margin: 0;
  padding: 0;
  line-height: 1;
  background: transparent;
  border-radius: 0;
  font: inherit;
}

button::after {
  border: 0;
}

.screen {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  color: #21362d;
  background: radial-gradient(circle at 10% 14%, rgba(255, 216, 106, 0.18), transparent 24%),
    radial-gradient(circle at 92% 6%, rgba(130, 203, 217, 0.16), transparent 24%),
    linear-gradient(180deg, #fffaf2 0%, #fffdf9 48%, #f8fbf6 100%);
}

.status-bar {
  height: 56rpx;
  padding: 16rpx 32rpx 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 24rpx;
  font-weight: 700;
  color: rgba(33, 54, 45, 0.9);
}

.status-right {
  display: flex;
  align-items: center;
  gap: 10rpx;
  font-size: 22rpx;
  color: rgba(33, 54, 45, 0.72);
}

.nav-bar {
  height: 92rpx;
  padding: 0 28rpx;
  display: grid;
  grid-template-columns: 72rpx 1fr 96rpx;
  align-items: center;
  position: relative;
  z-index: 5;
}

.nav-left,
.nav-left-placeholder {
  width: 64rpx;
  height: 64rpx;
}

.nav-left {
  border-radius: 32rpx;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.76);
  font-size: 48rpx;
  line-height: 1;
  color: #21362d;
}

.nav-title {
  text-align: center;
  font-size: 32rpx;
  font-weight: 900;
  color: #21362d;
}

.menu-capsule {
  justify-self: end;
  width: 92rpx;
  height: 48rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.76);
  border: 1rpx solid rgba(38, 56, 47, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}

.menu-capsule text {
  width: 8rpx;
  height: 8rpx;
  border-radius: 50%;
  background: #8a958f;
}

.content {
  height: calc(100vh - 148rpx);
  padding: 0 28rpx 48rpx;
  box-sizing: border-box;
}

.content.with-tab {
  padding-bottom: 160rpx;
}

.content.with-ai {
  padding-bottom: 286rpx;
}

.content.with-actions {
  padding-bottom: 150rpx;
}

.hero {
  position: relative;
  min-height: 244rpx;
  margin: -148rpx -28rpx 24rpx;
  padding: 166rpx 32rpx 20rpx;
  overflow: hidden;
  background: linear-gradient(120deg, rgba(255, 245, 214, 0.96), rgba(227, 248, 235, 0.96));
  border-bottom: 1rpx solid rgba(21, 121, 70, 0.09);
}

.hero-copy {
  width: 62%;
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.hero-title {
  font-size: 54rpx;
  line-height: 1.12;
  font-weight: 900;
}

.hero-desc {
  font-size: 25rpx;
  line-height: 1.45;
  color: #5f6f66;
}

.hero cooking-cat {
  position: absolute;
  right: 12rpx;
  bottom: 0;
}

.hero-dot {
  position: absolute;
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  background: #ffd86a;
  opacity: 0.75;
}

.dot-a { left: 46rpx; top: 88rpx; }
.dot-b { right: 74rpx; top: 110rpx; background: #ff8f7a; }
.dot-c { right: 150rpx; bottom: 42rpx; background: #82cbd9; }

.title-block {
  display: grid;
  grid-template-columns: 1fr 176rpx;
  gap: 22rpx;
  align-items: start;
  margin: 8rpx 0 24rpx;
}

.title-copy,
.profile-info,
.row-main,
.rec-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.page-title {
  font-size: 48rpx;
  line-height: 1.15;
  font-weight: 900;
}

.page-desc {
  margin-top: 12rpx;
  font-size: 25rpx;
  line-height: 1.45;
  color: #798881;
}

.mini-cat-card {
  width: 176rpx;
  height: 146rpx;
  border-radius: 16rpx;
  overflow: hidden;
  background: linear-gradient(145deg, #fff3dd, #eaf8ef);
  display: grid;
  place-items: center;
}

.detail-hero {
  height: 308rpx;
  margin-bottom: 22rpx;
  border-radius: 16rpx;
  position: relative;
  overflow: hidden;
  background: linear-gradient(145deg, #fff0dc, #eaf8ef);
  box-shadow: 0 12rpx 36rpx rgba(39, 76, 55, 0.08);
}

.detail-hero cooking-cat {
  position: absolute;
  right: 24rpx;
  top: 26rpx;
}

.detail-copy {
  position: absolute;
  left: 32rpx;
  bottom: 30rpx;
  display: flex;
  flex-direction: column;
}

.detail-title {
  font-size: 52rpx;
  font-weight: 900;
}

.detail-desc {
  margin-top: 12rpx;
  font-size: 25rpx;
  color: #68786f;
}

.profile-card {
  padding: 24rpx;
  display: grid;
  grid-template-columns: 132rpx 1fr auto;
  gap: 20rpx;
  align-items: center;
  margin-top: 8rpx;
}

.profile-card .mini-cat-card {
  width: 132rpx;
  height: 116rpx;
}

.profile-name {
  font-size: 40rpx;
  line-height: 1.2;
  font-weight: 900;
}

.profile-meta {
  margin-top: 12rpx;
  color: #798881;
  font-size: 25rpx;
}

.search-box {
  height: 84rpx;
  padding: 0 26rpx;
  border-radius: 42rpx;
  display: flex;
  align-items: center;
  gap: 18rpx;
  background: rgba(255, 255, 255, 0.92);
  border: 1rpx solid rgba(53, 84, 68, 0.06);
  box-shadow: 0 10rpx 32rpx rgba(38, 76, 55, 0.06);
  color: #798881;
  font-size: 27rpx;
  white-space: nowrap;
}

.search-icon {
  width: 28rpx;
  height: 28rpx;
  border: 4rpx solid #93a09a;
  border-radius: 50%;
  position: relative;
}

.search-icon::after {
  content: "";
  position: absolute;
  width: 14rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #93a09a;
  right: -12rpx;
  bottom: -6rpx;
  transform: rotate(45deg);
}

.action-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18rpx;
  margin: 22rpx 0;
}

.ui-btn {
  height: 84rpx;
  border-radius: 42rpx;
  display: grid;
  place-items: center;
  font-size: 30rpx;
  font-weight: 900;
}

.ui-btn.primary {
  background: linear-gradient(180deg, #19cc75, #12b963);
  color: #fff;
  box-shadow: 0 18rpx 42rpx rgba(24, 191, 107, 0.22);
}

.ui-btn.ghost {
  background: #fff;
  color: #079757;
  border: 1rpx solid rgba(24, 191, 107, 0.16);
  box-shadow: 0 12rpx 28rpx rgba(39, 76, 55, 0.07);
}

.chip-row {
  display: flex;
  gap: 14rpx;
  align-items: center;
  flex-wrap: nowrap;
  margin: 18rpx 0 8rpx;
}

.chip-row.tight {
  margin-top: 18rpx;
  flex-wrap: wrap;
}

.chip {
  height: 56rpx;
  padding: 0 22rpx;
  border-radius: 28rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 23rpx;
  font-weight: 900;
  white-space: nowrap;
}

.chip-green {
  background: #eaf8ef;
  color: #079757;
  border: 1rpx solid rgba(24, 191, 107, 0.08);
}

.chip-orange {
  background: #fff3df;
  color: #d36b22;
}

.chip-blue {
  background: #e9f7fb;
  color: #2a8aa0;
}

.section {
  margin-top: 26rpx;
}

.section-head {
  height: 58rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;
}

.section-title {
  font-size: 38rpx;
  line-height: 1;
  font-weight: 900;
}

.section-title.small {
  font-size: 34rpx;
}

.more {
  font-size: 25rpx;
  color: #079757;
  font-weight: 900;
}

.card {
  background: rgba(255, 255, 255, 0.94);
  border-radius: 16rpx;
  border: 1rpx solid rgba(39, 76, 55, 0.05);
  box-shadow: 0 14rpx 38rpx rgba(39, 76, 55, 0.08);
}

.list-card {
  padding: 8rpx 0;
}

.row {
  min-height: 136rpx;
  padding: 18rpx 22rpx;
  display: grid;
  grid-template-columns: 92rpx 1fr auto;
  gap: 22rpx;
  align-items: center;
  border-bottom: 1rpx solid #edf0e8;
}

.row:last-child,
.purchase-row:last-child {
  border-bottom: 0;
}

.thumb,
.round-mark {
  width: 92rpx;
  height: 92rpx;
  border-radius: 18rpx;
  display: grid;
  place-items: center;
  border: 1rpx solid rgba(24, 191, 107, 0.08);
  font-size: 32rpx;
  font-weight: 900;
}

.round-mark {
  width: 44rpx;
  height: 44rpx;
  border-radius: 22rpx;
  font-size: 22rpx;
}

.tone-green { background: linear-gradient(145deg, #f3fbf5, #eaf8ef); color: #079757; }
.tone-blue { background: linear-gradient(145deg, #eef9fb, #e2f4f8); color: #248aa2; }
.tone-orange { background: linear-gradient(145deg, #fff6e8, #ffedda); color: #d36b22; }
.tone-coral { background: linear-gradient(145deg, #fff0ec, #ffe5dd); color: #d65e50; }
.tone-yellow { background: linear-gradient(145deg, #fff8d9, #fff0b8); color: #b87900; }

.row-title {
  font-size: 32rpx;
  line-height: 1.25;
  font-weight: 900;
}

.row-title.compact {
  font-size: 29rpx;
}

.row-desc {
  margin-top: 10rpx;
  font-size: 24rpx;
  line-height: 1.35;
  color: #798881;
}

.badge {
  min-width: 88rpx;
  height: 52rpx;
  padding: 0 18rpx;
  border-radius: 26rpx;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #eaf8ef;
  color: #079757;
  font-size: 23rpx;
  font-weight: 900;
  white-space: nowrap;
}

.badge-gray {
  background: #f2f4ef;
  color: #7d8a83;
}

.badge-orange {
  background: #fff1df;
  color: #d56d22;
}

.badge-blue {
  background: #eaf7fb;
  color: #248aa2;
}

.recommend-stack {
  display: grid;
  gap: 18rpx;
  padding-bottom: 36rpx;
}

.rec-card {
  padding: 22rpx;
  display: grid;
  grid-template-columns: 98rpx 1fr;
  gap: 22rpx;
}

.mini-btn {
  min-width: 120rpx;
  height: 56rpx;
  padding: 0 20rpx;
  border-radius: 28rpx;
  display: inline-grid;
  place-items: center;
  background: #18bf6b;
  color: #fff;
  font-size: 23rpx;
  font-weight: 900;
  align-self: flex-start;
}

.form-card {
  padding: 24rpx;
  display: grid;
  gap: 20rpx;
}

.field {
  display: grid;
  gap: 10rpx;
}

.field-label {
  font-size: 23rpx;
  color: #798881;
  font-weight: 900;
}

.input {
  min-height: 76rpx;
  padding: 0 22rpx;
  border-radius: 16rpx;
  background: #f8fbf6;
  border: 1rpx solid #e7ede6;
  display: flex;
  align-items: center;
  color: #21362d;
  font-size: 27rpx;
  font-weight: 700;
  line-height: 1.5;
}

.textarea {
  min-height: 150rpx;
  align-items: flex-start;
  padding-top: 20rpx;
  color: #5f6f66;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18rpx;
  margin: 22rpx 0 8rpx;
}

.stat-card {
  min-height: 132rpx;
  padding: 24rpx 12rpx;
  text-align: center;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.stat-value {
  font-size: 42rpx;
  line-height: 1;
  font-weight: 900;
}

.stat-label {
  margin-top: 14rpx;
  color: #798881;
  font-size: 23rpx;
  font-weight: 800;
}

.tile-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18rpx;
}

.tile-card {
  min-height: 214rpx;
  padding: 24rpx;
}

.tile-title {
  font-size: 32rpx;
  line-height: 1.25;
  font-weight: 900;
}

.tile-desc {
  margin-top: 14rpx;
  color: #798881;
  font-size: 23rpx;
  line-height: 1.45;
}

.purchase-row {
  min-height: 106rpx;
  padding: 22rpx;
  display: grid;
  grid-template-columns: 52rpx 1fr auto;
  gap: 18rpx;
  align-items: center;
  border-bottom: 1rpx solid #edf0e8;
}

.calendar {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 14rpx;
  padding: 22rpx;
  margin-top: 12rpx;
}

.day {
  height: 62rpx;
  border-radius: 16rpx;
  background: #f8fbf6;
  display: grid;
  place-items: center;
  color: #76867e;
  font-size: 23rpx;
  font-weight: 900;
}

.day.done {
  background: #eaf8ef;
  color: #079757;
}

.day.today {
  background: #18bf6b;
  color: #fff;
}

.code-card {
  height: 220rpx;
  display: grid;
  place-items: center;
  text-align: center;
  border: 1rpx dashed rgba(24, 191, 107, 0.35);
  background: linear-gradient(135deg, #ffffff, #effaf4);
}

.code-text {
  font-size: 66rpx;
  letter-spacing: 8rpx;
  color: #079757;
  font-weight: 900;
}

.code-desc {
  margin-top: -42rpx;
  font-size: 24rpx;
  color: #798881;
}

.notice {
  padding: 24rpx;
  border-radius: 16rpx;
  background: #fff8e8;
  border: 1rpx solid rgba(255, 173, 104, 0.25);
  color: #815625;
  font-size: 24rpx;
  line-height: 1.55;
  font-weight: 800;
}

.image-mock {
  height: 386rpx;
  display: grid;
  place-items: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(145deg, #fff2df, #eaf8ef);
}

.plate {
  position: absolute;
  left: 170rpx;
  right: 170rpx;
  bottom: 80rpx;
  height: 92rpx;
  border-radius: 50%;
  background: #fff;
  border: 6rpx solid #42534a;
}

.food {
  position: absolute;
  border-radius: 60rpx;
}

.food-a {
  width: 150rpx;
  height: 88rpx;
  left: 260rpx;
  top: 142rpx;
  background: #ff8f7a;
}

.food-b {
  width: 90rpx;
  height: 74rpx;
  left: 300rpx;
  top: 102rpx;
  background: #ffd86a;
}

.food-c {
  width: 220rpx;
  height: 36rpx;
  left: 228rpx;
  top: 222rpx;
  background: #18bf6b;
}

.bottom-ai {
  position: fixed;
  left: 28rpx;
  right: 28rpx;
  bottom: 126rpx;
  height: 138rpx;
  padding: 22rpx 24rpx;
  box-sizing: border-box;
  border-radius: 16rpx;
  background: linear-gradient(110deg, rgba(255, 255, 255, 0.96), rgba(236, 250, 242, 0.96));
  border: 1rpx solid rgba(24, 191, 107, 0.12);
  box-shadow: 0 18rpx 44rpx rgba(33, 54, 45, 0.12);
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  z-index: 8;
}

.bottom-ai-title {
  display: block;
  font-size: 32rpx;
  font-weight: 900;
}

.bottom-ai-desc {
  display: block;
  margin-top: 10rpx;
  color: #798881;
  font-size: 23rpx;
}

.bottom-actions {
  position: fixed;
  left: 28rpx;
  right: 28rpx;
  bottom: 30rpx;
  display: grid;
  grid-template-columns: 1fr 1.25fr;
  gap: 18rpx;
  z-index: 8;
}

.bottom-actions.single {
  grid-template-columns: 1fr;
}

.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 112rpx;
  padding: 12rpx 52rpx 18rpx;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18rpx;
  background: rgba(255, 255, 255, 0.96);
  border-top: 1rpx solid rgba(39, 76, 55, 0.08);
  z-index: 7;
}

.tab-item {
  display: grid;
  justify-items: center;
  align-content: center;
  gap: 6rpx;
  color: #87928d;
  font-size: 23rpx;
  font-weight: 800;
}

.tab-item.active {
  color: #079757;
}

.tab-icon {
  width: 46rpx;
  height: 46rpx;
  border-radius: 23rpx;
  border: 1rpx solid #dfe5df;
  display: grid;
  place-items: center;
  font-size: 24rpx;
  font-weight: 900;
}

.tab-item.active .tab-icon {
  background: #eaf8ef;
  border-color: rgba(24, 191, 107, 0.28);
}

.toast {
  position: fixed;
  left: 150rpx;
  right: 150rpx;
  bottom: 160rpx;
  height: 68rpx;
  border-radius: 34rpx;
  display: grid;
  place-items: center;
  background: rgba(33, 54, 45, 0.84);
  color: #fff;
  font-size: 24rpx;
  font-weight: 900;
  z-index: 9;
}

.center-state {
  min-height: 680rpx;
  display: grid;
  align-content: center;
  justify-items: center;
  text-align: center;
  padding: 0 32rpx;
}

.center-cat {
  width: 220rpx;
  height: 180rpx;
  overflow: hidden;
}

.center-title {
  margin-top: 22rpx;
  font-size: 44rpx;
  font-weight: 900;
}

.center-desc {
  margin-top: 18rpx;
  color: #798881;
  font-size: 27rpx;
  line-height: 1.6;
}
`;

const catWxml = `<view class="cat {{size}}">
  <view class="shadow"></view>
  <view class="ear ear-left"></view>
  <view class="ear ear-right"></view>
  <view class="head">
    <view class="eye eye-left"></view>
    <view class="eye eye-right"></view>
    <view class="nose"></view>
    <view class="mouth"></view>
    <view class="whisker whisker-left-a"></view>
    <view class="whisker whisker-left-b"></view>
    <view class="whisker whisker-right-a"></view>
    <view class="whisker whisker-right-b"></view>
  </view>
  <view class="apron">
    <view class="apron-line line-a"></view>
    <view class="apron-line line-b"></view>
  </view>
  <view class="arm arm-left"></view>
  <view class="arm arm-right"></view>
  <view class="pan"></view>
  <view class="pan-lip"></view>
  <view class="steam steam-a"></view>
  <view class="steam steam-b"></view>
</view>`;

const catWxss = `:host {
  display: inline-block;
}

.cat {
  position: relative;
  width: 250rpx;
  height: 210rpx;
}

.cat.small {
  transform: scale(0.62);
  transform-origin: center center;
}

.shadow {
  position: absolute;
  left: 44rpx;
  right: 32rpx;
  bottom: 6rpx;
  height: 28rpx;
  border-radius: 50%;
  background: rgba(63, 105, 77, 0.12);
}

.ear {
  position: absolute;
  top: 24rpx;
  width: 58rpx;
  height: 58rpx;
  background: #ffd7a6;
  border: 5rpx solid #4b3a2d;
  transform: rotate(45deg);
  z-index: 1;
}

.ear-left {
  left: 54rpx;
}

.ear-right {
  right: 34rpx;
}

.head {
  position: absolute;
  left: 58rpx;
  top: 42rpx;
  width: 146rpx;
  height: 124rpx;
  border-radius: 58% 58% 52% 52%;
  background: #ffddb0;
  border: 5rpx solid #4b3a2d;
  z-index: 2;
}

.eye {
  position: absolute;
  top: 48rpx;
  width: 10rpx;
  height: 10rpx;
  border-radius: 50%;
  background: #4b3a2d;
}

.eye-left {
  left: 44rpx;
}

.eye-right {
  right: 44rpx;
}

.nose {
  position: absolute;
  left: 68rpx;
  top: 64rpx;
  width: 18rpx;
  height: 14rpx;
  border-radius: 0 0 12rpx 12rpx;
  background: #ee8b78;
}

.mouth {
  position: absolute;
  left: 48rpx;
  top: 84rpx;
  width: 50rpx;
  height: 16rpx;
  border-bottom: 5rpx solid #4b3a2d;
  border-radius: 0 0 50rpx 50rpx;
}

.whisker {
  position: absolute;
  width: 42rpx;
  height: 4rpx;
  border-radius: 4rpx;
  background: #4b3a2d;
}

.whisker-left-a { left: -34rpx; top: 64rpx; }
.whisker-left-b { left: -30rpx; top: 82rpx; }
.whisker-right-a { right: -34rpx; top: 64rpx; }
.whisker-right-b { right: -30rpx; top: 82rpx; }

.apron {
  position: absolute;
  left: 72rpx;
  top: 142rpx;
  width: 128rpx;
  height: 58rpx;
  background: linear-gradient(180deg, #dff6ea, #9fe4bd);
  border: 5rpx solid #4b3a2d;
  border-radius: 42rpx 42rpx 10rpx 10rpx;
  z-index: 1;
}

.apron-line {
  position: absolute;
  left: 36rpx;
  right: 26rpx;
  height: 5rpx;
  border-radius: 5rpx;
  background: #fff;
}

.line-a { top: 18rpx; }
.line-b { top: 34rpx; }

.arm {
  position: absolute;
  width: 60rpx;
  height: 40rpx;
  border: 8rpx solid #4b3a2d;
  border-top: 0;
  border-radius: 0 0 40rpx 40rpx;
  z-index: 3;
}

.arm-left {
  left: 26rpx;
  top: 134rpx;
  transform: rotate(18deg);
}

.arm-right {
  right: 12rpx;
  top: 122rpx;
  transform: rotate(-34deg);
}

.pan {
  position: absolute;
  left: 32rpx;
  bottom: 20rpx;
  width: 138rpx;
  height: 42rpx;
  border: 8rpx solid #42534a;
  border-top: 0;
  border-radius: 0 0 74rpx 74rpx;
  z-index: 4;
}

.pan-lip {
  position: absolute;
  left: 38rpx;
  bottom: 62rpx;
  width: 148rpx;
  height: 8rpx;
  border-radius: 8rpx;
  background: #42534a;
  z-index: 5;
}

.steam {
  position: absolute;
  width: 10rpx;
  height: 28rpx;
  border-radius: 20rpx;
  background: #fff;
  opacity: 0.9;
  z-index: 5;
}

.steam-a {
  left: 164rpx;
  top: 112rpx;
}

.steam-b {
  left: 188rpx;
  top: 108rpx;
  height: 22rpx;
}
`;

function write(file, content) {
  const target = path.join(root, file);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(target, content, "utf8");
}

function writeJson(file, value) {
  write(file, `${JSON.stringify(value, null, 2)}\n`);
}

writeJson("app.json", {
  pages: pages.map(([, route]) => route),
  window: {
    navigationStyle: "custom",
    backgroundColor: "#fffaf2",
    backgroundTextStyle: "dark"
  },
  usingComponents: {
    "ui-screen": "/components/ui-screen/index"
  },
  style: "v2",
  lazyCodeLoading: "requiredComponents",
  sitemapLocation: "sitemap.json"
});

write("app.js", "App({});\n");
write("app.wxss", `page {
  margin: 0;
  min-height: 100vh;
  background: #fffaf2;
  font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Microsoft YaHei", sans-serif;
}

view,
text,
button,
scroll-view {
  box-sizing: border-box;
}
`);
writeJson("sitemap.json", {
  rules: [
    {
      action: "allow",
      page: "*"
    }
  ]
});
writeJson("project.config.json", {
  miniprogramRoot: "./",
  compileType: "miniprogram",
  appid: "touristappid",
  projectname: "OrderFoodMiniApp-static-ui",
  setting: {
    urlCheck: false,
    es6: true,
    postcss: true,
    minified: false
  }
});

write("common/screens.js", `const screens = ${JSON.stringify(screens, null, 2)};\n\nmodule.exports = screens;\n`);

write("components/ui-screen/index.wxml", componentWxml);
write("components/ui-screen/index.wxss", componentWxss);
write("components/ui-screen/index.js", componentJs);
write("components/ui-screen/index.json", componentJson);

write("components/cooking-cat/index.wxml", catWxml);
write("components/cooking-cat/index.wxss", catWxss);
write("components/cooking-cat/index.js", "Component({ properties: { size: { type: String, value: 'large' } } });\n");
write("components/cooking-cat/index.json", "{\n  \"component\": true\n}\n");

for (const [key, route] of pages) {
  write(`${route}.wxml`, `<ui-screen screen="{{screen}}" />\n`);
  write(`${route}.js`, `const screens = require("../../common/screens.js");\n\nPage({\n  data: {\n    screen: screens.${key}\n  }\n});\n`);
}

console.log(`Generated ${pages.length} static mini program pages.`);
