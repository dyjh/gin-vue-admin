const { resolveAssetUrl } = require("../../utils/assets");

function iconSource(name, tone) {
  return resolveAssetUrl(`/assets/icons/${name}-${tone}.png`);
}

Component({
  properties: {
    name: { type: String, value: "chevron-right" },
    tone: { type: String, value: "green" },
    size: { type: Number, value: 20 },
  },
  data: { src: "" },
  observers: {
    "name,tone": function observe(name, tone) {
      this.setData({ src: iconSource(name, tone) });
    },
  },
  lifetimes: {
    attached() {
      this.setData({ src: iconSource(this.data.name, this.data.tone) });
    },
  },
});
