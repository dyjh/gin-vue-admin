const fs = require("fs");
const sharp = require("sharp");

const WIDTH = 418;
const HEIGHT = 2200;

function pageSvg(saved) {
  const actions = saved
    ? `
  <rect x="26" y="2084" width="154" height="52" rx="26" fill="#E9F8EF" stroke="#B8DFC8"/>
  <path d="M55 2110l6 6 12-14" fill="none" stroke="#159B55" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="119" y="2117" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#159B55">已保存</text>
  <rect x="194" y="2084" width="198" height="52" rx="26" fill="url(#green)"/>
  <text x="293" y="2117" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="16" font-weight="800" fill="#FFFFFF">查看我的菜品</text>`
    : `
  <rect x="26" y="2084" width="366" height="52" rx="26" fill="url(#green)"/>
  <text x="209" y="2117" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="17" font-weight="800" fill="#FFFFFF">保存到菜品库</text>`;

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
  <text x="70" y="96" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="23" font-weight="800" fill="#151F1A">AI推荐详情</text>
  <text x="70" y="124" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#6F7D76">看完做法，再决定是否保存</text>

  <rect y="224" width="${WIDTH}" height="${HEIGHT - 224}" fill="#FEFEFA"/>

  <g filter="url(#shadow)"><rect x="24" y="238" width="370" height="500" rx="18" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <rect x="43" y="448" width="86" height="28" rx="14" fill="#E9F8EF"/>
  <text x="86" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">AI生成推荐</text>
  <rect x="139" y="448" width="78" height="28" rx="14" fill="#FFF3E8"/>
  <text x="178" y="467" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#C87932">经典川菜</text>
  <text x="43" y="512" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="28" font-weight="850" fill="#151F1A">宫保鸡丁</text>
  <text x="43" y="543" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#7B8982">川菜 · 经典家常菜 · 1份适合 2 人</text>
  <rect x="43" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="76" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">下饭</text>
  <rect x="122" y="562" width="66" height="30" rx="15" fill="#E9F8EF"/>
  <text x="155" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#159B55">快手</text>
  <rect x="201" y="562" width="66" height="30" rx="15" fill="#FFF1EA"/>
  <text x="234" y="582" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" font-weight="700" fill="#D06F3D">微辣</text>
  <line x1="43" y1="616" x2="373" y2="616" stroke="#EEF1EB"/>
  <text x="43" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#1B241F">简介</text>
  <text x="94" y="649" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#3A4740">鸡肉滑嫩、花生酥香，酸甜中带一点辣。</text>
  <text x="94" y="674" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#8A958F">经典川味家常菜，辣度可按偏好降低。</text>
  <circle cx="53" cy="708" r="10" fill="#E9F8EF"/>
  <path d="M48 708l3 3 6-7" fill="none" stroke="#159B55" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
  <text x="72" y="713" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#63736B">标准菜名 · 川菜</text>

  <g filter="url(#shadow)"><rect x="24" y="762" width="370" height="140" rx="16" fill="#EDF7F8" stroke="#D8EBEE"/></g>
  <text x="43" y="798" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="18" font-weight="800" fill="#31525A">内容校验</text>
  <rect x="43" y="818" width="92" height="30" rx="15" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="89" y="838" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#4B737C">菜名与菜系</text>
  <rect x="145" y="818" width="92" height="30" rx="15" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="191" y="838" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#4B737C">标准食材</text>
  <rect x="247" y="818" width="92" height="30" rx="15" fill="#FFFFFF" fill-opacity="0.82"/>
  <text x="293" y="838" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#4B737C">步骤一致</text>
  <text x="43" y="874" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#698087">校验未通过的 AI 结果不会展示</text>

  <g filter="url(#shadow)"><rect x="24" y="926" width="370" height="382" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="962" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">主要配料</text>
  <rect x="296" y="943" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="961" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">按1份</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" fill="#3A4740">
    <text x="52" y="1014">鸡腿肉</text><text x="352" y="1014" text-anchor="end">250g</text>
    <line x1="43" y1="1036" x2="373" y2="1036" stroke="#EEF1EB"/>
    <text x="52" y="1069">花生米</text><text x="352" y="1069" text-anchor="end">50g</text>
    <line x1="43" y1="1091" x2="373" y2="1091" stroke="#EEF1EB"/>
    <text x="52" y="1124">干辣椒</text><text x="352" y="1124" text-anchor="end">8g</text>
    <line x1="43" y1="1146" x2="373" y2="1146" stroke="#EEF1EB"/>
    <text x="52" y="1179">大葱</text><text x="352" y="1179" text-anchor="end">40g</text>
    <line x1="43" y1="1201" x2="373" y2="1201" stroke="#EEF1EB"/>
    <text x="52" y="1234">腌料</text><text x="352" y="1234" text-anchor="end">生抽1勺 · 淀粉1勺</text>
    <line x1="43" y1="1256" x2="373" y2="1256" stroke="#EEF1EB"/>
    <text x="52" y="1289">宫保汁</text><text x="352" y="1289" text-anchor="end">醋1勺 · 糖1勺 · 生抽1勺</text>
  </g>

  <g filter="url(#shadow)"><rect x="24" y="1332" width="370" height="610" rx="16" fill="#FFFFFF" stroke="#EEF1EB"/></g>
  <text x="43" y="1368" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="21" font-weight="800" fill="#1B241F">做法步骤</text>
  <rect x="296" y="1349" width="70" height="26" rx="13" fill="#E9F8EF"/>
  <text x="331" y="1367" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" font-weight="700" fill="#159B55">图文步骤</text>
  <g font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="14" fill="#3A4740">
    <circle cx="52" cy="1424" r="11" fill="#1F9D4E"/><text x="52" y="1429" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">1</text>
    <text x="78" y="1423">鸡腿肉切丁，加生抽和淀粉抓匀</text>
    <text x="78" y="1446" fill="#8A958F">静置十分钟，让鸡肉保持滑嫩。</text>
    <line x1="43" y1="1474" x2="373" y2="1474" stroke="#EEF1EB"/>
    <circle cx="52" cy="1512" r="11" fill="#1F9D4E"/><text x="52" y="1517" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">2</text>
    <text x="78" y="1511">醋、糖、生抽和少量清水调成宫保汁</text>
    <text x="78" y="1534" fill="#8A958F">提前调匀，入锅后更容易控制火候。</text>
    <line x1="43" y1="1562" x2="373" y2="1562" stroke="#EEF1EB"/>
    <circle cx="52" cy="1600" r="11" fill="#1F9D4E"/><text x="52" y="1605" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">3</text>
    <text x="78" y="1599">热锅少油，炒香干辣椒后下鸡丁</text>
    <text x="78" y="1622" fill="#8A958F">鸡丁变色后加入葱段，保持中大火。</text>
    <line x1="43" y1="1782" x2="373" y2="1782" stroke="#EEF1EB"/>
    <circle cx="52" cy="1820" r="11" fill="#1F9D4E"/><text x="52" y="1825" text-anchor="middle" fill="#FFFFFF" font-size="12" font-weight="700">4</text>
    <text x="78" y="1819">倒入宫保汁翻匀，最后加入花生米</text>
    <text x="78" y="1842" fill="#8A958F">汤汁包裹鸡丁后立即出锅，保持花生酥脆。</text>
  </g>
  <rect x="43" y="1880" width="330" height="40" rx="12" fill="#FFF8F1"/>
  <text x="208" y="1905" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#8A6A4B">口味可调整，但不要替换标准主料和核心做法</text>

  <g filter="url(#shadow)"><rect x="24" y="1966" width="370" height="86" rx="16" fill="#F2F8F3" stroke="#E0EEE3"/></g>
  <circle cx="54" cy="2009" r="18" fill="#DFF3E6"/>
  <rect x="48" y="2005" width="12" height="10" rx="2" fill="none" stroke="#159B55" stroke-width="1.7"/>
  <path d="M50 2005v-4a4 4 0 0 1 8 0v4" fill="none" stroke="#159B55" stroke-width="1.7"/>
  <text x="82" y="2003" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="15" font-weight="800" fill="#315644">保存后进入你的菜品库</text>
  <text x="82" y="2027" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#728078">默认未公开，可继续调整份量和做法</text>

  <rect x="0" y="2066" width="${WIDTH}" height="134" fill="#FEFEFA"/>
  ${actions}
  <text x="209" y="2168" text-anchor="middle" font-family="Microsoft YaHei, PingFang SC, sans-serif" font-size="12" fill="#7B8982">仅保存后记录为采用 · 仅查看不会创建菜品</text>
</svg>`;
}

async function roundedPhoto(source, width, height, radius, border) {
  const mask = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect width="${width}" height="${height}" rx="${radius}" fill="#fff"/></svg>`);
  const frame = Buffer.from(`<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg"><rect x="0.5" y="0.5" width="${width - 1}" height="${height - 1}" rx="${radius}" fill="none" stroke="${border}"/></svg>`);
  return sharp(source)
    .resize({ width, height, fit: "cover", position: "center" })
    .composite([{ input: mask, blend: "dest-in" }, { input: frame }])
    .png()
    .toBuffer();
}

async function render(output, saved, hero, cover, step) {
  const buffer = await sharp({ create: { width: WIDTH, height: HEIGHT, channels: 4, background: "#FEFEFA" } })
    .composite([
      { input: hero, left: 0, top: 0 },
      { input: Buffer.from(pageSvg(saved)), left: 0, top: 0 },
      { input: cover, left: 43, top: 258 },
      { input: step, left: 78, top: 1638 },
    ])
    .png()
    .toBuffer();
  await sharp(buffer).png().toFile(output + ".tmp");
  fs.renameSync(output + ".tmp", output);
  return sharp(output).metadata();
}

async function main() {
  const heroArtwork = await sharp("miniApp/.preview/assets/ai-recommendation-detail-cat-check.png")
    .resize({ width: 418, height: 224, fit: "cover", position: "center" })
    .png()
    .toBuffer();
  const hero = heroArtwork;
  const cover = await roundedPhoto("miniApp/.preview/assets/what-to-eat-kung-pao-chicken.png", 332, 170, 18, "#E6EEE6");
  const step = await roundedPhoto("miniApp/.preview/assets/ai-kung-pao-chicken-step.png", 286, 132, 16, "#E9E7D8");

  const viewOutput = "miniApp/.preview/08-ai_recommendation_detail-preview.png";
  const savedOutput = "miniApp/.preview/08-ai_recommendation_detail-saved-preview.png";
  const viewMeta = await render(viewOutput, false, hero, cover, step);
  const savedMeta = await render(savedOutput, true, hero, cover, step);
  console.log(JSON.stringify({
    outputs: [viewOutput, savedOutput],
    view: { width: viewMeta.width, height: viewMeta.height },
    saved: { width: savedMeta.width, height: savedMeta.height },
  }, null, 2));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
