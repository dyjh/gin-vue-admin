Component({
  properties: {
    image: { type: String, value: "" },
    title: { type: String, value: "" },
    subtitle: { type: String, value: "" },
    back: { type: Boolean, value: true },
    height: { type: Number, value: 448 },
  },
  data: {
    navHeight: 88,
    menuTop: 24,
    menuHeight: 32,
  },
  lifetimes: {
    attached() {
      const app = getApp();
      this.setData(app.globalData.nav || {});
    },
  },
  methods: {
    goBack() {
      const pages = getCurrentPages();
      if (pages.length > 1) {
        wx.navigateBack();
        return;
      }
      wx.switchTab({ url: "/pages/home/index" });
    },
  },
});
