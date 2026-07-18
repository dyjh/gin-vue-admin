const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    tab: "manual",
    aiText: "",
    aiImage: "",
    recognizing: false,
    recognition: null,
    form: {
      name: "",
      category: "家常菜",
      tags: [],
      serving: 2,
      description: "",
      image: "",
      ingredients: [{ id: "ingredient-1", name: "", amount: "" }],
      steps: [{ id: "step-1", text: "", image: "" }],
    },
  },

  onLoad(options) {
    this.setData({ tab: options.tab === "ai" ? "ai" : "manual" });
  },

  switchTab(event) {
    this.setData({ tab: event.currentTarget.dataset.tab });
  },

  inputAiText(event) {
    this.setData({ aiText: event.detail.value });
  },

  chooseRecognitionImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      success: (result) => this.setData({ aiImage: result.tempFiles[0].tempFilePath }),
    });
  },

  async recognize() {
    if (!this.data.aiText.trim() && !this.data.aiImage) {
      wx.showToast({ title: "粘贴菜谱文字或上传截图", icon: "none" });
      return;
    }
    this.setData({ recognizing: true });
    try {
      const recognition = await api.parseDish({ text: this.data.aiText, image: this.data.aiImage });
      this.setData({ recognition });
    } finally {
      this.setData({ recognizing: false });
    }
  },

  smartFill() {
    this.setData({ form: this.data.recognition, tab: "manual" });
    wx.showToast({ title: "已智能填充", icon: "success" });
  },

  async save(event) {
    const dish = await api.createDish(event.detail.form);
    wx.showToast({ title: event.detail.form.status === "draft" ? "草稿已保存" : "菜品已保存", icon: "success" });
    setTimeout(() => go("/pages/dish/detail", { id: dish.id }), 300);
  },
});
