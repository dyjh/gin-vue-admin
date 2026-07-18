const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1168;

const dishes = [
  {
    name: "青椒牛柳",
    meta: "下饭 · 快手",
    source: "平台精选",
    image: "miniApp/.preview/assets/recommended-green-pepper-beef.png",
  },
  {
    name: "番茄炒蛋",
    meta: "家常 · 适合孩子",
    source: "平台精选",
    image: "miniApp/.preview/assets/recommended-tomato-egg.png",
  },
  {
    name: "冬瓜丸子汤",
    meta: "清淡 · 汤菜",
    source: "平台精选",
    image: "miniApp/.preview/assets/recommended-winter-melon-soup.png",
  },
  {
    name: "蒜蓉西兰花",
    meta: "素菜 · 少油",
    source: "平台精选",
    image: "miniApp/.preview/assets/recommended-garlic-broccoli.png",
  },
  {
    name: "香菇鸡腿饭",
    meta: "主食 · 一锅出",
    source: "平台精选",
    image: "miniApp/.preview/assets/dish-detail-cover-realistic.png",
  },
];

function pageSvg() {
  const rows = dishes
    .map((dish, index) => {
      const top = 462 + index * 126;
      const separator =
        index < dishes.length - 1
          ? `<line x1="38" y1="${top + 126}" x2="380" y2="${top + 126}" stroke="#EDF1EC"/>`
          : "";

      return `
      <text x="140" y="${top + 40}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="800" fill="#17231D">${dish.name}</text>
      <text x="140" y="${top + 66}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#7B8982">${dish.meta}</text>
      <circle cx="146" cy="${top + 90}" r="5" fill="#DFF3E6"/>
      <path d="M143 ${top + 91}c3-5 7-5 9-6c-1 5-4 9-9 9c1-1 2-2 4-3" fill="none" stroke="#159B55" stroke-width="1.3" stroke-linecap="round"/>
      <text x="158" y="${top + 94}" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#87948D">${dish.source}</text>
      <text x="354" y="${top + 72}" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">查看详情</text>
      <path d="M366 ${top + 62}l6 6-6 6" fill="none" stroke="#159B55" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
      ${separator}`;
    })
    .join("");

  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.05"/>
    </filter>
  </defs>

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
  <path d="M48 79L38 89l10 10M39 89h20" fill="none" stroke="#151F1A" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">菜品推荐</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">精选好菜，先看看再决定</text>

  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>

  <rect x="24" y="238" width="370" height="46" rx="19" fill="#FFFFFF" stroke="#E5ECE6"/>
  <circle cx="48" cy="261" r="7" fill="none" stroke="#8A958F" stroke-width="1.8"/>
  <path d="M53 266l5 5" fill="none" stroke="#8A958F" stroke-width="1.8" stroke-linecap="round"/>
  <text x="68" y="267" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#9AA49E">搜索推荐菜品</text>

  <rect x="24" y="302" width="54" height="34" rx="17" fill="#159B55"/>
  <text x="51" y="324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="800" fill="#FFFFFF">全部</text>
  <rect x="88" y="302" width="62" height="34" rx="17" fill="#F1F6F2"/>
  <text x="119" y="324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#536159">快手</text>
  <rect x="160" y="302" width="62" height="34" rx="17" fill="#F1F6F2"/>
  <text x="191" y="324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#536159">清淡</text>
  <rect x="232" y="302" width="62" height="34" rx="17" fill="#F1F6F2"/>
  <text x="263" y="324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#536159">汤菜</text>
  <rect x="304" y="302" width="90" height="34" rx="17" fill="#F1F6F2"/>
  <text x="349" y="324" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#536159">适合孩子</text>

  <rect x="24" y="354" width="370" height="54" rx="14" fill="#EEF8F1"/>
  <circle cx="48" cy="381" r="12" fill="#DDF2E5"/>
  <path d="M43 381h10v8H43zM45 381v-3a3 3 0 0 1 6 0v3" fill="none" stroke="#159B55" stroke-width="1.6" stroke-linejoin="round"/>
  <text x="68" y="377" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#355846">平台精选，保留分享来源</text>
  <text x="68" y="397" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#738079">来自公开分享的家常好菜</text>

  <text x="24" y="444" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">精选菜品</text>
  <text x="394" y="443" text-anchor="end" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">共 32 道</text>

  <g filter="url(#shadow)">
    <rect x="24" y="462" width="370" height="630" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  ${rows}

  <text x="209" y="1134" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#9AA49E">继续下滑查看更多精选菜品</text>
</svg>`;
}

async function roundedPhoto(source, width, height, radius) {
  const mask = Buffer.from(
    `<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`,
  );

  return sharp(source)
    .resize({ width, height, fit: "cover", position: "center" })
    .composite([{ input: mask, blend: "dest-in" }])
    .png()
    .toBuffer();
}

async function main() {
  const heroSource = "miniApp/.preview/assets/recommended-dishes-cat-flag.png";
  const output = "miniApp/.preview/05-recommended_dishes-preview.png";

  const heroArtwork = await sharp(heroSource)
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();
  const heroCard = heroArtwork;

  const photos = await Promise.all(
    dishes.map((dish) => roundedPhoto(dish.image, 88, 88, 14)),
  );
  const photoLayers = photos.map((photo, index) => ({
    input: photo,
    left: 38,
    top: 476 + index * 126,
  }));

  const buffer = await sharp({
    create: {
      width: WIDTH,
      height: HEIGHT,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([
      { input: heroCard, left: 0, top: 0 },
      { input: Buffer.from(pageSvg()), left: 0, top: 0 },
      ...photoLayers,
    ])
    .png()
    .toBuffer();

  await sharp(buffer).png().toFile(output + ".tmp");
  fs.renameSync(output + ".tmp", output);
  const meta = await sharp(output).metadata();
  console.log(JSON.stringify({ output, width: meta.width, height: meta.height }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
