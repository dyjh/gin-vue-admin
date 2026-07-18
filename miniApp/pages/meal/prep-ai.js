const api = require("../../services/api");

Page({
  data: {
    plan: null,
    generating: false,
    points: 128,
    cost: 6,
  },

  generate() {
    wx.showModal({
      title: "使用 6 积分生成备菜顺序？",
      content: "AI 会根据最终菜单，把等待时间穿插起来。生成失败或超时会自动退还积分。",
      confirmText: "确认生成",
      confirmColor: "#159B55",
      success: ({ confirm }) => {
        if (confirm) this.runGenerate();
      },
    });
  },

  async runGenerate() {
    this.setData({ generating: true });
    try {
      const plan = await api.generatePrepPlan({ mealId: "meal-1" });
      this.setData({ plan });
    } finally {
      this.setData({ generating: false });
    }
  },

  ignore() {
    this.setData({ plan: null });
  },

  regenerate() {
    this.generate();
  },
});
