const api = require("../../services/api");
const { go } = require("../../utils/navigation");
const { SHARE_IMAGES } = require("../../config/share");

const UNIT_OPTIONS = ["个", "克", "千克", "斤", "颗", "袋", "盒", "瓶", "把", "根", "瓣", "包", "毫升", "升"];
const UNITS_BY_LENGTH = [...UNIT_OPTIONS].sort((left, right) => right.length - left.length);

function splitAmount(value = "") {
  const amount = String(value).trim();
  const unit = UNITS_BY_LENGTH.find((candidate) => amount.endsWith(candidate)) || "";
  return {
    quantity: unit ? amount.slice(0, -unit.length).trim() : amount,
    unit,
  };
}

Page({
  data: {
    shareToken: "",
    meal: null,
    items: [],
    scope: "pending",
    pendingCount: 0,
    completedCount: 0,
    showEditor: false,
    editingId: "",
    unitOptions: UNIT_OPTIONS,
    unitIndex: 0,
    form: { name: "", quantity: "", unit: "", note: "" },
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
    const amount = splitAmount(item.amount);
    this.setData({
      showEditor: true,
      editingId: id,
      unitIndex: Math.max(UNIT_OPTIONS.indexOf(amount.unit), 0),
      form: { name: item.name, quantity: amount.quantity, unit: amount.unit, note: item.note || "" },
    });
  },

  openAdd() {
    this.setData({
      showEditor: true,
      editingId: "",
      unitIndex: 0,
      form: { name: "", quantity: "", unit: "", note: "" },
    });
  },

  closeEditor() {
    this.setData({ showEditor: false });
  },

  noop() {},

  inputField(event) {
    const field = event.currentTarget.dataset.field;
    this.setData({ [`form.${field}`]: event.detail.value });
  },

  selectUnit(event) {
    const unitIndex = Number(event.detail.value);
    this.setData({ unitIndex, "form.unit": UNIT_OPTIONS[unitIndex] });
  },

  async saveItem() {
    const { name, quantity, unit, note } = this.data.form;
    if (!name.trim()) {
      wx.showToast({ title: "请填写采购项名称", icon: "none" });
      return;
    }
    if (!quantity.trim()) {
      wx.showToast({ title: "请填写数量", icon: "none" });
      return;
    }
    if (!unit) {
      wx.showToast({ title: "请选择单位", icon: "none" });
      return;
    }

    const editingId = this.data.editingId;
    const payload = {
      name: name.trim(),
      amount: `${quantity.trim()} ${unit}`,
      note: note.trim(),
    };
    if (editingId) await api.updateShoppingItem(editingId, payload);
    else await api.addShoppingItem(payload);
    this.setData({ showEditor: false, editingId: "" });
    await this.load();
    wx.showToast({ title: editingId ? "采购项已更新" : "采购项已添加", icon: "success" });
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
    const token = encodeURIComponent(this.data.shareToken || "");
    return {
      title: `${this.data.meal.name}采购清单`,
      path: token ? `/pages/shopping/share?token=${token}` : "/pages/shopping/share",
      imageUrl: SHARE_IMAGES.shoppingList,
    };
  },
});
