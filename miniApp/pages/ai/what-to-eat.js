const api = require("../../services/api");
const { go } = require("../../utils/navigation");

Page({
  data: {
    status: null,
    selectedTags: ["下饭", "少油"],
    tagOptions: ["下饭", "少油", "清淡", "快手", "适合孩子", "有汤"],
    people: 2,
    timeLimit: "30 分钟内",
    result: null,
    generating: false,
  },

  async onLoad() {
    const status = await api.getWhatToEatStatus();
    this.setData({ status });
  },

  toggleTag(event) {
    const tag = event.currentTarget.dataset.tag;
    const selectedTags = [...this.data.selectedTags];
    const index = selectedTags.indexOf(tag);
    if (index >= 0) selectedTags.splice(index, 1);
    else selectedTags.push(tag);
    this.setData({ selectedTags });
  },

  choosePeople(event) {
    this.setData({ people: Number(event.detail.value) + 1 });
  },

  chooseTime(event) {
    const options = ["15 分钟内", "30 分钟内", "45 分钟内", "不限时间"];
    this.setData({ timeLimit: options[event.detail.value] });
  },

  generate() {
    if (!this.data.status.unlocked) return;
    if (this.data.status.freeQuota > 0) {
      this.runGenerate(true);
      return;
    }
    wx.showModal({
      title: "使用 8 积分生成推荐？",
      content: `当前有 ${this.data.status.points} 积分。生成失败或超时会自动退还。`,
      confirmText: "确认生成",
      confirmColor: "#159B55",
      success: ({ confirm }) => {
        if (confirm) this.runGenerate(false);
      },
    });
  },

  async runGenerate(useFreeQuota) {
    this.setData({ generating: true });
    try {
      const result = await api.generateWhatToEat({
        tags: this.data.selectedTags,
        people: this.data.people,
        timeLimit: this.data.timeLimit,
        useFreeQuota,
      });
      this.setData({ result, "status.freeQuota": useFreeQuota ? 0 : this.data.status.freeQuota });
    } finally {
      this.setData({ generating: false });
    }
  },

  openDetail() {
    go("/pages/ai/what-to-eat-detail", { id: this.data.result.id, source: this.data.result.source });
  },

  regenerate() {
    this.generate();
  },

  ignore() {
    this.setData({ result: null });
  },
});
