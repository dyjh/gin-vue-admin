const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1840;

function pageSvg() {
  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261a" flood-opacity="0.05"/>
    </filter>
    <linearGradient id="green" x1="0" x2="1">
      <stop offset="0" stop-color="#20B866"/>
      <stop offset="1" stop-color="#0A8E52"/>
    </linearGradient>
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
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="24" font-weight="800" fill="#151F1A">菜品详情</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">看做法，也能继续编辑</text>
  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>
  <g filter="url(#shadow)">
    <rect x="24" y="238" width="370" height="456" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <rect x="43" y="448" width="62" height="28" rx="14" fill="#E9F8EF"/>
  <text x="74" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">可用</text>
  <rect x="115" y="448" width="62" height="28" rx="14" fill="#F1F7F2"/>
  <text x="146" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#6E7A73">未公开</text>
  <text x="43" y="512" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="28" font-weight="850" fill="#151F1A">香菇鸡腿饭</text>
  <text x="43" y="543" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#7B8982">主食 · 家常菜 · 1份适合 2 人</text>
  <rect x="43" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="76" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">下饭</text>
  <rect x="122" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="155" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">快手</text>
  <rect x="201" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="234" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">少油</text>
  <line x1="43" y1="616" x2="373" y2="616" stroke="#EEF1EB"/>
  <text x="43" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">简介</text>
  <text x="94" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">鲜香下饭，适合工作日晚餐。</text>
  <text x="94" y="674" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#8A958F">少油也有香味，孩子版不放辣。</text>

  <g filter="url(#shadow)">
    <rect x="24" y="718" width="370" height="286" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="754" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">主要配料</text>
  <rect x="296" y="735" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="753" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">按1份</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" fill="#3A4740">
    <text x="52" y="806">鸡腿肉</text>
    <text x="327" y="806" text-anchor="end">220g</text>
    <line x1="43" y1="828" x2="373" y2="828" stroke="#EEF1EB"/>
    <text x="52" y="861">香菇</text>
    <text x="327" y="861" text-anchor="end">120g</text>
    <line x1="43" y1="883" x2="373" y2="883" stroke="#EEF1EB"/>
    <text x="52" y="916">胡萝卜</text>
    <text x="327" y="916" text-anchor="end">半根</text>
    <line x1="43" y1="938" x2="373" y2="938" stroke="#EEF1EB"/>
    <text x="52" y="946" dy="22">米饭</text>
    <text x="327" y="968" text-anchor="end">1碗</text>
  </g>

  <g filter="url(#shadow)">
    <rect x="24" y="1028" width="370" height="548" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="1064" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">做法步骤</text>
  <rect x="284" y="1045" width="82" height="26" rx="13" fill="#E9F8EF"/>
  <text x="325" y="1063" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">纯文字/图文</text>
  <text x="43" y="1090" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">步骤可只有文字，也可带步骤图</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">
    <circle cx="52" cy="1131" r="11" fill="#1F9D4E"/>
    <text x="52" y="1136" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">1</text>
    <text x="78" y="1130">鸡腿切块，香菇切片</text>
    <text x="78" y="1153" fill="#8A958F">先把食材处理好。</text>
    <line x1="43" y1="1328" x2="373" y2="1328" stroke="#EEF1EB"/>
    <circle cx="52" cy="1371" r="11" fill="#1F9D4E"/>
    <text x="52" y="1376" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">2</text>
    <text x="78" y="1370">少油煎香，加入米饭焖熟</text>
    <text x="78" y="1395" fill="#8A958F">中小火收汁，避免米饭太湿。</text>
    <line x1="43" y1="1422" x2="373" y2="1422" stroke="#EEF1EB"/>
    <circle cx="52" cy="1464" r="11" fill="#1F9D4E"/>
    <text x="52" y="1469" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">3</text>
    <text x="78" y="1463">出锅前撒少许葱花</text>
    <text x="78" y="1488" fill="#8A958F">拌匀后装盘，趁热吃。</text>
  </g>

  <g filter="url(#shadow)">
    <rect x="24" y="1600" width="370" height="112" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="1636" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">允许被发现</text>
  <text x="43" y="1670" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#7B8982">开启后可能进入推荐，被其他用户复制。</text>
  <rect x="306" y="1624" width="58" height="32" rx="16" fill="#E6ECE8"/>
  <circle cx="322" cy="1640" r="13" fill="#FFFFFF" stroke="#CED9D2"/>

  <rect x="0" y="1718" width="${WIDTH}" height="122" fill="#FEFEFA"/>
  <rect x="26" y="1732" width="132" height="50" rx="25" fill="#FFFFFF" stroke="#9FCFB2"/>
  <text x="92" y="1763" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#159B55">编辑</text>
  <rect x="174" y="1732" width="218" height="50" rx="25" fill="#FFF5F2" stroke="#F2BFAE"/>
  <text x="283" y="1763" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#D0603C">删除菜品</text>
  <text x="209" y="1803" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">菜谱在创建或编辑菜谱时选择菜品</text>
</svg>`;
}

async function roundedPhoto(source, width, height, radius, border) {
  const mask = Buffer.from(
    `<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`,
  );
  const frame = Buffer.from(
    `<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect x="0.5" y="0.5" width="${width - 1}" height="${height - 1}" rx="${radius}" fill="none" stroke="${border}"/></svg>`,
  );

  return sharp(source)
    .resize({ width, height, fit: "cover", position: "center" })
    .composite([
      { input: mask, blend: "dest-in" },
      { input: frame },
    ])
    .png()
    .toBuffer();
}

async function main() {
  const source = "miniApp/.preview/assets/dish-detail-cat-watercolor-v4.png";
  const output = "miniApp/.preview/04-dish_detail-preview.png";
  const coverPhoto = await roundedPhoto(
    "miniApp/.preview/assets/dish-detail-cover-realistic.png",
    332,
    170,
    18,
    "#E6EEE6",
  );
  const stepPhoto = await roundedPhoto(
    "miniApp/.preview/assets/dish-detail-step-1-realistic.png",
    286,
    132,
    16,
    "#E9E7D8",
  );
  const heroArtwork = await sharp(source)
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();

  const heroCard = await sharp({
    create: {
      width: 418,
      height: 224,
      channels: 4,
      background: "#FFF7E8",
    },
  })
    .composite([{ input: heroArtwork, left: 0, top: 0 }])
    .png()
    .toBuffer();

  const top = await sharp({
    create: {
      width: WIDTH,
      height: 224,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([{ input: heroCard, left: 0, top: 0 }])
    .png()
    .toBuffer();

  const buffer = await sharp({
    create: {
      width: WIDTH,
      height: HEIGHT,
      channels: 4,
      background: "#FEFEFA",
    },
  })
    .composite([
      { input: top, left: 0, top: 0 },
      { input: Buffer.from(pageSvg()), left: 0, top: 0 },
      { input: coverPhoto, left: 43, top: 258 },
      { input: stepPhoto, left: 78, top: 1170 },
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
