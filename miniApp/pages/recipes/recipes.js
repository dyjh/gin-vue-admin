Page({
  data: {
    recipes: [
      { id: 1, name: '周末家常菜', count: 8, note: '适合三四个人一起吃' },
      { id: 2, name: '清淡晚餐', count: 5, note: '少油、软烂、好消化' }
    ]
  },

  onCreateRecipe() {
    wx.showToast({
      title: '创建菜谱稍后接入',
      icon: 'none'
    })
  },

  goHome() {
    wx.redirectTo({
      url: '/pages/home/home'
    })
  },

  goProfile() {
    wx.redirectTo({
      url: '/pages/profile/profile'
    })
  }
})
