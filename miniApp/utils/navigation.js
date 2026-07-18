function go(path, query) {
  const pairs = Object.keys(query || {})
    .filter((key) => query[key] !== undefined && query[key] !== null)
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(query[key])}`);
  const url = pairs.length ? `${path}?${pairs.join("&")}` : path;
  wx.navigateTo({ url });
}

function home() {
  wx.switchTab({ url: "/pages/home/index" });
}

module.exports = { go, home };
