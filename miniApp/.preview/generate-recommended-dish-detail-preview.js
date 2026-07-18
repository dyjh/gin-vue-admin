const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 1860;

function pageSvg(joined, official) {
  const sourceBadges = official
    ? `
  <rect x="43" y="448" width="78" height="28" rx="14" fill="#E9F8EF"/>
  <text x="82" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">官方发布</text>
  <rect x="131" y="448" width="90" height="28" rx="14" fill="#F1F7F2"/>
  <text x="176" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#6E7A73">平台原创</text>`
    : `
  <rect x="43" y="448" width="78" height="28" rx="14" fill="#E9F8EF"/>
  <text x="82" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#159B55">平台精选</text>
  <rect x="131" y="448" width="90" height="28" rx="14" fill="#F1F7F2"/>
  <text x="176" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" font-weight="700" fill="#6E7A73">公开分享</text>`;

  const sourceCard = official
    ? `
  <text x="43" y="750" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#6E7A73">内容来源</text>
  <circle cx="66" cy="790" r="22" fill="#E9F8EF"/>
  <text x="66" y="797" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#159B55">官</text>
  <text x="103" y="786" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#1B241F">来干饭官方</text>
  <text x="103" y="809" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">平台原创发布</text>
  <rect x="292" y="774" width="78" height="28" rx="14" fill="#E9F8EF"/>
  <text x="331" y="793" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">官方发布</text>`
    : `
  <text x="43" y="750" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="800" fill="#6E7A73">内容来源</text>
  <circle cx="66" cy="790" r="22" fill="#E9F8EF"/>
  <text x="66" y="797" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="19" font-weight="800" fill="#159B55">林</text>
  <text x="103" y="786" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#1B241F">林一</text>
  <text x="103" y="809" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A958F">原作者公开分享</text>
  <rect x="292" y="774" width="78" height="28" rx="14" fill="#E9F8EF"/>
  <text x="331" y="793" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">平台精选</text>`;

  const bottomActions = joined
    ? `
  <rect x="26" y="1762" width="154" height="52" rx="26" fill="#E9F8EF" stroke="#B8DFC8"/>
  <path d="M55 1788l6 6 12-14" fill="none" stroke="#159B55" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="119" y="1795" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#159B55">已加入</text>
  <rect x="194" y="1762" width="198" height="52" rx="26" fill="url(#green)"/>
  <text x="293" y="1795" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">查看我的菜品</text>`
    : `
  <rect x="26" y="1762" width="366" height="52" rx="26" fill="url(#green)"/>
  <text x="209" y="1795" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#FFFFFF">加入菜品库</text>`;

  return `
<svg width="${WIDTH}" height="${HEIGHT}" viewBox="0 0 ${WIDTH} ${HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <filter id="shadow" x="-8%" y="-20%" width="116%" height="150%">
      <feDropShadow dx="0" dy="8" stdDeviation="14" flood-color="#10261A" flood-opacity="0.05"/>
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
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="23" font-weight="800" fill="#151F1A">推荐菜品详情</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="13" fill="#6F7D76">先看做法，再决定是否加入</text>

  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>

  <g filter="url(#shadow)">
    <rect x="24" y="238" width="370" height="456" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  ${sourceBadges}

  <text x="43" y="512" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="28" font-weight="850" fill="#151F1A">青椒牛柳</text>
  <text x="43" y="543" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#7B8982">主菜 · 家常菜 · 1份适合 2 人</text>
  <rect x="43" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="76" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">下饭</text>
  <rect x="122" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="155" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">快手</text>
  <rect x="201" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="234" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">少油</text>
  <line x1="43" y1="616" x2="373" y2="616" stroke="#EEF1EB"/>
  <text x="43" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">简介</text>
  <text x="94" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">牛肉滑嫩，青椒清脆，十分钟出锅。</text>
  <text x="94" y="674" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#8A958F">适合忙碌工作日，也很适合搭配米饭。</text>

  <g filter="url(#shadow)">
    <rect x="24" y="718" width="370" height="116" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  ${sourceCard}


  <g filter="url(#shadow)">
    <rect x="24" y="858" width="370" height="286" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="894" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">主要配料</text>
  <rect x="296" y="875" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="893" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">按1份</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" fill="#3A4740">
    <text x="52" y="946">牛里脊</text>
    <text x="327" y="946" text-anchor="end">250g</text>
    <line x1="43" y1="968" x2="373" y2="968" stroke="#EEF1EB"/>
    <text x="52" y="1001">青椒</text>
    <text x="327" y="1001" text-anchor="end">2个</text>
    <line x1="43" y1="1023" x2="373" y2="1023" stroke="#EEF1EB"/>
    <text x="52" y="1056">洋葱</text>
    <text x="327" y="1056" text-anchor="end">1/4个</text>
    <line x1="43" y1="1078" x2="373" y2="1078" stroke="#EEF1EB"/>
    <text x="52" y="1111">生抽</text>
    <text x="327" y="1111" text-anchor="end">1勺</text>
  </g>

  <g filter="url(#shadow)">
    <rect x="24" y="1168" width="370" height="472" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/>
  </g>
  <text x="43" y="1204" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">做法步骤</text>
  <rect x="296" y="1185" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="1203" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">图文步骤</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">
    <circle cx="52" cy="1260" r="11" fill="#1F9D4E"/>
    <text x="52" y="1265" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">1</text>
    <text x="78" y="1259">牛肉切片，加生抽和淀粉抓匀</text>
    <text x="78" y="1282" fill="#8A958F">静置十分钟，青椒切成细条。</text>
    <line x1="43" y1="1312" x2="373" y2="1312" stroke="#EEF1EB"/>
    <circle cx="52" cy="1354" r="11" fill="#1F9D4E"/>
    <text x="52" y="1359" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">2</text>
    <text x="78" y="1353">热锅少油，牛肉与青椒快速翻炒</text>
    <text x="78" y="1376" fill="#8A958F">牛肉变色后加入青椒，保持大火。</text>
    <line x1="43" y1="1542" x2="373" y2="1542" stroke="#EEF1EB"/>
    <circle cx="52" cy="1582" r="11" fill="#1F9D4E"/>
    <text x="52" y="1587" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">3</text>
    <text x="78" y="1581">沿锅边加少许生抽，翻匀出锅</text>
    <text x="78" y="1604" fill="#8A958F">尝味后再补盐，避免口味过重。</text>
  </g>

  <g filter="url(#shadow)">
    <rect x="24" y="1664" width="370" height="78" rx="16" fill="#F2F8F3" stroke="#E0EEE3"/>
  </g>
  <circle cx="54" cy="1703" r="18" fill="#DFF3E6"/>
  <rect x="48" y="1699" width="12" height="10" rx="2" fill="none" stroke="#159B55" stroke-width="1.7"/>
  <path d="M50 1699v-4a4 4 0 0 1 8 0v4" fill="none" stroke="#159B55" stroke-width="1.7"/>
  <text x="82" y="1698" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#315644">加入后仅自己可见</text>
  <text x="82" y="1721" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#728078">可继续编辑使用，但无法再次公开</text>

  <rect x="0" y="1748" width="${WIDTH}" height="112" fill="#FEFEFA"/>
  ${bottomActions}
  <text x="209" y="1838" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">副本会保留原作者与推荐来源</text>
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

async function render(output, joined, official, heroCard, coverPhoto, stepPhoto) {
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
      { input: Buffer.from(pageSvg(joined, official)), left: 0, top: 0 },
      { input: coverPhoto, left: 43, top: 258 },
      { input: stepPhoto, left: 78, top: 1392 },
    ])
    .png()
    .toBuffer();

  await sharp(buffer).png().toFile(output + ".tmp");
  fs.renameSync(output + ".tmp", output);
  return sharp(output).metadata();
}

async function main() {
  const heroArtwork = await sharp(
    "miniApp/.preview/assets/recommended-dish-detail-cat-card.png",
  )
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();
  const heroCard = await sharp({
    create: {
      width: 418,
      height: 224,
      channels: 4,
      background: "#FDF6ED",
    },
  })
    .composite([{ input: heroArtwork, left: 0, top: 0 }])
    .png()
    .toBuffer();

  const coverPhoto = await roundedPhoto(
    "miniApp/.preview/assets/recommended-green-pepper-beef.png",
    332,
    170,
    18,
    "#E6EEE6",
  );
  const stepPhoto = await roundedPhoto(
    "miniApp/.preview/assets/recommended-green-pepper-beef-step.png",
    286,
    132,
    16,
    "#E9E7D8",
  );

  const defaultOutput = "miniApp/.preview/06-recommended_dish_detail-preview.png";
  const officialOutput = "miniApp/.preview/06-recommended_dish_detail-official-preview.png";
  const addedOutput = "miniApp/.preview/06-recommended_dish_detail-added-preview.png";
  const defaultMeta = await render(defaultOutput, false, false, heroCard, coverPhoto, stepPhoto);
  const officialMeta = await render(officialOutput, false, true, heroCard, coverPhoto, stepPhoto);
  const addedMeta = await render(addedOutput, true, false, heroCard, coverPhoto, stepPhoto);
  console.log(
    JSON.stringify(
      {
        outputs: [defaultOutput, officialOutput, addedOutput],
        width: defaultMeta.width,
        height: defaultMeta.height,
        officialWidth: officialMeta.width,
        officialHeight: officialMeta.height,
        addedWidth: addedMeta.width,
        addedHeight: addedMeta.height,
      },
      null,
      2,
    ),
  );
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
