const api = require("../../services/api");

function buildCalendar(checkins) {
  const checked = new Set(checkins.map((item) => Number(item.date.slice(-2))));
  const blanks = Array.from({ length: 3 }, (_, index) => ({ key: `blank-${index}`, day: 0 }));
  const days = Array.from({ length: 31 }, (_, index) => ({
    key: `day-${index + 1}`,
    day: index + 1,
    checked: checked.has(index + 1),
    today: index + 1 === 18,
  }));
  return [...blanks, ...days];
}

Page({
  data: {
    checkins: [],
    calendar: [],
    showAdd: false,
    dishes: [],
    dishIndex: 0,
    photo: "",
    note: "",
  },

  onLoad() {
    this.load();
  },

  async load() {
    const [checkins, dishes] = await Promise.all([
      api.listCheckins({ page: 1, pageSize: 20 }),
      api.listDishes({ page: 1, pageSize: 100, status: "usable" }),
    ]);
    this.setData({
      checkins: checkins.list,
      calendar: buildCalendar(checkins.list),
      dishes: dishes.list,
    });
  },

  openAdd() {
    this.setData({ showAdd: true, photo: "", note: "", dishIndex: 0 });
  },

  closeAdd() {
    this.setData({ showAdd: false });
  },

  noop() {},

  chooseDish(event) {
    this.setData({ dishIndex: Number(event.detail.value) });
  },

  choosePhoto() {
    wx.chooseMedia({
      count: 1,
      mediaType: ["image"],
      sourceType: ["album", "camera"],
      success: (result) => this.setData({ photo: result.tempFiles[0].tempFilePath }),
    });
  },

  inputNote(event) {
    this.setData({ note: event.detail.value });
  },

  async submit() {
    if (!this.data.photo) {
      wx.showToast({ title: "请添加本次做菜照片", icon: "none" });
      return;
    }
    const dish = this.data.dishes[this.data.dishIndex];
    const result = await api.createCheckin({
      dishId: dish.id,
      dishName: dish.name,
      image: this.data.photo,
      note: this.data.note,
    });
    this.setData({ showAdd: false });
    await this.load();
    wx.showModal({
      title: "打卡成功",
      content: result.rewardPoints ? `今天首次打卡，获得 ${result.rewardPoints} 积分。` : "今天的做菜记录已保存。",
      showCancel: false,
      confirmColor: "#159B55",
    });
  },
});
