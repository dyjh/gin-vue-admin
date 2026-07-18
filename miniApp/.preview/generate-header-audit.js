const sharp = require("sharp");

const files = [
  "home-preview.png",
  "02-dish_add_entry-entry-preview.png",
  "03-dish_edit-preview.png",
  "04-dish_detail-preview.png",
  "05-recommended_dishes-preview.png",
  "06-recommended_dish_detail-preview.png",
  "07-what_to_eat-input-preview.png",
  "08-ai_recommendation_detail-preview.png",
  "09-recipes-preview.png",
  "10-recipe_detail-preview.png",
  "11-profile-preview.png",
];

async function main() {
  const layers = [];
  for (let index = 0; index < files.length; index += 1) {
    const input = await sharp(`miniApp/.preview/${files[index]}`)
      .extract({ left: 0, top: 0, width: 418, height: 224 })
      .resize(209, 112)
      .png()
      .toBuffer();
    const column = index % 3;
    const row = Math.floor(index / 3);
    const left = 8 + column * 220;
    const top = 8 + row * 150;
    layers.push({ input, left, top });
    layers.push({
      input: Buffer.from(`<svg width="209" height="24" xmlns="http://www.w3.org/2000/svg"><text x="0" y="17" font-family="Arial, sans-serif" font-size="12" font-weight="700" fill="#1B241F">${String(index + 1).padStart(2, "0")} ${files[index]}</text></svg>`),
      left,
      top: top + 116,
    });
  }

  await sharp({ create: { width: 668, height: 608, channels: 4, background: "#F2F4F2" } })
    .composite(layers)
    .png()
    .toFile("miniApp/.preview/header-fullbleed-audit.png");
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
