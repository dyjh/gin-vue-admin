Component({
  properties: {
    dish: { type: Object, value: {} },
    action: { type: String, value: "" },
    selected: { type: Boolean, value: false },
    compact: { type: Boolean, value: false },
  },
  methods: {
    open() {
      this.triggerEvent("open", { id: this.data.dish.id });
    },
    act() {
      this.triggerEvent("action", { id: this.data.dish.id, selected: this.data.selected });
    },
  },
});
