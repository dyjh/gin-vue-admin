const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..");
const api = JSON.parse(fs.readFileSync(path.join(root, "api.json"), "utf8"));
const source = fs.readFileSync(path.join(root, "services", "api.js"), "utf8");

function normalize(value) {
  return value
    .replace(/\$\{[^}]+\}/g, "{}")
    .replace(/\{[^}]+\}/g, "{}");
}

const contract = new Set(api.interfaces.map((item) => `${item.method} ${normalize(item.path)}`));
const used = [];
const pattern = /request\(\{\s*url:\s*([\`"])(.*?)\1(?:,\s*method:\s*"([A-Z]+)")?/gs;
let match;
while ((match = pattern.exec(source))) {
  used.push(`${match[3] || "GET"} ${normalize(match[2])}`);
}

const missing = used.filter((signature) => !contract.has(signature));
if (missing.length) {
  console.error(`Client request paths missing from api.json:\n${missing.join("\n")}`);
  process.exit(1);
}

console.log(JSON.stringify({ clientRequests: used.length, matchedContracts: used.length }));
