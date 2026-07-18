Component({
  properties: {
    name: { type: String, value: "chevron-right" },
    tone: { type: String, value: "green" },
    size: { type: Number, value: 20 },
  },
  data: { src: "" },
  observers: {
    "name,tone": function observe(name, tone) {
      this.setData({ src: `/assets/icons/${name}-${tone}.png` });
    },
  },
  lifetimes: {
    attached() {
      this.setData({ src: `/assets/icons/${this.data.name}-${this.data.tone}.png` });
    },
  },
});
