Component({
  properties: {
    icon: { type: String, value: "inbox" },
    title: { type: String, value: "这里还没有内容" },
    description: { type: String, value: "" },
    action: { type: String, value: "" },
  },
  methods: {
    act() {
      this.triggerEvent("action");
    },
  },
});
