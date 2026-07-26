const fs = require("fs");
const path = require("path");

const root = path.resolve(__dirname, "..", "..", "..");
const contractPath = path.join(root, "aiDoc", "miniApp", "api.json");
const swaggerPath = path.join(root, "server", "docs", "swagger.json");
const contract = JSON.parse(fs.readFileSync(contractPath, "utf8"));
const swagger = JSON.parse(fs.readFileSync(swaggerPath, "utf8"));
const failures = [];

function fail(message) {
  failures.push(message);
}

function signature(method, endpointPath) {
  return `${String(method).toUpperCase()} ${endpointPath}`;
}

const contractEndpoints = new Set(
  (contract.interfaces || []).map((item) => (
    signature(item.method, `/miniapp/v1${item.path}`)
  )),
);
const swaggerEndpoints = new Set();
for (const [endpointPath, operations] of Object.entries(swagger.paths || {})) {
  if (!endpointPath.startsWith("/miniapp/v1")) continue;
  for (const method of Object.keys(operations || {})) {
    if (!["get", "post", "put", "patch", "delete"].includes(method)) continue;
    swaggerEndpoints.add(signature(method, endpointPath));
  }
}
for (const endpoint of contractEndpoints) {
  if (!swaggerEndpoints.has(endpoint)) {
    fail(`Swagger is missing mini-program endpoint ${endpoint}`);
  }
}
for (const endpoint of swaggerEndpoints) {
  if (!contractEndpoints.has(endpoint)) {
    fail(`Swagger contains undeclared mini-program endpoint ${endpoint}`);
  }
}

function contractModelFields(modelName, seen = new Set()) {
  if (seen.has(modelName)) return [];
  seen.add(modelName);
  const model = contract.models?.[modelName];
  if (!model || typeof model !== "object" || Array.isArray(model)) return [];
  const fields = Object.keys(model).filter((key) => key !== "allOf");
  for (const parent of model.allOf || []) {
    fields.push(...contractModelFields(parent, seen));
  }
  return [...new Set(fields)].sort();
}

function swaggerDefinitionName(modelName) {
  const preferred = [`response.${modelName}`, `request.${modelName}`];
  for (const candidate of preferred) {
    if (swagger.definitions?.[candidate]) return candidate;
  }
  return Object.keys(swagger.definitions || {}).find((name) => name.endsWith(`.${modelName}`));
}

let comparedModels = 0;
for (const modelName of Object.keys(contract.models || {})) {
  const definitionName = swaggerDefinitionName(modelName);
  if (!definitionName) continue;
  comparedModels += 1;
  const contractFields = contractModelFields(modelName);
  const swaggerFields = Object.keys(
    swagger.definitions[definitionName].properties || {},
  ).sort();
  const missing = contractFields.filter((field) => !swaggerFields.includes(field));
  const extra = swaggerFields.filter((field) => !contractFields.includes(field));
  if (missing.length) {
    fail(`${definitionName} is missing contract fields: ${missing.join(", ")}`);
  }
  if (extra.length) {
    fail(`${definitionName} contains undeclared fields: ${extra.join(", ")}`);
  }
}

if (failures.length) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(JSON.stringify({
  contractInterfaces: contractEndpoints.size,
  swaggerInterfaces: swaggerEndpoints.size,
  comparedModels,
}, null, 2));
