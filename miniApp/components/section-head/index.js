Component({
  properties: {
    title: { type: String, value: "" },
    subtitle: { type: String, value: "" },
    action: { type: String, value: "" },
  },
  methods: {
    tapAction() {
      this.triggerEvent("action");
    },
  },
});
