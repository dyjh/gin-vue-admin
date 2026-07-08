Component({
  properties: {
    screen: {
      type: Object,
      value: {}
    }
  },
  methods: {
    goBack() {
      const pages = getCurrentPages();
      if (pages.length > 1) {
        wx.navigateBack();
        return;
      }
      wx.redirectTo({ url: "/pages/home/home" });
    },
    handleNav(event) {
      const to = event.currentTarget.dataset.to;
      if (!to) return;
      wx.navigateTo({ url: to });
    },
    handleRedirect(event) {
      const to = event.currentTarget.dataset.to;
      if (!to) return;
      wx.redirectTo({ url: to });
    },
    noop() {}
  }
});