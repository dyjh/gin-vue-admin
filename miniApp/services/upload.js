const env = require("../config/env");
const auth = require("./auth");
const {
  createIdempotencyKey,
  createNetworkError,
  createResponseError,
  isAuthExpiredError,
  showErrorToast,
} = require("./request-helpers");

function isRemoteOrBundled(filePath) {
  return !filePath || filePath.startsWith("/assets/") || /^https?:\/\//.test(filePath);
}

function sendUpload(filePath, scene, idempotencyKey) {
  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: `${env.baseUrl}/uploads/images`,
      filePath,
      name: "file",
      formData: { scene },
      timeout: env.timeout,
      header: {
        Authorization: `Bearer ${auth.getAccessToken()}`,
        "X-Idempotency-Key": idempotencyKey,
      },
      success(response) {
        let body;
        try {
          body = JSON.parse(response.data);
        } catch (originalError) {
          const error = new Error("图片上传响应格式错误");
          error.code = "INVALID_UPLOAD_RESPONSE";
          error.cause = originalError;
          reject(error);
          return;
        }
        const normalizedResponse = {
          ...response,
          data: body,
        };
        if (response.statusCode >= 200 && response.statusCode < 300 && body.code === 0) {
          resolve(body.data);
          return;
        }
        reject(createResponseError(normalizedResponse, "图片上传失败"));
      },
      fail(originalError) {
        reject(createNetworkError(originalError, "图片上传失败，请稍后重试"));
      },
    });
  });
}

async function uploadImage(filePath, scene, options = {}) {
  if (isRemoteOrBundled(filePath)) {
    return { fileId: filePath || "", url: filePath || "" };
  }

  const idempotencyKey = options.idempotencyKey || createIdempotencyKey("upload");
  try {
    await auth.ensureAuthenticated();
    return await sendUpload(filePath, scene, idempotencyKey);
  } catch (error) {
    if (isAuthExpiredError(error)) {
      try {
        await auth.reauthenticate();
      } catch (authenticationError) {
        if (options.showError !== false) {
          showErrorToast(authenticationError, "登录失败，请稍后重试");
        }
        throw authenticationError;
      }
      error.reauthenticated = true;
      error.retryRequired = true;
      error.message = "登录已恢复，请重新选择图片上传";
    }
    if (options.showError !== false) showErrorToast(error, "图片上传失败，请稍后重试");
    throw error;
  }
}

async function normalizeDishPayload(form) {
  const cover = form.coverFileId
    ? { fileId: form.coverFileId, url: form.coverUrl }
    : await uploadImage(form.coverUrl, "dish_cover");
  const steps = await Promise.all((form.steps || []).map(async (step, index) => {
    const image = step.imageFileId
      ? { fileId: step.imageFileId, url: step.imageUrl }
      : step.imageUrl
        ? await uploadImage(step.imageUrl, "dish_step")
        : null;
    return {
      text: step.text,
      imageFileId: image ? image.fileId : null,
      sortOrder: index + 1,
    };
  }));
  return {
    name: form.name,
    category: form.category,
    tags: form.tags || [],
    serving: Number(form.serving),
    description: form.description || "",
    coverFileId: cover.fileId,
    status: form.status,
    ingredients: (form.ingredients || []).map((item, index) => ({
      name: item.name,
      amount: item.amount,
      unit: item.unit,
      note: item.note || "",
      sortOrder: index + 1,
    })),
    steps,
  };
}

module.exports = { uploadImage, normalizeDishPayload };
