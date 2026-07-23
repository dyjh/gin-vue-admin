const fs = require("fs");
const path = require("path");
const sharp = require("sharp");

const root = path.resolve(__dirname, "..");
const iconDir = path.join(root, "assets", "icons");
const tabDir = path.join(root, "assets", "tabbar");
const imageDir = path.join(root, "assets", "images");
const sourceDir = path.join(root, ".preview", "assets");

for (const directory of [iconDir, tabDir, imageDir]) {
  fs.mkdirSync(directory, { recursive: true });
}

for (const file of fs.readdirSync(sourceDir)) {
  if (file.toLowerCase().endsWith(".png")) {
    fs.copyFileSync(path.join(sourceDir, file), path.join(imageDir, file));
  }
}

const shapes = {
  "arrow-left": '<path d="M19 6l-9 9 9 9M10 15h15"/>',
  "chevron-right": '<path d="M12 6l9 9-9 9"/>',
  plus: '<path d="M15 6v18M6 15h18"/>',
  minus: '<path d="M6 15h18"/>',
  check: '<path d="M6 15l6 6L25 8"/>',
  x: '<path d="M8 8l14 14M22 8L8 22"/>',
  close: '<path d="M8 8l14 14M22 8L8 22"/>',
  search: '<circle cx="13" cy="13" r="8"/><path d="M19 19l7 7"/>',
  sparkles: '<path d="M12 3l2.2 6.8L21 12l-6.8 2.2L12 21l-2.2-6.8L3 12l6.8-2.2zM24 19l1.2 3.8L29 24l-3.8 1.2L24 29l-1.2-3.8L19 24l3.8-1.2z"/>',
  image: '<rect x="4" y="5" width="24" height="22" rx="3"/><circle cx="11" cy="12" r="2"/><path d="M5 23l7-7 5 5 3-3 7 7"/>',
  camera: '<path d="M5 10h5l2-3h6l2 3h7v16H5z"/><circle cx="16" cy="18" r="5"/>',
  trash: '<path d="M6 9h20M12 5h8l1 4M10 9l1 18h10l1-18M14 13v9M18 13v9"/>',
  edit: '<path d="M6 24l1-6L21 4l5 5-14 14zM18 7l5 5"/>',
  copy: '<rect x="10" y="10" width="16" height="17" rx="2"/><path d="M22 10V5H6v17h4"/>',
  share: '<circle cx="7" cy="16" r="3"/><circle cx="24" cy="7" r="3"/><circle cx="24" cy="25" r="3"/><path d="M10 15l11-6M10 17l11 6"/>',
  clock: '<circle cx="16" cy="16" r="12"/><path d="M16 9v8l5 3"/>',
  calendar: '<rect x="5" y="7" width="22" height="20" rx="3"/><path d="M10 4v6M22 4v6M5 13h22"/>',
  cart: '<path d="M4 6h4l3 14h12l3-10H9M13 26h.1M23 26h.1"/>',
  bell: '<path d="M8 23h16l-3-4v-5a5 5 0 0 0-10 0v5zM13 26c1 3 5 3 6 0"/>',
  book: '<path d="M4 6c6-2 10 0 12 3v19c-2-3-6-5-12-3zM28 6c-6-2-10 0-12 3v19c2-3 6-5 12-3z"/>',
  home: '<path d="M4 15L16 5l12 10v13H19v-8h-6v8H4z"/>',
  user: '<circle cx="16" cy="11" r="6"/><path d="M5 28c1-8 6-11 11-11s10 3 11 11"/>',
  dish: '<circle cx="16" cy="16" r="12"/><circle cx="16" cy="16" r="6"/><path d="M16 4v3"/>',
  lock: '<rect x="7" y="14" width="18" height="14" rx="3"/><path d="M11 14V9a5 5 0 0 1 10 0v5"/>',
  unlock: '<rect x="7" y="14" width="18" height="14" rx="3"/><path d="M21 14V9a5 5 0 0 0-9-3"/>',
  filter: '<path d="M5 7h22L19 16v9l-6 3V16z"/>',
  info: '<circle cx="16" cy="16" r="12"/><path d="M16 14v8M16 9h.1"/>',
  refresh: '<path d="M26 10V4l-4 4a11 11 0 1 0 3 13M26 4h-6"/>',
  send: '<path d="M4 5l25 11L4 27l5-11zM9 16h12"/>',
  award: '<circle cx="16" cy="13" r="8"/><path d="M11 20L9 29l7-4 7 4-2-9"/>',
  list: '<path d="M11 8h17M11 16h17M11 24h17M5 8h.1M5 16h.1M5 24h.1"/>',
  eye: '<path d="M3 16s5-8 13-8 13 8 13 8-5 8-13 8S3 16 3 16z"/><circle cx="16" cy="16" r="3"/>',
  "eye-off": '<path d="M4 4l24 24M10 8c2-1 4-2 6-2 8 0 13 10 13 10a21 21 0 0 1-4 5M7 10c-3 3-4 6-4 6s5 10 13 10c2 0 4-1 6-2"/>',
  "circle-check": '<circle cx="16" cy="16" r="12"/><path d="M9 16l5 5 9-11"/>',
  "circle-stop-solid": '<circle cx="16" cy="16" r="12"/><rect x="12" y="12" width="8" height="8" rx="1" fill="currentColor" stroke="none"/>',
  utensils: '<path d="M7 4v10M11 4v10M5 9h8M9 14v14M22 4v24M22 4c5 4 5 10 0 13"/>',
  inbox: '<path d="M5 6h22v20H5zM5 18h6l2 4h6l2-4h6"/>',
  upload: '<path d="M16 22V5M9 12l7-7 7 7M5 24v4h22v-4"/>',
  link: '<path d="M13 20l-2 2a5 5 0 0 1-7-7l5-5a5 5 0 0 1 7 0M19 12l2-2a5 5 0 0 1 7 7l-5 5a5 5 0 0 1-7 0M11 16h10"/>',
  "qr-code": '<path d="M5 5h8v8H5zM19 5h8v8h-8zM5 19h8v8H5zM19 19h3v3h-3zM24 19h3v8h-5M19 24h3"/>'
};

const tones = {
  dark: "#1B241F",
  green: "#159B55",
  gray: "#8F9A94",
  white: "#FFFFFF",
  danger: "#D8564F"
};

function svg(shape, color, size = 48) {
  return Buffer.from(`<svg width="${size}" height="${size}" viewBox="0 0 32 32" color="${color}" xmlns="http://www.w3.org/2000/svg"><g fill="none" stroke="${color}" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round">${shape}</g></svg>`);
}

async function build() {
  const jobs = [];
  for (const [name, shape] of Object.entries(shapes)) {
    for (const [tone, color] of Object.entries(tones)) {
      const size = name === "circle-stop-solid" ? 96 : 48;
      jobs.push(sharp(svg(shape, color, size)).png().toFile(path.join(iconDir, `${name}-${tone}.png`)));
    }
  }
  jobs.push(sharp(svg(shapes.award, "#5795AD")).png().toFile(path.join(iconDir, "award-blue.png")));

  for (const name of ["home", "book", "user"]) {
    jobs.push(sharp(svg(shapes[name], tones.gray, 81)).png().toFile(path.join(tabDir, `${name}.png`)));
    jobs.push(sharp(svg(shapes[name], tones.green, 81)).png().toFile(path.join(tabDir, `${name}-active.png`)));
  }

  await Promise.all(jobs);
  console.log(JSON.stringify({ images: fs.readdirSync(imageDir).length, icons: fs.readdirSync(iconDir).length, tabIcons: fs.readdirSync(tabDir).length }));
}

build().catch((error) => {
  console.error(error);
  process.exit(1);
});
