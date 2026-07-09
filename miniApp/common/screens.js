const screens = {
  "home": {
    "navTitle": "首页",
    "tab": "home",
    "hero": {
      "title": "来干饭",
      "desc": "把爱吃的菜，整理成自己的菜品库"
    },
    "search": "搜索菜名、分类、食材",
    "actions": [
      {
        "text": "添加菜品",
        "variant": "primary",
        "to": "/pages/dish/add"
      },
      {
        "text": "AI 识别",
        "variant": "ghost",
        "to": "/pages/dish/ai-parse"
      }
    ],
    "chips": [
      {
        "text": "手动录入",
        "tone": "green"
      },
      {
        "text": "生成菜品图",
        "tone": "orange"
      },
      {
        "text": "分类筛选",
        "tone": "blue"
      }
    ],
    "sections": [
      {
        "id": "home-dishes",
        "type": "rows",
        "title": "我的菜品",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "清淡 · 适合晚饭",
            "badge": "已完善",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "汤菜 · 可提前准备",
            "badge": "菜谱内",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "blue",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "快手 · 素菜",
            "badge": "常点",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "orange",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      },
      {
        "id": "home-rec",
        "type": "recommend",
        "title": "菜品推荐",
        "more": "查看更多",
        "moreTo": "/pages/dish/recommendations",
        "items": [
          {
            "title": "虾仁蒸蛋",
            "desc": "公开精选 · 适合孩子",
            "mark": "虾",
            "tone": "coral",
            "action": "加入菜品库",
            "image": "/assets/dishes/shrimp-egg.png"
          },
          {
            "title": "南瓜小米粥",
            "desc": "清淡暖胃 · 早餐晚餐都合适",
            "mark": "粥",
            "tone": "yellow",
            "action": "加入菜品库",
            "image": "/assets/dishes/pumpkin-porridge.png"
          },
          {
            "title": "青椒牛柳",
            "desc": "下饭快手 · 家常热菜",
            "mark": "牛",
            "tone": "coral",
            "action": "加入菜品库",
            "image": "/assets/dishes/beef-pepper.png"
          },
          {
            "title": "冬瓜丸子汤",
            "desc": "汤菜 · 适合多人",
            "mark": "汤",
            "tone": "blue",
            "action": "加入菜品库",
            "image": "/assets/dishes/winter-soup.png"
          },
          {
            "title": "凉拌黄瓜",
            "desc": "爽口素菜 · 可提前准备",
            "mark": "瓜",
            "tone": "green",
            "action": "加入菜品库",
            "image": "/assets/dishes/cucumber.png"
          }
        ]
      }
    ],
    "bottomAI": true
  },
  "dishAdd": {
    "navTitle": "添加菜品",
    "back": true,
    "titleBlock": {
      "title": "添加菜品",
      "desc": "先把关键信息记下来，缺的部分之后再补。"
    },
    "sections": [
      {
        "id": "dish-add-form",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "菜名",
            "value": "清炒荷兰豆",
            "large": false
          },
          {
            "label": "分类",
            "value": "素菜",
            "large": false
          },
          {
            "label": "基础份量",
            "value": "2 人份",
            "large": false
          },
          {
            "label": "标签",
            "value": "快手 / 清淡 / 适合晚饭",
            "large": false
          },
          {
            "label": "配料",
            "value": "荷兰豆 300g，蒜 2 瓣，盐 少许",
            "large": true
          },
          {
            "label": "做法",
            "value": "去筋洗净，热锅快炒，出锅前调味。",
            "large": true
          },
          {
            "label": "备注",
            "value": "口感保持脆嫩，少放油。",
            "large": false
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "保存草稿",
        "variant": "ghost"
      },
      {
        "text": "保存为可用",
        "variant": "primary"
      }
    ]
  },
  "dishAiParse": {
    "navTitle": "AI 录入",
    "back": true,
    "titleBlock": {
      "title": "AI 辅助录入",
      "desc": "粘贴菜品文字，或上传长截图，让系统先生成草稿。"
    },
    "sections": [
      {
        "id": "ai-entry",
        "type": "tiles",
        "title": "",
        "items": [
          {
            "title": "粘贴文字",
            "desc": "适合从聊天、网页、备忘录复制来的做法。",
            "chips": [
              "文本解析"
            ]
          },
          {
            "title": "上传长截图",
            "desc": "适合菜谱 App 或网页长图。",
            "chips": [
              "图片理解"
            ]
          }
        ]
      },
      {
        "id": "ai-preview",
        "type": "fields",
        "title": "解析预览",
        "more": "可手动修改",
        "fields": [
          {
            "label": "识别菜名",
            "value": "番茄牛腩",
            "large": false
          },
          {
            "label": "分类",
            "value": "荤菜",
            "large": false
          },
          {
            "label": "份量",
            "value": "3 人份",
            "large": false
          },
          {
            "label": "需要补全",
            "value": "配料数量、收汁时间",
            "large": true
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "重新解析",
        "variant": "ghost"
      },
      {
        "text": "填入草稿",
        "variant": "primary"
      }
    ]
  },
  "dishImage": {
    "navTitle": "生成菜品图",
    "back": true,
    "titleBlock": {
      "title": "生成菜品图",
      "desc": "根据菜名和做法生成一张默认展示图。"
    },
    "sections": [
      {
        "id": "image-mock",
        "type": "imageMock",
        "title": ""
      },
      {
        "id": "image-fields",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "生成依据",
            "value": "番茄牛腩，汤汁浓郁，家常晚饭风格",
            "large": true
          },
          {
            "label": "图片风格",
            "value": "自然光 / 家常餐桌 / 浅色背景",
            "large": false
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "换一张",
        "variant": "ghost"
      },
      {
        "text": "采用图片",
        "variant": "primary"
      }
    ]
  },
  "dishDetail": {
    "navTitle": "菜品详情",
    "back": true,
    "detailHero": {
      "title": "清蒸鲈鱼",
      "desc": "清淡 · 2 人份 · 可用",
      "image": "/assets/dishes/fish.png"
    },
    "chips": [
      {
        "text": "清淡",
        "tone": "green"
      },
      {
        "text": "适合晚饭",
        "tone": "orange"
      },
      {
        "text": "孩子爱吃",
        "tone": "blue"
      }
    ],
    "sections": [
      {
        "id": "dish-ingredients",
        "type": "purchaseRows",
        "title": "配料",
        "items": [
          {
            "title": "鲈鱼",
            "desc": "1 条，处理干净",
            "badge": "主料",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "姜葱",
            "desc": "适量，去腥增香",
            "badge": "辅料",
            "mark": "葱",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/cucumber.png"
          }
        ]
      },
      {
        "id": "dish-method",
        "type": "fields",
        "title": "做法",
        "fields": [
          {
            "label": "步骤",
            "value": "鱼身划刀，铺姜葱，上锅蒸 8 分钟；出锅淋热油和蒸鱼豉油。",
            "large": true
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "加入菜谱",
        "variant": "ghost"
      },
      {
        "text": "编辑菜品",
        "variant": "primary",
        "to": "/pages/dish/add"
      }
    ]
  },
  "dishSearch": {
    "navTitle": "搜索菜品",
    "back": true,
    "search": "搜索菜名、分类、食材",
    "chips": [
      {
        "text": "全部",
        "tone": "green"
      },
      {
        "text": "荤菜",
        "tone": "orange"
      },
      {
        "text": "素菜",
        "tone": "blue"
      },
      {
        "text": "汤菜",
        "tone": "green"
      },
      {
        "text": "快手",
        "tone": "green"
      }
    ],
    "sections": [
      {
        "id": "search-result",
        "type": "rows",
        "title": "筛选结果",
        "items": [
          {
            "title": "蒜香鸡翅",
            "desc": "荤菜 · 下饭 · 常点",
            "badge": "查看",
            "mark": "翅",
            "tone": "coral",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/beef-pepper.png"
          },
          {
            "title": "山药排骨汤",
            "desc": "汤菜 · 适合周末",
            "badge": "查看",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "清炒芦笋",
            "desc": "素菜 · 快手",
            "badge": "查看",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          },
          {
            "title": "虾皮蒸蛋",
            "desc": "蛋奶 · 适合孩子",
            "badge": "查看",
            "mark": "蛋",
            "tone": "yellow",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/shrimp-egg.png"
          }
        ]
      }
    ]
  },
  "dishRecommendations": {
    "navTitle": "菜品推荐",
    "back": true,
    "titleBlock": {
      "title": "菜品推荐",
      "desc": "后台精选公开菜品，可复制到自己的菜品库。"
    },
    "sections": [
      {
        "id": "rec-all",
        "type": "recommend",
        "title": "",
        "items": [
          {
            "title": "虾仁蒸蛋",
            "desc": "公开精选 · 适合孩子",
            "mark": "虾",
            "tone": "coral",
            "action": "加入菜品库",
            "image": "/assets/dishes/shrimp-egg.png"
          },
          {
            "title": "南瓜小米粥",
            "desc": "清淡暖胃 · 早餐晚餐都合适",
            "mark": "粥",
            "tone": "yellow",
            "action": "加入菜品库",
            "image": "/assets/dishes/pumpkin-porridge.png"
          },
          {
            "title": "青椒牛柳",
            "desc": "下饭快手 · 家常热菜",
            "mark": "牛",
            "tone": "coral",
            "action": "加入菜品库",
            "image": "/assets/dishes/beef-pepper.png"
          },
          {
            "title": "冬瓜丸子汤",
            "desc": "汤菜 · 适合多人",
            "mark": "汤",
            "tone": "blue",
            "action": "加入菜品库",
            "image": "/assets/dishes/winter-soup.png"
          },
          {
            "title": "凉拌黄瓜",
            "desc": "爽口素菜 · 可提前准备",
            "mark": "瓜",
            "tone": "green",
            "action": "加入菜品库",
            "image": "/assets/dishes/cucumber.png"
          }
        ]
      }
    ],
    "toast": "已加入菜品库"
  },
  "recipeList": {
    "navTitle": "菜谱",
    "tab": "recipe",
    "hero": {
      "title": "我的菜谱",
      "desc": "把自己的菜品整理成不同场景"
    },
    "search": "搜索菜谱名称、备注",
    "actions": [
      {
        "text": "新建菜谱",
        "variant": "primary",
        "to": "/pages/recipe/edit"
      },
      {
        "text": "从菜品选",
        "variant": "ghost",
        "to": "/pages/dish/search"
      }
    ],
    "sections": [
      {
        "id": "recipe-tiles",
        "type": "tiles",
        "title": "",
        "items": [
          {
            "title": "家常晚饭",
            "desc": "6 道菜 · 适合工作日",
            "chips": [
              "清淡",
              "快手"
            ]
          },
          {
            "title": "孩子爱吃",
            "desc": "4 道菜 · 少辣少油",
            "chips": [
              "蒸煮"
            ]
          },
          {
            "title": "周末小聚",
            "desc": "8 道菜 · 有汤有硬菜",
            "chips": [
              "多人"
            ]
          },
          {
            "title": "清淡少油",
            "desc": "5 道菜 · 适合长辈",
            "chips": [
              "软烂"
            ]
          }
        ]
      }
    ]
  },
  "recipeEdit": {
    "navTitle": "新建菜谱",
    "back": true,
    "titleBlock": {
      "title": "新建菜谱",
      "desc": "菜谱是自己的菜品合集，可以写一点口味备注。"
    },
    "sections": [
      {
        "id": "recipe-form",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "菜谱名称",
            "value": "家常晚饭",
            "large": false
          },
          {
            "label": "适用场景",
            "value": "工作日晚餐，2 到 3 人",
            "large": false
          },
          {
            "label": "备注",
            "value": "少油，优先快手菜，汤菜可提前准备。",
            "large": true
          }
        ]
      },
      {
        "id": "recipe-dishes",
        "type": "rows",
        "title": "已选菜品",
        "more": "添加菜品",
        "moreTo": "/pages/dish/search",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "清淡 · 适合晚饭",
            "badge": "已完善",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "汤菜 · 可提前准备",
            "badge": "菜谱内",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "blue",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "快手 · 素菜",
            "badge": "常点",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "orange",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "保存草稿",
        "variant": "ghost"
      },
      {
        "text": "保存菜谱",
        "variant": "primary"
      }
    ]
  },
  "recipeDetail": {
    "navTitle": "菜谱详情",
    "back": true,
    "titleBlock": {
      "title": "家常晚饭",
      "desc": "6 道菜 · 备注：清淡、快手、适合 2 到 3 人。"
    },
    "sections": [
      {
        "id": "recipe-stats",
        "type": "stats",
        "title": "",
        "items": [
          {
            "value": "6",
            "label": "菜品"
          },
          {
            "value": "2",
            "label": "汤菜"
          },
          {
            "value": "25",
            "label": "分钟"
          }
        ]
      },
      {
        "id": "recipe-list",
        "type": "rows",
        "title": "菜品合集",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "清淡 · 2 人份",
            "badge": "查看",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "汤菜 · 可提前准备",
            "badge": "查看",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "快手 · 素菜",
            "badge": "查看",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "编辑菜谱",
        "variant": "ghost",
        "to": "/pages/recipe/edit"
      },
      {
        "text": "用于饭局",
        "variant": "primary",
        "to": "/pages/party/create"
      }
    ]
  },
  "profile": {
    "navTitle": "我的",
    "tab": "profile",
    "profileCard": true,
    "sections": [
      {
        "id": "profile-stats",
        "type": "stats",
        "title": "",
        "items": [
          {
            "value": "18",
            "label": "菜品"
          },
          {
            "value": "4",
            "label": "菜谱"
          },
          {
            "value": "3",
            "label": "饭局"
          }
        ]
      },
      {
        "id": "party-entry",
        "type": "tiles",
        "title": "饭局入口",
        "items": [
          {
            "title": "创建饭局",
            "desc": "生成点餐码，收集大家想吃什么。",
            "chips": []
          },
          {
            "title": "加入饭局",
            "desc": "输入点餐码，选择想吃的菜。",
            "chips": []
          }
        ]
      },
      {
        "id": "profile-list",
        "type": "purchaseRows",
        "title": "常用功能",
        "items": [
          {
            "title": "做菜打卡",
            "desc": "记录做过的菜，积累偏好",
            "badge": "去看看",
            "mark": "打",
            "tone": "green",
            "badgeTone": "gray",
            "action": ""
          },
          {
            "title": "采购清单",
            "desc": "查看已生成的采购任务",
            "badge": "打开",
            "mark": "购",
            "tone": "blue",
            "badgeTone": "gray",
            "action": ""
          },
          {
            "title": "通知中心",
            "desc": "查看审核、积分和饭局提醒",
            "badge": "2 条",
            "mark": "信",
            "tone": "orange",
            "badgeTone": "orange",
            "action": ""
          }
        ]
      }
    ]
  },
  "points": {
    "navTitle": "积分流水",
    "back": true,
    "titleBlock": {
      "title": "当前积分 128",
      "desc": "积分用于部分 AI 能力，失败会自动退还。"
    },
    "sections": [
      {
        "id": "points-list",
        "type": "purchaseRows",
        "title": "本月记录",
        "items": [
          {
            "title": "做菜打卡奖励",
            "desc": "7 月 8 日 19:21",
            "badge": "+5",
            "mark": "+",
            "tone": "green",
            "badgeTone": "green",
            "action": ""
          },
          {
            "title": "生成菜品图",
            "desc": "7 月 7 日 20:04",
            "badge": "-8",
            "mark": "-",
            "tone": "orange",
            "badgeTone": "orange",
            "action": ""
          },
          {
            "title": "AI 失败退还",
            "desc": "7 月 6 日 18:42",
            "badge": "+8",
            "mark": "+",
            "tone": "green",
            "badgeTone": "green",
            "action": ""
          }
        ]
      }
    ]
  },
  "aiRecords": {
    "navTitle": "AI 使用记录",
    "back": true,
    "titleBlock": {
      "title": "AI 使用记录",
      "desc": "查看解析、图片生成、备菜提醒和推荐结果。"
    },
    "sections": [
      {
        "id": "ai-record-list",
        "type": "purchaseRows",
        "title": "最近调用",
        "items": [
          {
            "title": "不知道吃什么",
            "desc": "成功 · 免费次数 1/2",
            "badge": "成功",
            "mark": "AI",
            "tone": "green",
            "badgeTone": "green",
            "action": ""
          },
          {
            "title": "菜品文本解析",
            "desc": "成功 · 未扣积分",
            "badge": "成功",
            "mark": "文",
            "tone": "blue",
            "badgeTone": "green",
            "action": ""
          },
          {
            "title": "生成菜品图",
            "desc": "超时 · 已退还积分",
            "badge": "已退还",
            "mark": "图",
            "tone": "orange",
            "badgeTone": "orange",
            "action": ""
          }
        ]
      }
    ]
  },
  "checkin": {
    "navTitle": "做菜打卡",
    "back": true,
    "titleBlock": {
      "title": "做菜打卡",
      "desc": "记录今天做过的菜，连续打卡会让推荐更懂你。"
    },
    "sections": [
      {
        "id": "calendar",
        "type": "calendar",
        "title": "",
        "days": [
          {
            "label": "1",
            "done": false,
            "today": false
          },
          {
            "label": "2",
            "done": true,
            "today": false
          },
          {
            "label": "3",
            "done": true,
            "today": false
          },
          {
            "label": "4",
            "done": false,
            "today": false
          },
          {
            "label": "5",
            "done": false,
            "today": false
          },
          {
            "label": "6",
            "done": true,
            "today": false
          },
          {
            "label": "7",
            "done": true,
            "today": false
          },
          {
            "label": "8",
            "done": false,
            "today": false
          },
          {
            "label": "9",
            "done": true,
            "today": false
          },
          {
            "label": "10",
            "done": false,
            "today": false
          },
          {
            "label": "11",
            "done": false,
            "today": false
          },
          {
            "label": "12",
            "done": false,
            "today": false
          },
          {
            "label": "13",
            "done": true,
            "today": false
          },
          {
            "label": "14",
            "done": false,
            "today": false
          },
          {
            "label": "15",
            "done": false,
            "today": false
          },
          {
            "label": "16",
            "done": true,
            "today": false
          },
          {
            "label": "17",
            "done": false,
            "today": false
          },
          {
            "label": "18",
            "done": false,
            "today": false
          },
          {
            "label": "19",
            "done": false,
            "today": true
          },
          {
            "label": "20",
            "done": false,
            "today": false
          },
          {
            "label": "21",
            "done": false,
            "today": false
          },
          {
            "label": "22",
            "done": false,
            "today": false
          },
          {
            "label": "23",
            "done": false,
            "today": false
          },
          {
            "label": "24",
            "done": false,
            "today": false
          },
          {
            "label": "25",
            "done": false,
            "today": false
          },
          {
            "label": "26",
            "done": false,
            "today": false
          },
          {
            "label": "27",
            "done": false,
            "today": false
          },
          {
            "label": "28",
            "done": false,
            "today": false
          }
        ]
      },
      {
        "id": "today-checkin",
        "type": "rows",
        "title": "今日记录",
        "more": "新增",
        "items": [
          {
            "title": "清炒荷兰豆",
            "desc": "已获得今日首次打卡积分",
            "badge": "已记录",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          },
          {
            "title": "紫菜蛋花汤",
            "desc": "只记录，不重复奖励",
            "badge": "记录",
            "mark": "汤",
            "tone": "blue",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/egg-soup.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "新增打卡",
        "variant": "primary"
      }
    ]
  },
  "notifications": {
    "navTitle": "通知中心",
    "back": true,
    "titleBlock": {
      "title": "通知中心",
      "desc": "站内通知会保留审核、积分和饭局状态变化。"
    },
    "sections": [
      {
        "id": "unread",
        "type": "purchaseRows",
        "title": "未读",
        "items": [
          {
            "title": "菜品公开审核未通过",
            "desc": "原因：图片不够清晰，可修改后再提交。",
            "badge": "处理",
            "mark": "!",
            "tone": "orange",
            "badgeTone": "orange",
            "action": ""
          },
          {
            "title": "饭局状态变化",
            "desc": "周末晚饭已关闭点菜，等待确认菜单。",
            "badge": "查看",
            "mark": "饭",
            "tone": "green",
            "badgeTone": "green",
            "action": ""
          }
        ]
      },
      {
        "id": "read",
        "type": "purchaseRows",
        "title": "已读",
        "items": [
          {
            "title": "AI 失败退还积分",
            "desc": "生成菜品图超时，积分已返还。",
            "badge": "已读",
            "mark": "AI",
            "tone": "blue",
            "badgeTone": "gray",
            "action": ""
          }
        ]
      }
    ]
  },
  "parties": {
    "navTitle": "我的饭局",
    "back": true,
    "titleBlock": {
      "title": "我的饭局",
      "desc": "饭局入口放在这里，不抢首页菜品库的主路径。"
    },
    "actions": [
      {
        "text": "创建饭局",
        "variant": "primary",
        "to": "/pages/party/create"
      },
      {
        "text": "加入饭局",
        "variant": "ghost",
        "to": "/pages/party/join"
      }
    ],
    "sections": [
      {
        "id": "active-party",
        "type": "purchaseRows",
        "title": "进行中",
        "items": [
          {
            "title": "周末晚饭",
            "desc": "收集中 · 还有 2 小时过期",
            "badge": "点餐码",
            "mark": "饭",
            "tone": "green",
            "badgeTone": "green",
            "action": ""
          },
          {
            "title": "家庭小聚",
            "desc": "待确认菜单 · 5 人参与",
            "badge": "确认",
            "mark": "待",
            "tone": "orange",
            "badgeTone": "orange",
            "action": ""
          }
        ]
      },
      {
        "id": "history-party",
        "type": "purchaseRows",
        "title": "历史饭局",
        "items": [
          {
            "title": "端午午餐",
            "desc": "已生成采购清单 · 8 道菜",
            "badge": "查看",
            "mark": "✓",
            "tone": "blue",
            "badgeTone": "gray",
            "action": ""
          }
        ]
      }
    ]
  },
  "partyCreate": {
    "navTitle": "创建饭局",
    "back": true,
    "titleBlock": {
      "title": "创建饭局",
      "desc": "设置名称、有效期，再选择候选菜品。"
    },
    "sections": [
      {
        "id": "party-form",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "饭局名称",
            "value": "周末晚饭",
            "large": false
          },
          {
            "label": "点餐码有效期",
            "value": "今晚 22:00 前",
            "large": false
          },
          {
            "label": "候选来源",
            "value": "全部菜品 / 从菜谱选 / 常点菜品",
            "large": false
          }
        ]
      },
      {
        "id": "candidate-dishes",
        "type": "rows",
        "title": "候选菜品",
        "more": "继续添加",
        "moreTo": "/pages/dish/search",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "清淡 · 适合晚饭",
            "badge": "已完善",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "汤菜 · 可提前准备",
            "badge": "菜谱内",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "blue",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "快手 · 素菜",
            "badge": "常点",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "orange",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "保存稍后发",
        "variant": "ghost"
      },
      {
        "text": "生成点餐码",
        "variant": "primary",
        "to": "/pages/party/code"
      }
    ]
  },
  "partyCode": {
    "navTitle": "点餐码",
    "back": true,
    "titleBlock": {
      "title": "周末晚饭",
      "desc": "点餐码有效到今晚 22:00，可分享给家人朋友。"
    },
    "sections": [
      {
        "id": "code",
        "type": "code",
        "title": "",
        "code": "4826",
        "desc": "输入点餐码加入饭局"
      },
      {
        "id": "code-stats",
        "type": "stats",
        "title": "",
        "items": [
          {
            "value": "5",
            "label": "候选菜"
          },
          {
            "value": "3",
            "label": "已参与"
          },
          {
            "value": "2h",
            "label": "剩余"
          }
        ]
      },
      {
        "id": "code-notice",
        "type": "notice",
        "title": "当前状态",
        "text": "收集中。创建者可提前关闭点菜，关闭后参与者不能再修改选择。"
      }
    ],
    "bottomActions": [
      {
        "text": "分享点餐码",
        "variant": "ghost"
      },
      {
        "text": "提前关闭",
        "variant": "primary",
        "to": "/pages/party/summary"
      }
    ]
  },
  "partyJoin": {
    "navTitle": "加入饭局",
    "back": true,
    "titleBlock": {
      "title": "加入饭局",
      "desc": "输入别人分享的点餐码，进入候选菜单。"
    },
    "sections": [
      {
        "id": "join-code",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "点餐码",
            "value": "4826",
            "large": false
          }
        ]
      },
      {
        "id": "join-notice",
        "type": "notice",
        "title": "",
        "text": "同一个微信用户不能重复加入同一个饭局；再次输入会回到原参与记录。"
      }
    ],
    "bottomActions": [
      {
        "text": "进入点菜",
        "variant": "primary",
        "to": "/pages/party/order"
      }
    ]
  },
  "partyOrder": {
    "navTitle": "点菜",
    "back": true,
    "titleBlock": {
      "title": "周末晚饭",
      "desc": "点选想吃的菜，最终做几份由创建者确认。"
    },
    "sections": [
      {
        "id": "order-list",
        "type": "rows",
        "title": "",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "2 人想吃 · 清淡",
            "badge": "想吃",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "3 人想吃 · 可提前准备",
            "badge": "已选",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "1 人想吃 · 快手",
            "badge": "想吃",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          },
          {
            "title": "青椒牛柳",
            "desc": "4 人想吃 · 下饭",
            "badge": "想吃",
            "mark": "牛",
            "tone": "coral",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/beef-pepper.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "提交选择",
        "variant": "primary"
      }
    ],
    "toast": "已保存你的选择"
  },
  "partySummary": {
    "navTitle": "确认菜单",
    "back": true,
    "titleBlock": {
      "title": "确认最终菜单",
      "desc": "根据想吃人数给出建议份数，可手动调整。"
    },
    "sections": [
      {
        "id": "summary-list",
        "type": "purchaseRows",
        "title": "",
        "items": [
          {
            "title": "清蒸鲈鱼",
            "desc": "4 人想吃 · 建议 2 份",
            "badge": "2 份",
            "mark": "鱼",
            "tone": "blue",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇鸡汤",
            "desc": "5 人想吃 · 建议 2 份",
            "badge": "2 份",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "1 人想吃 · 可不做",
            "badge": "移除",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      },
      {
        "id": "summary-notice",
        "type": "notice",
        "title": "",
        "text": "确认后会生成饭局快照和采购清单，历史记录不受菜品后续编辑影响。"
      }
    ],
    "bottomActions": [
      {
        "text": "继续调整",
        "variant": "ghost"
      },
      {
        "text": "生成采购清单",
        "variant": "primary",
        "to": "/pages/purchase/list"
      }
    ]
  },
  "purchaseList": {
    "navTitle": "采购清单",
    "back": true,
    "titleBlock": {
      "title": "周末晚饭采购",
      "desc": "按最终制作份数汇总，可勾选、编辑、复制或分享。"
    },
    "sections": [
      {
        "id": "purchase-stats",
        "type": "stats",
        "title": "",
        "items": [
          {
            "value": "9",
            "label": "总项"
          },
          {
            "value": "3",
            "label": "已买"
          },
          {
            "value": "6",
            "label": "待买"
          }
        ]
      },
      {
        "id": "purchase-items",
        "type": "purchaseRows",
        "title": "",
        "items": [
          {
            "title": "鲈鱼",
            "desc": "2 条 · 来自清蒸鲈鱼",
            "badge": "已买",
            "mark": "✓",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "菌菇",
            "desc": "500g · 来自菌菇鸡汤",
            "badge": "待买",
            "mark": "菇",
            "tone": "orange",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "青菜",
            "desc": "300g · 手动新增",
            "badge": "待买",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          },
          {
            "title": "姜葱",
            "desc": "适量 · 多菜合并",
            "badge": "已买",
            "mark": "✓",
            "tone": "blue",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/cucumber.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "复制清单",
        "variant": "ghost"
      },
      {
        "text": "新增采购项",
        "variant": "primary"
      }
    ]
  },
  "aiPrepTips": {
    "navTitle": "AI 备菜提醒",
    "back": true,
    "titleBlock": {
      "title": "备菜提醒",
      "desc": "给出处理优先级和并行建议，不强制变成时间表。"
    },
    "sections": [
      {
        "id": "prep-list",
        "type": "purchaseRows",
        "title": "",
        "items": [
          {
            "title": "先处理菌菇鸡汤",
            "desc": "汤菜可提前炖煮，出餐前保温。",
            "badge": "优先",
            "mark": "高",
            "tone": "coral",
            "badgeTone": "orange",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "鲈鱼最后上锅蒸",
            "desc": "蒸好后口感最佳，临近开饭再做。",
            "badge": "中等",
            "mark": "中",
            "tone": "orange",
            "badgeTone": "orange",
            "action": "",
            "image": "/assets/dishes/fish.png"
          },
          {
            "title": "青菜提前洗切",
            "desc": "沥干水分，最后快炒。",
            "badge": "可提前",
            "mark": "备",
            "tone": "blue",
            "badgeTone": "blue",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      },
      {
        "id": "prep-notice",
        "type": "notice",
        "title": "",
        "text": "提醒仅作参考，修改最终菜单或采购清单不会被这页限制。"
      }
    ],
    "bottomActions": [
      {
        "text": "重新生成",
        "variant": "ghost"
      },
      {
        "text": "保存提醒",
        "variant": "primary"
      }
    ]
  },
  "whatToEat": {
    "navTitle": "不知道吃什么",
    "back": true,
    "titleBlock": {
      "title": "不知道吃什么？",
      "desc": "基于打卡、菜品标签、点菜记录和采购历史推荐。"
    },
    "sections": [
      {
        "id": "what-stats",
        "type": "stats",
        "title": "",
        "items": [
          {
            "value": "2",
            "label": "今日免费"
          },
          {
            "value": "7",
            "label": "打卡天数"
          },
          {
            "value": "18",
            "label": "菜品样本"
          }
        ]
      },
      {
        "id": "what-fields",
        "type": "fields",
        "title": "",
        "fields": [
          {
            "label": "今天偏好",
            "value": "清淡 / 30 分钟内 / 有汤菜",
            "large": false
          },
          {
            "label": "补充说明",
            "value": "今晚 3 个人吃饭，想简单一点。",
            "large": true
          }
        ]
      },
      {
        "id": "what-results",
        "type": "rows",
        "title": "推荐结果",
        "items": [
          {
            "title": "菌菇鸡汤",
            "desc": "可提前准备 · 适合 3 人",
            "badge": "采用",
            "mark": "汤",
            "tone": "orange",
            "badgeTone": "green",
            "action": "",
            "image": "/assets/dishes/mushroom-soup.png"
          },
          {
            "title": "蒜蓉西兰花",
            "desc": "快手素菜 · 配汤刚好",
            "badge": "备选",
            "mark": "菜",
            "tone": "green",
            "badgeTone": "gray",
            "action": "",
            "image": "/assets/dishes/broccoli.png"
          }
        ]
      }
    ],
    "bottomActions": [
      {
        "text": "换一组",
        "variant": "ghost"
      },
      {
        "text": "加入今日菜单",
        "variant": "primary"
      }
    ]
  },
  "disabled": {
    "navTitle": "账号状态",
    "centerState": {
      "title": "账号暂不可用",
      "desc": "当前账号已被禁用，核心功能无法继续使用。请查看原因并按提示联系平台处理。",
      "notice": "禁用原因：内容违规处理未完成。已有菜品、饭局和采购记录不会自动删除。"
    }
  }
};

module.exports = screens;
