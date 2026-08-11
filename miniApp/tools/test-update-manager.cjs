const assert = require("assert");
const path = require("path");

const modulePath = path.resolve(__dirname, "..", "utils", "update-manager.js");

function loadFreshModule() {
  delete require.cache[require.resolve(modulePath)];
  return require(modulePath);
}

function createUpdateManager() {
  const handlers = {};
  let applyCount = 0;
  const updateManager = {
    onUpdateReady(callback) {
      handlers.ready = callback;
    },
    onUpdateFailed(callback) {
      handlers.failed = callback;
    },
    applyUpdate() {
      applyCount += 1;
    },
  };

  return {
    handlers,
    updateManager,
    getApplyCount: () => applyCount,
  };
}

function testUnsupportedBaseLibrary() {
  const { registerUpdateManager } = loadFreshModule();
  assert.strictEqual(
    registerUpdateManager({}),
    null,
    "low base-library versions without getUpdateManager must degrade safely"
  );
}

function testReadyUpdateRestartsAfterConfirmation() {
  const { handlers, updateManager, getApplyCount } = createUpdateManager();
  const modalCalls = [];
  let getManagerCount = 0;
  const wxApi = {
    getUpdateManager() {
      getManagerCount += 1;
      return updateManager;
    },
    showModal(options) {
      modalCalls.push(options);
    },
  };
  const { registerUpdateManager } = loadFreshModule();

  assert.strictEqual(registerUpdateManager(wxApi), updateManager);
  assert.strictEqual(
    registerUpdateManager(wxApi),
    null,
    "update listeners must only be registered once"
  );
  assert.strictEqual(getManagerCount, 1);

  handlers.ready();
  handlers.ready();
  assert.strictEqual(modalCalls.length, 1, "ready callback must not show duplicate prompts");
  assert.strictEqual(modalCalls[0].title, "发现新版本");
  assert.strictEqual(modalCalls[0].confirmText, "重启更新");
  assert.strictEqual(modalCalls[0].showCancel, false);
  assert.strictEqual(getApplyCount(), 0);

  modalCalls[0].success({ confirm: true });
  assert.strictEqual(
    getApplyCount(),
    1,
    "confirmed ready update must restart through applyUpdate"
  );
}

function testReadyUpdateFallsBackToImmediateRestart() {
  const { handlers, updateManager, getApplyCount } = createUpdateManager();
  const { registerUpdateManager } = loadFreshModule();

  registerUpdateManager({
    getUpdateManager: () => updateManager,
  });
  handlers.ready();

  assert.strictEqual(
    getApplyCount(),
    1,
    "missing modal API must not block an already downloaded update"
  );
}

function testUpdateFailureExplainsRecovery() {
  const { handlers, updateManager, getApplyCount } = createUpdateManager();
  const modalCalls = [];
  const { registerUpdateManager } = loadFreshModule();

  registerUpdateManager({
    getUpdateManager: () => updateManager,
    showModal(options) {
      modalCalls.push(options);
    },
  });
  handlers.failed();
  handlers.failed();

  assert.strictEqual(modalCalls.length, 1, "download failure must be reported once");
  assert.strictEqual(modalCalls[0].title, "更新失败");
  assert.ok(modalCalls[0].content.includes("重新打开小程序"));
  assert.strictEqual(getApplyCount(), 0, "failed downloads must never call applyUpdate");
}

testUnsupportedBaseLibrary();
testReadyUpdateRestartsAfterConfirmation();
testReadyUpdateFallsBackToImmediateRestart();
testUpdateFailureExplainsRecovery();

console.log(JSON.stringify({
  unsupportedBaseLibrarySafe: true,
  listenersRegisteredOnce: true,
  readyUpdateRestarts: true,
  failedUpdateExplainsRecovery: true,
}, null, 2));
