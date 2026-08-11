const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { getFeature } = require("../../utils/features");

Page({
  data: {
    tab: "manual",
    sourceText: "",
    sourceImage: "",
    recognizing: false,
    recognition: null,
    extractFeature: null,
    coverFeature: null,
    form: {
      name: "",
      category: "",
      tags: [],
      serving: 2,
      description: "",
      coverUrl: "",
      coverFileId: "",
      ingredients: [{ id: "ingredient-1", name: "", amount: "", unit: "克", note: "" }],
      steps: [{ id: "step-1", text: "", imageUrl: "", imageFileId: "" }],
    },
  },

  async onLoad(options) {
    let name = options.name || "";
    try {
      name = decodeURIComponent(name);
    } catch (error) {
      name = options.name || "";
    }
    await api.getRuntimeConfig();
    const extractFeature = getFeature("dish_extract");
    const coverFeature = getFeature("cover_create");
    const data = {
      tab: options.tab === "extract" && extractFeature ? "extract" : "manual",
      extractFeature,
      coverFeature,
    };
    if (name) data["form.name"] = name;
    this.setData(data);
  },

  switchTab(event) {
    this.setData({ tab: event.currentTarget.dataset.tab });
  },

  inputSourceText(event) {
    this.setData({ sourceText: event.detail.value });
  },

  chooseRecognitionImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      success: (result) => this.setData({ sourceImage: result.tempFiles[0].tempFilePath }),
    });
  },

  async recognize() {
    if (!this.data.sourceText.trim() && !this.data.sourceImage) {
      wx.showToast({ title: "粘贴菜谱文字或上传截图", icon: "none" });
      return;
    }
    this.setData({ recognizing: true });
    try {
      const result = await api.parseDish({ text: this.data.sourceText, image: this.data.sourceImage });
      this.setData({ recognition: result.dishDraft });
    } finally {
      this.setData({ recognizing: false });
    }
  },

  smartFill() {
    this.setData({ form: this.data.recognition, tab: "manual" });
    wx.showToast({ title: "内容已填入表单", icon: "success" });
  },

  async save(event) {
    const dish = await api.createDish(event.detail.form);
    wx.showToast({ title: event.detail.form.status === "draft" ? "草稿已保存" : "菜品已保存", icon: "success" });
    setTimeout(() => go("/pages/dish/detail", { id: dish.id }), 300);
  },
});
