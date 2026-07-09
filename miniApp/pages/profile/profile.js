Page({
  data: {
    actions: [
      { id: 1, name: '创建饭局', desc: '发起一次点菜收集' },
      { id: 2, name: '加入饭局', desc: '通过点菜码参与' },
      { id: 3, name: '采购清单', desc: '查看待买和已买食材' },
      { id: 4, name: '做菜打卡', desc: '记录今天做了什么' }
    ]
  },

  onActionTap(event) {
    wx.showToast({
      title: event.currentTarget.dataset.name,
      icon: 'none'
    })
  },

  goHome() {
    wx.redirectTo({
      url: '/pages/home/home'
    })
  },

  goRecipes() {
    wx.redirectTo({
      url: '/pages/recipe/list'
    })
  }
})


