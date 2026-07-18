const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    meal: null,
    items: [],
    scope: "pending",
    pendingCount: 0,
    completedCount: 0,
    showEditor: false,
    editingId: "",
    form: { name: "", amount: "", note: "" },
  },

  onLoad() {
    this.load();
  },

  async load() {
    const data = await api.getShoppingList();
    this.setData(data);
  },

  setScope(event) {
    this.setData({ scope: event.currentTarget.dataset.scope });
  },

  async toggle(event) {
    const id = event.currentTarget.dataset.id;
    const item = this.data.items.find((entry) => entry.id === id);
    await api.updateShoppingItem(id, { completed: !item.completed });
    await this.load();
  },

  openEdit(event) {
    const id = event.currentTarget.dataset.id;
    const item = this.data.items.find((entry) => entry.id === id);
    this.setData({
      showEditor: true,
      editingId: id,
      form: { name: item.name, amount: item.amount, note: item.note || "" },
    });
  },

  openAdd() {
    this.setData({ showEditor: true, editingId: "", form: { name: "", amount: "", note: "" } });
  },

  closeEditor() {
    this.setData({ showEditor: false });
  },

  noop() {},

  inputField(event) {
    const field = event.currentTarget.dataset.field;
    this.setData({ [`form.${field}`]: event.detail.value });
  },

  async saveItem() {
    if (!this.data.form.name.trim()) {
      wx.showToast({ title: "请填写采购项名称", icon: "none" });
      return;
    }
    if (this.data.editingId) await api.updateShoppingItem(this.data.editingId, this.data.form);
    else await api.addShoppingItem(this.data.form);
    this.setData({ showEditor: false });
    await this.load();
    wx.showToast({ title: this.data.editingId ? "采购项已更新" : "采购项已添加", icon: "success" });
  },

  async removeItem() {
    await api.deleteShoppingItem(this.data.editingId);
    this.setData({ showEditor: false });
    await this.load();
    wx.showToast({ title: "采购项已删除", icon: "success" });
  },

  copy() {
    const text = this.data.items
      .filter((item) => !item.completed)
      .map((item) => `□ ${item.name} ${item.amount}`)
      .join("\n");
    wx.setClipboardData({ data: `${this.data.meal.name}采购清单\n${text}` });
  },

  openPrep() {
    go("/pages/meal/prep-ai");
  },

  onShareAppMessage() {
    return { title: `${this.data.meal.name}采购清单`, path: "/pages/shopping/list" };
  },
});
