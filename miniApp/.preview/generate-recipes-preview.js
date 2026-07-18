const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1240;

const recipes = [
  {
    name: "工作日晚餐",
    note: "快手、少油，30 分钟内",
    count: "4 道菜",
    images: [
      "miniApp/.preview/assets/recommended-green-pepper-beef.png",
      "miniApp/.preview/assets/recommended-tomato-egg.png",
      "miniApp/.preview/assets/recommended-garlic-broccoli.png",
    ],
  },
  {
    name: "周末家常菜",
    note: "一家人慢慢吃",
    count: "6 道菜",
    images: [
      "miniApp/.preview/assets/what-to-eat-kung-pao-chicken.png",
      "miniApp/.preview/assets/recommended-winter-melon-soup.png",
      "miniApp/.preview/assets/dish-detail-cover-realistic.png",
    ],
  },
  {
    name: "清淡早餐",
    note: "粥、鸡蛋和简单小菜",
    count: "3 道菜",
    images: [
      "miniApp/.preview/assets/recommended-tomato-egg.png",
      "miniApp/.preview/assets/recommended-garlic-broccoli.png",
      "miniApp/.preview/assets/recommended-winter-melon-soup.png",
    ],
  },
];

function defs() {
  return `
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/>
      <stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
  </defs>`;
}

function header() {
  return `
  <text x="44" y="39" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="700" fill="#151F1A">9:41</text>
  <g fill="#151F1A">
    <rect x="326" y="32" width="4" height="8" rx="1"/>
    <rect x="332" y="28" width="4" height="12" rx="1"/>
    <rect x="338" y="24" width="4" height="16" rx="1"/>
    <path d="M349 30c6-5 13-5 19 0M353 35c3-3 8-3 11 0" fill="none" stroke="#151F1A" stroke-width="2" stroke-linecap="round"/>
    <rect x="373" y="28" width="21" height="11" rx="3" fill="none" stroke="#151F1A" stroke-width="1.6"/>
    <rect x="395" y="31" width="2" height="5" rx="1"/>
    <rect x="376" y="31" width="15" height="5" rx="1"/>
  </g>
  <g>
    <rect x="312" y="60" width="82" height="36" rx="18" fill="#FFFFFF" fill-opacity="0.94" stroke="#E7E9E5"/>
    <circle cx="338" cy="76" r="2.4" fill="#151F1A"/>
    <circle cx="348" cy="76" r="2.4" fill="#151F1A"/>
    <line x1="359" y1="66" x2="359" y2="86" stroke="#E7E9E5"/>
    <circle cx="378" cy="76" r="9" fill="none" stroke="#151F1A" stroke-width="2.6"/>
  </g>
  <text x="32" y="98" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="28" font-weight="850" fill="#151F1A">菜谱</text>
  <text x="32" y="126" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">把常做的组合整理在一起</text>
  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>`;
}

function toolbar(empty) {
  return `
  <text x="24" y="276" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">我的菜谱</text>
  <text x="118" y="275" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">${empty ? "0 份" : "3 份"}</text>
  <rect x="268" y="244" width="126" height="42" rx="21" fill="url(#green)"/>
  <path d="M291 258v14M284 265h14" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round"/>
  <text x="347" y="271" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#FFFFFF">创建菜谱</text>`;
}

function nav() {
  return `
  <rect x="0" y="1064" width="418" height="176" fill="#FFFFFF"/>
  <line x1="0" y1="1064" x2="418" y2="1064" stroke="#E8EEE9"/>
  <g fill="none" stroke="#8A958F" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M67 1110l18-15 18 15v21H90v-13H80v13H67z"/>
    <path d="M318 1106c0-9 7-16 15-16s15 7 15 16-7 16-15 16-15-7-15-16zM309 1141c4-12 13-18 24-18s20 6 24 18"/>
  </g>
  <rect x="191" y="1087" width="36" height="36" rx="10" fill="#E9F8EF"/>
  <path d="M199 1097c7-3 12-1 10 5v13c-2-5-7-7-10-4zM219 1097c-7-3-12-1-10 5v13c2-5 7-7 10-4z" fill="none" stroke="#159B55" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="85" y="1160" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">首页</text>
  <text x="209" y="1160" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#159B55">菜谱</text>
  <text x="333" y="1160" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">我的</text>
  <rect x="147" y="1210" width="124" height="5" rx="2.5" fill="#202521" fill-opacity="0.74"/>`;
}

function populatedSvg() {
  const cards = recipes.map((recipe, index) => {
    const top = 316 + index * 232;
    return `
  <g filter="url(#shadow)"><rect x="24" y="${top}" width="370" height="208" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="${top + 40}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="800" fill="#1B241F">${recipe.name}</text>
  <text x="43" y="${top + 66}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">${recipe.note}</text>
  <text x="348" y="${top + 40}" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">${recipe.count}</text>
  <path d="M362 ${top + 30}l5.5 5.5-5.5 5.5" fill="none" stroke="#9AA49E" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>`;
  }).join("");

  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  ${defs()}
  ${header()}
  ${toolbar(false)}
  ${cards}
  ${nav()}
</svg>`;
}

function emptySvg() {
  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  ${defs()}
  ${header()}
  ${toolbar(true)}
  <g filter="url(#shadow)"><rect x="24" y="316" width="370" height="474" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <circle cx="209" cy="434" r="54" fill="#EDF7F1"/>
  <path d="M164 421c20-9 36-5 45 7v70c-8-18-26-23-45-14zM254 421c-20-9-36-5-45 7v70c8-18 26-23 45-14z" fill="#FFFFFF" stroke="#7FB795" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
  <path d="M209 428v70" fill="none" stroke="#7FB795" stroke-width="2"/>
  <text x="209" y="560" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="22" font-weight="850" fill="#1B241F">还没有菜谱</text>
  <text x="209" y="594" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#7B8982">把常做的一组菜整理起来</text>
  <text x="209" y="620" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#7B8982">下次选菜会更省心</text>
  <rect x="91" y="660" width="236" height="52" rx="26" fill="url(#green)"/>
  <path d="M126 679v14M119 686h14" fill="none" stroke="#FFFFFF" stroke-width="2" stroke-linecap="round"/>
  <text x="221" y="693" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">创建第一份菜谱</text>
  ${nav()}
</svg>`;
}

async function roundedPhoto(source, width, height, radius) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  return sharp(source)
    .resize({ width, height, fit: "cover", position: "center" })
    .composite([{ input: mask, blend: "dest-in" }])
    .png()
    .toBuffer();
}

async function heroCard() {
  const artwork = await sharp("miniApp/.preview/assets/recipes-cat-book.png")
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();
  return artwork;
}

async function main() {
  const hero = await heroCard();
  const photoLayers = [];
  for (let recipeIndex = 0; recipeIndex < recipes.length; recipeIndex += 1) {
    const top = 316 + recipeIndex * 232 + 94;
    for (let imageIndex = 0; imageIndex < recipes[recipeIndex].images.length; imageIndex += 1) {
      const photo = await roundedPhoto(recipes[recipeIndex].images[imageIndex], 98, 82, 12);
      photoLayers.push({ input: photo, left: 43 + imageIndex * 108, top });
    }
  }

  const populatedOutput = "miniApp/.preview/09-recipes-preview.png";
  const emptyOutput = "miniApp/.preview/09-recipes-empty-preview.png";
  const populatedBuffer = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#FEFEFA" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(populatedSvg()), left: 0, top: 0 }, ...photoLayers])
    .png()
    .toBuffer();
  const emptyBuffer = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#FEFEFA" } })
    .composite([{ input: hero, left: 0, top: 0 }, { input: Buffer.from(emptySvg()), left: 0, top: 0 }])
    .png()
    .toBuffer();

  await sharp(populatedBuffer).png().toFile(populatedOutput + ".tmp");
  fs.renameSync(populatedOutput + ".tmp", populatedOutput);
  await sharp(emptyBuffer).png().toFile(emptyOutput + ".tmp");
  fs.renameSync(emptyOutput + ".tmp", emptyOutput);
  const populatedMeta = await sharp(populatedOutput).metadata();
  const emptyMeta = await sharp(emptyOutput).metadata();
  console.log(JSON.stringify({
    outputs: [populatedOutput, emptyOutput],
    populated: { width: populatedMeta.width, height: populatedMeta.height },
    empty: { width: emptyMeta.width, height: emptyMeta.height },
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
