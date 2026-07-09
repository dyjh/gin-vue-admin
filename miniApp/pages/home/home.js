Page({
  data: {
    categories: ['全部', '家常', '快手', '下饭', '清淡'],
    activeCategory: '全部',
    myDishes: [
      { id: 1, name: '番茄牛腩', meta: '适合 3-4 人', mark: '牛', tags: ['下饭', '可提前炖'] },
      { id: 2, name: '清炒时蔬', meta: '10 分钟快手', mark: '青', tags: ['清淡', '快手'] },
      { id: 3, name: '蒜香排骨', meta: '周末硬菜', mark: '排', tags: ['家常', '孩子爱吃'] },
      { id: 4, name: '虾仁蒸蛋', meta: '软嫩少油', mark: '蛋', tags: ['清淡', '老人友好'] },
      { id: 5, name: '菌菇鸡汤', meta: '适合晚餐', mark: '汤', tags: ['汤菜', '暖胃'] },
      { id: 6, name: '葱油拌面', meta: '一人食也方便', mark: '面', tags: ['快手', '主食'] }
    ],
    recommendations: [
      {
        id: 101,
        name: '香菇滑鸡',
        desc: '鸡腿肉和香菇同蒸，省油烟也适合工作日晚餐。',
        mark: '鸡',
        reason: '家常高匹配',
        tags: ['下饭', '少油烟']
      },
      {
        id: 102,
        name: '南瓜蒸排骨',
        desc: '软糯南瓜配排骨，老人孩子都更容易接受。',
        mark: '南',
        reason: '适合多人',
        tags: ['蒸菜', '软烂']
      },
      {
        id: 103,
        name: '芦笋虾仁',
        desc: '清爽高蛋白，适合和重口味菜搭配平衡。',
        mark: '虾',
        reason: '清爽搭配',
        tags: ['清淡', '快手']
      },
      {
        id: 104,
        name: '土豆焖鸡翅',
        desc: '食材常见，汤汁拌饭很稳，适合临时加菜。',
        mark: '翅',
        reason: '备菜简单',
        tags: ['下饭', '家常']
      },
      {
        id: 105,
        name: '冬瓜丸子汤',
        desc: '有汤有菜，口味轻，适合晚餐收尾。',
        mark: '汤',
        reason: '补充汤菜',
        tags: ['汤菜', '清淡']
      }
    ]
  },

  onCategoryTap(event) {
    this.setData({
      activeCategory: event.currentTarget.dataset.name
    })
  },

  onSearchTap() {
    wx.showToast({
      title: '搜索功能稍后接入',
      icon: 'none'
    })
  },

  onAddDish() {
    wx.showToast({
      title: '进入添加菜品',
      icon: 'none'
    })
  },

  onViewAllDishes() {
    wx.showToast({
      title: '查看全部菜品',
      icon: 'none'
    })
  },

  onViewAllRecommended() {
    wx.showToast({
      title: '查看更多推荐',
      icon: 'none'
    })
  },

  onAddRecommendedDish(event) {
    wx.showToast({
      title: `${event.currentTarget.dataset.name} 已加入`,
      icon: 'none'
    })
  },

  onAskAI() {
    wx.showToast({
      title: '开始帮你挑菜',
      icon: 'none'
    })
  },

  goRecipes() {
    wx.redirectTo({
      url: '/pages/recipe/list'
    })
  },

  goProfile() {
    wx.redirectTo({
      url: '/pages/profile/profile'
    })
  }
})

