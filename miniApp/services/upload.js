const env = require("../config/env");

function isRemoteOrBundled(filePath) {
  return !filePath || filePath.startsWith("/assets/") || /^https?:\/\//.test(filePath);
}

function uploadImage(filePath, scene) {
  if (isRemoteOrBundled(filePath)) {
    return Promise.resolve({ fileId: filePath || "", url: filePath || "" });
  }

  if (env.useMock) {
    return Promise.resolve({ fileId: `mock-file:${filePath}`, url: filePath, reviewStatus: "passed" });
  }

  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: `${env.baseUrl}/uploads/images`,
      filePath,
      name: "file",
      formData: { scene },
      header: {
        authorization: wx.getStorageSync("access_token") ? `Bearer ${wx.getStorageSync("access_token")}` : "",
      },
      success(response) {
        let body;
        try {
          body = JSON.parse(response.data);
        } catch (error) {
          reject(new Error("图片上传响应格式错误"));
          return;
        }
        if (response.statusCode >= 200 && response.statusCode < 300 && body.code === 0) {
          resolve(body.data);
          return;
        }
        const error = new Error(body.msg || "图片上传失败");
        wx.showToast({ title: error.message, icon: "none" });
        reject(error);
      },
      fail(error) {
        wx.showToast({ title: "图片上传失败，请稍后重试", icon: "none" });
        reject(error);
      },
    });
  });
}

async function normalizeDishPayload(form) {
  const cover = form.coverFileId ? { fileId: form.coverFileId, url: form.image } : await uploadImage(form.image, "dish_cover");
  const steps = await Promise.all((form.steps || []).map(async (step, index) => {
    const image = step.imageFileId ? { fileId: step.imageFileId, url: step.image } : step.image ? await uploadImage(step.image, "dish_step") : null;
    return {
      ...step,
      sortOrder: index,
      imageFileId: image ? image.fileId : "",
      image: image ? image.url : "",
    };
  }));
  return {
    ...form,
    coverFileId: cover.fileId,
    image: cover.url,
    ingredients: (form.ingredients || []).map((item, index) => ({ ...item, sortOrder: index })),
    steps,
  };
}

module.exports = { uploadImage, normalizeDishPayload };
