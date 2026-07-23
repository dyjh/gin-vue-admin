const api = require("../../services/api");
const { home } = require("../../utils/navigation");

function normalizeCode(value) {
  if (value === undefined || value === null) return "";
  return String(value).replace(/\D/g, "").slice(0, 6);
}

Page({
  data: {
    state: "collecting",
    code: "",

    checkingCode: false,
    codeMatched: false,
    matchedMealName: "",
    codeError: "",
    meal: null,    categories: [],
    category: "全部",
    selectedIds: [],
    joining: false,
    saving: false,
  },

  async onLoad(options) {
    const code = normalizeCode(options && options.code);
    this.setData({ state: options.state || "collecting", code });
    if (options.state !== "join") await this.load();
    else if (code.length === 6) await this.lookupCode(code);
  },
  async load() {
    const meal = await api.getCurrentMeal();
    const categories = ["全部", ...Array.from(new Set(meal.candidates.map((dish) => dish.category)))];
    this.setData({
      meal,
      categories,
      selectedIds: [...meal.selectedIds],
      state: meal.status === "cancelled" ? "cancelled" : meal.status === "collecting" ? "collecting" : "closed",
    });
  },

  inputCode(event) {
    const code = normalizeCode(event && event.detail && event.detail.value);
    this.setData({ code, codeMatched: false, matchedMealName: "", codeError: "", checkingCode: false });
    if (code.length === 6) this.lookupCode(code);
    return code;
  },


  async lookupCode(code) {
    this.setData({ checkingCode: true, codeMatched: false, codeError: "" });
    try {
      const preview = await api.previewMealByCode({ code });
      if (this.data.code !== code) return false;
      if (preview.status !== "collecting") {
        this.setData({ checkingCode: false, codeError: preview.status === "cancelled" ? "这个饭局已取消" : "这个饭局已停止点餐" });
        return false;
      }
      this.setData({ checkingCode: false, codeMatched: true, matchedMealName: preview.name });
      return true;
    } catch (error) {
      if (this.data.code === code) this.setData({ checkingCode: false, codeMatched: false, codeError: "没有找到这个饭局，请检查点餐码" });
      return false;
    }
  },

  async join() {
    if (this.data.code.length !== 6) {
      wx.showToast({ title: "请输入 6 位点餐码", icon: "none" });
      return;
    }
    const matched = this.data.codeMatched || await this.lookupCode(this.data.code);
    if (!matched) return;
    this.setData({ joining: true });
    try {
      const meal = await api.joinMeal({ code: this.data.code });
      const categories = ["全部", ...Array.from(new Set(meal.candidates.map((dish) => dish.category)))];
      this.setData({ meal, categories, selectedIds: [...meal.selectedIds], state: "collecting" });
      wx.showToast({ title: "已加入饭局", icon: "success" });
    } catch (error) {
      // The request layer already presents the reason.
    } finally {
      this.setData({ joining: false });
    }
  },
  selectCategory(event) {
    this.setData({ category: event.currentTarget.dataset.category });
  },

  toggleDish(event) {
    const id = event.detail.id;
    const selectedIds = [...this.data.selectedIds];
    const index = selectedIds.indexOf(id);
    if (index >= 0) selectedIds.splice(index, 1);
    else selectedIds.push(id);
    this.setData({ selectedIds });
  },

  async save() {
    this.setData({ saving: true });
    try {
      await api.saveVotes(this.data.meal.id, this.data.selectedIds);
      wx.showToast({ title: "点菜已保存", icon: "success" });
    } finally {
      this.setData({ saving: false });
    }
  },

  home,
});
