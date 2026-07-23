const { seed } = require("./data");

const KEY = "order-food-miniapp-state-v1";

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function ensureSeeded() {
  const current = wx.getStorageSync(KEY);
  if (!current || !current.dishes) {
    wx.setStorageSync(KEY, clone(seed));
    return;
  }
  if (!current.mealHistories) {
    current.mealHistories = clone(seed.mealHistories);
    wx.setStorageSync(KEY, current);
  }
}

function getState() {
  ensureSeeded();
  return clone(wx.getStorageSync(KEY));
}

function saveState(state) {
  wx.setStorageSync(KEY, clone(state));
  return clone(state);
}

function mutate(updater) {
  const state = getState();
  const next = updater(state) || state;
  return saveState(next);
}

function reset() {
  return saveState(seed);
}

module.exports = { ensureSeeded, getState, saveState, mutate, reset };
