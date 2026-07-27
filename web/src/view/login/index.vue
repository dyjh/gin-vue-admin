<template>
  <main id="userLayout" class="auth-page login-page">
    <section class="auth-shell">
      <AuthBrandPanel mode="login" />

      <div class="auth-main">
        <header class="auth-main-header">
          <span class="auth-system-label">管理后台</span>
          <span class="auth-secure-badge">
            <i></i>
            安全访问
          </span>
        </header>

        <div class="login-content">
          <div class="login-heading">
            <h2>管理员登录</h2>
            <p>请输入管理员账号和密码。</p>
          </div>

          <el-form
            ref="loginForm"
            :model="loginFormData"
            :rules="rules"
            :validate-on-rule-change="false"
            class="login-form"
            @keyup.enter="submitForm"
          >
            <div class="form-field-label">账号</div>
            <el-form-item prop="username">
              <el-input
                v-model="loginFormData.username"
                size="large"
                placeholder="请输入管理员账号"
                autocomplete="username"
              />
            </el-form-item>

            <div class="form-field-label">密码</div>
            <el-form-item prop="password">
              <el-input
                v-model="loginFormData.password"
                show-password
                size="large"
                type="password"
                placeholder="请输入登录密码"
                autocomplete="current-password"
              />
            </el-form-item>

            <template v-if="loginFormData.openCaptcha">
              <div class="form-field-label">安全验证</div>
              <el-form-item prop="captcha">
                <div class="captcha-row">
                  <el-input
                    v-model="loginFormData.captcha"
                    placeholder="请输入数字验证码"
                    size="large"
                    inputmode="numeric"
                    autocomplete="off"
                  />
                  <button
                    class="captcha-image"
                    type="button"
                    title="点击刷新验证码"
                    @click="loginVerify"
                  >
                    <img v-if="picPath" :src="picPath" alt="验证码，点击刷新" />
                    <span v-else>刷新验证码</span>
                  </button>
                </div>
              </el-form-item>
            </template>

            <el-button
              class="login-submit"
              type="primary"
              size="large"
              :loading="submitting"
              @click="submitForm"
            >
              登录后台
            </el-button>

            <button
              v-if="initNeeded"
              class="init-entry"
              type="button"
              :disabled="initChecking"
              @click="checkInit"
            >
              <span>首次部署？</span>
              开始系统初始化
              <strong>→</strong>
            </button>
          </el-form>

          <div class="login-help">
            <span>登录遇到问题？请联系系统管理员重置账号或密码。</span>
          </div>
        </div>
      </div>
    </section>

    <BottomInfo class="auth-footer" />
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { captcha } from '@/api/user'
import { checkDB } from '@/api/initdb'
import { useUserStore } from '@/pinia/modules/user'
import { isDev } from '@/utils/env.js'
import BottomInfo from '@/components/bottomInfo/bottomInfo.vue'
import AuthBrandPanel from '@/components/auth/AuthBrandPanel.vue'

defineOptions({
  name: 'Login'
})

const router = useRouter()
const userStore = useUserStore()
const loginForm = ref(null)
const picPath = ref('')
const captchaRequiredLength = ref(6)
const submitting = ref(false)
const initNeeded = ref(false)
const initChecking = ref(false)

const loginFormData = reactive({
  username: '',
  password: '',
  captcha: '',
  captchaId: '',
  openCaptcha: false
})

const checkUsername = (_rule, value, callback) => {
  if ((value || '').length < 5) {
    callback(new Error('请输入正确的用户名'))
    return
  }
  callback()
}

const checkPassword = (_rule, value, callback) => {
  if ((value || '').length < 6) {
    callback(new Error('请输入正确的密码'))
    return
  }
  callback()
}

const checkCaptcha = (_rule, value, callback) => {
  if (!loginFormData.openCaptcha) {
    callback()
    return
  }

  const sanitizedValue = (value || '').replace(/\s+/g, '')
  if (!sanitizedValue) {
    callback(new Error('请输入验证码'))
    return
  }
  if (!/^\d+$/.test(sanitizedValue)) {
    callback(new Error('验证码须为数字'))
    return
  }
  if (sanitizedValue.length < captchaRequiredLength.value) {
    callback(new Error(`请输入至少${captchaRequiredLength.value}位数字验证码`))
    return
  }
  if (sanitizedValue !== value) {
    loginFormData.captcha = sanitizedValue
  }
  callback()
}

const rules = reactive({
  username: [{ validator: checkUsername, trigger: 'blur' }],
  password: [{ validator: checkPassword, trigger: 'blur' }],
  captcha: [{ validator: checkCaptcha, trigger: 'blur' }]
})

const loginVerify = async () => {
  const result = await captcha()
  captchaRequiredLength.value = Number(result.data?.captchaLength) || 6
  picPath.value = result.data?.picPath
  loginFormData.captchaId = result.data?.captchaId
  loginFormData.openCaptcha = Boolean(result.data?.openCaptcha)
}

const submitForm = () => {
  loginForm.value?.validate(async (valid) => {
    if (!valid) {
      ElMessage({
        type: 'error',
        message: '请正确填写登录信息',
        showClose: true
      })
      return
    }

    submitting.value = true
    try {
      const success = await userStore.LoginIn(loginFormData)
      if (!success) {
        await loginVerify()
      }
    } finally {
      submitting.value = false
    }
  })
}

const resolveInitState = async () => {
  initNeeded.value = false
  if (!isDev) return

  initChecking.value = true
  try {
    const result = await checkDB()
    initNeeded.value = result.code === 0 && Boolean(result.data?.needInit)
  } catch (_) {
    initNeeded.value = false
  } finally {
    initChecking.value = false
  }
}

const checkInit = async () => {
  if (initChecking.value) return

  initChecking.value = true
  try {
    const result = await checkDB()
    if (result.code === 0 && result.data?.needInit) {
      userStore.NeedInit()
      await router.push({ name: 'Init' })
      return
    }

    initNeeded.value = false
    ElMessage({
      type: 'info',
      message: '系统已经初始化，无需重复配置'
    })
  } finally {
    initChecking.value = false
  }
}

onMounted(() => {
  loginVerify()
  resolveInitState()
})
</script>

<style scoped>
.auth-page {
  --auth-bg: #f3f7fc;
  --auth-panel: #ffffff;
  --auth-text: #17233a;
  --auth-muted: #708097;
  position: relative;
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 28px 28px 64px;
  background: var(--auth-bg);
  color: var(--auth-text);
  font-family: "Avenir Next", "HarmonyOS Sans SC", "Microsoft YaHei", sans-serif;
}

.auth-shell {
  position: relative;
  z-index: 1;
  display: grid;
  width: min(920px, 100%);
  min-height: 560px;
  grid-template-columns: 340px minmax(420px, 1fr);
  gap: 0;
  overflow: hidden;
  border: 1px solid rgb(173 190 211 / 34%);
  border-radius: 14px;
  background: var(--auth-panel);
  box-shadow: 0 14px 38px rgb(31 57 90 / 10%);
}

.auth-shell :deep(.auth-brand-panel) {
  min-height: 560px;
  border-radius: 0;
}

.auth-main {
  display: flex;
  min-width: 0;
  flex-direction: column;
  padding: 30px 48px 36px;
  background: #fff;
}

.auth-main-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--auth-muted);
  font-size: 12px;
}

.auth-system-label {
  font-weight: 650;
  letter-spacing: 0.12em;
}

.auth-secure-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
}

.auth-secure-badge i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #20b486;
  box-shadow: 0 0 0 4px rgb(32 180 134 / 10%);
}

.login-content {
  width: min(390px, 100%);
  margin: auto;
}

.login-heading h2 {
  margin: 0;
  color: var(--auth-text);
  font-size: 26px;
  font-weight: 650;
  letter-spacing: -0.04em;
  line-height: 1.3;
}

.login-heading p {
  margin: 9px 0 0;
  color: var(--auth-muted);
  font-size: 13px;
  line-height: 1.7;
}

.login-form {
  margin-top: 28px;
}

.form-field-label {
  margin: 0 0 8px;
  color: #35445b;
  font-size: 13px;
  font-weight: 600;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.login-form :deep(.el-input__wrapper) {
  min-height: 48px;
  border: 1px solid #dfe7f1;
  border-radius: 10px;
  background: #f9fbfe;
  box-shadow: none;
}

.login-form :deep(.el-input__wrapper:hover) {
  border-color: var(--el-color-primary-light-5);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  border-color: var(--el-color-primary);
  background: #fff;
  box-shadow: 0 0 0 3px var(--el-color-primary-light-9);
}
.login-form :deep(.el-input__inner) {
  color: #24344b;
  -webkit-text-fill-color: #24344b;
}

.login-form :deep(.el-input__inner::placeholder) {
  color: #9aa6b5;
  -webkit-text-fill-color: #9aa6b5;
}

.captcha-row {
  display: grid;
  width: 100%;
  grid-template-columns: minmax(0, 1fr) 128px;
  gap: 10px;
}

.captcha-image {
  height: 48px;
  overflow: hidden;
  padding: 0;
  border: 1px solid #dfe7f1;
  border-radius: 10px;
  background: #eef5ff;
  color: var(--el-color-primary);
  cursor: pointer;
  font-size: 12px;
}

.captcha-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.login-submit {
  width: 100%;
  height: 48px;
  margin-top: 4px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.08em;
  box-shadow: none;
}

.init-entry {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 7px;
  margin-top: 14px;
  padding: 11px 14px;
  border: 1px dashed #c9d7e8;
  border-radius: 9px;
  background: #f8fbff;
  color: #4f6078;
  cursor: pointer;
  font-size: 12px;
}

.init-entry span {
  color: #8b99aa;
}

.init-entry strong {
  color: var(--el-color-primary);
  font-size: 15px;
}

.init-entry:hover {
  border-color: var(--el-color-primary-light-5);
  color: var(--el-color-primary);
}

.login-help {
  margin-top: 22px;
  padding-top: 17px;
  border-top: 1px solid #edf1f6;
  color: #9aa6b5;
  font-size: 11px;
  line-height: 1.6;
  text-align: center;
}

.auth-footer {
  position: absolute;
  z-index: 2;
  right: 0;
  bottom: 12px;
  left: 0;
  width: 100%;
  color: #8290a3;
}

.auth-footer :deep(a) {
  color: #55749b;
}

@media (max-width: 900px) {
  .auth-page {
    align-items: flex-start;
    overflow: auto;
    padding: 18px 16px 76px;
  }

  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-shell :deep(.auth-brand-panel) {
    min-height: 210px;
  }

  .auth-main {
    min-height: 570px;
    padding: 25px 28px 34px;
  }
}

@media (max-width: 520px) {
  .auth-main {
    padding: 22px 20px 30px;
  }

  .captcha-row {
    grid-template-columns: minmax(0, 1fr) 112px;
  }
}
</style>