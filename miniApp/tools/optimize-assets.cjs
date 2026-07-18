const fs = require("fs");
const path = require("path");
const sharp = require("sharp");

const root = path.resolve(__dirname, "..");
const sourceDir = path.join(root, ".preview", "assets");
const outputDir = path.join(root, "assets", "images");

if (!outputDir.startsWith(root + path.sep)) {
  throw new Error("Refusing to modify an image directory outside miniApp.");
}

const used = [
  "dish-detail-cover-realistic.png",
  "dish-detail-step-1-realistic.png",
  "recommended-tomato-egg.png",
  "recommended-garlic-broccoli.png",
  "recommended-winter-melon-soup.png",
  "recommended-green-pepper-beef.png",
  "recommended-green-pepper-beef-step.png",
  "what-to-eat-kung-pao-chicken.png",
  "ai-kung-pao-chicken-step.png",
  "dish-detail-cat-watercolor-v4.png",
  "recommended-dish-detail-cat-card.png",
  "recommended-dishes-cat-flag.png",
  "what-to-eat-cat-fridge.png",
  "ai-recommendation-detail-cat-check.png",
  "recipes-cat-book.png",
  "recipe-detail-cat-bookmark.png",
  "profile-cat-washing-v3.png",
  "checkin-cat-camera-v1.png",
  "meal-create-cat-invite-v1.png",
  "meal-invite-cat-code-v3.png",
  "profile-cat-spoon-v2.png",
  "meal-stats-cat-checklist-v1.png",
  "shopping-list-cat-basket-v1.png",
  "prep-ai-cat-timer-v1.png",
  "points-cat-jar-v1.png",
  "notifications-cat-envelope-v1.png"
];

const photoNames = new Set([
  "dish-detail-cover-realistic.png",
  "recommended-tomato-egg.png",
  "recommended-garlic-broccoli.png",
  "recommended-winter-melon-soup.png",
  "recommended-green-pepper-beef.png",
  "what-to-eat-kung-pao-chicken.png"
]);

const stepNames = new Set([
  "dish-detail-step-1-realistic.png",
  "recommended-green-pepper-beef-step.png",
  "ai-kung-pao-chicken-step.png"
]);

function walk(directory, output = []) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    if ([".preview", "design", "assets", "tools"].includes(entry.name)) continue;
    const full = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(full, output);
    else if (/\.(js|json|wxml|wxss)$/.test(entry.name)) output.push(full);
  }
  return output;
}

async function optimize(name) {
  const source = path.join(sourceDir, name);
  if (!fs.existsSync(source)) throw new Error(`Missing source image: ${name}`);
  const target = path.join(outputDir, name.replace(/\.png$/i, ".jpg"));
  let pipeline = sharp(source).flatten({ background: "#FDF6ED" });
  if (photoNames.has(name)) {
    pipeline = pipeline.resize(600, 420, { fit: "cover", position: "centre" });
  } else if (stepNames.has(name)) {
    pipeline = pipeline.resize(720, 430, { fit: "cover", position: "centre" });
  } else {
    pipeline = pipeline.resize(750, 448, { fit: "cover", position: "centre" });
  }
  await pipeline.jpeg({ quality: 72, progressive: true, mozjpeg: true }).toFile(target);
}

async function main() {
  fs.mkdirSync(outputDir, { recursive: true });
  await Promise.all(used.map(optimize));

  for (const file of fs.readdirSync(outputDir)) {
    if (file.toLowerCase().endsWith(".png")) {
      fs.unlinkSync(path.join(outputDir, file));
    }
  }

  for (const file of walk(root)) {
    let source = fs.readFileSync(file, "utf8");
    let changed = false;
    for (const name of used) {
      const before = `/assets/images/${name}`;
      const after = before.replace(/\.png$/i, ".jpg");
      if (source.includes(before)) {
        source = source.split(before).join(after);
        changed = true;
      }
    }
    if (changed) fs.writeFileSync(file, source, "utf8");
  }

  const files = fs.readdirSync(outputDir).map((name) => path.join(outputDir, name));
  const bytes = files.reduce((sum, file) => sum + fs.statSync(file).size, 0);
  console.log(JSON.stringify({ images: files.length, bytes, megabytes: Number((bytes / 1024 / 1024).toFixed(2)) }));
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
