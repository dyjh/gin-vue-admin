<template>
  <main class="auth-page init-page">
    <section class="auth-shell init-shell">
      <AuthBrandPanel mode="init" />

      <div class="auth-main init-main">
        <header class="setup-header">
          <button type="button" class="back-login" @click="backToLogin">
            ← 返回登录
          </button>
          <div class="setup-steps" aria-label="初始化进度">
            <div :class="['setup-step', { 'is-active': currentStep === 0, 'is-done': currentStep > 0 }]">
              <span>1</span>
              环境确认
            </div>
            <i></i>
            <div :class="['setup-step', { 'is-active': currentStep === 1 }]">
              <span>2</span>
              数据库配置
            </div>
          </div>
        </header>

        <section v-if="currentStep === 0" class="setup-welcome">
          <div class="setup-heading">

            <h2>开始系统初始化</h2>
            <p>整个过程通常只需几分钟。开始前，请确认下面的运行条件已经准备完成。</p>
          </div>

          <div class="readiness-grid">
            <article>
              <span class="readiness-number">01</span>
              <div>
                <strong>数据库服务</strong>
                <p>目标数据库已启动，并允许当前服务连接。</p>
              </div>
            </article>
            <article>
              <span class="readiness-number">02</span>
              <div>
                <strong>连接账户</strong>
                <p>账号具备建库、建表与写入初始数据权限。</p>
              </div>
            </article>
            <article>
              <span class="readiness-number">03</span>
              <div>
                <strong>字符编码</strong>
                <p>MySQL 建议使用 utf8mb4 与 InnoDB 引擎。</p>
              </div>
            </article>
            <article>
              <span class="readiness-number">04</span>
              <div>
                <strong>管理员密码</strong>
                <p>准备一个不少于 6 位的后台管理员密码。</p>
              </div>
            </article>
          </div>

          <div class="setup-notice">
            <span>i</span>
            <p>
              初始化会创建系统所需的数据结构和基础数据。已有业务数据库请勿重复执行。
              <button type="button" @click="goDoc">查看环境配置说明</button>
            </p>
          </div>

          <div class="setup-actions">
            <el-button size="large" @click="backToLogin">暂不初始化</el-button>
            <el-button type="primary" size="large" @click="currentStep = 1">
              开始配置
              <span class="button-arrow">→</span>
            </el-button>
          </div>
        </section>

        <section v-else class="setup-form-section">
          <div class="setup-heading setup-form-heading">

            <h2>连接数据库</h2>
            <p>填写服务端可访问的数据库信息，初始化完成后即可使用管理员账号登录。</p>
          </div>

          <div class="setup-form-scroll">
            <el-form
              ref="formRef"
              :model="form"
              label-position="top"
              size="large"
              class="setup-form"
            >
              <div class="form-grid">
                <el-form-item label="管理员密码" class="is-full">
                  <el-input
                    v-model="form.adminPassword"
                    type="password"
                    show-password
                    autocomplete="new-password"
                    placeholder="不少于 6 位，用于 admin 账号"
                  />
                </el-form-item>

                <el-form-item label="数据库类型">
                  <el-select
                    v-model="form.dbType"
                    placeholder="请选择数据库类型"
                    class="field-full"
                    @change="changeDB"
                  >
                    <el-option key="mysql" label="MySQL" value="mysql" />
                    <el-option key="pgsql" label="PostgreSQL" value="pgsql" />
                    <el-option key="oracle" label="Oracle" value="oracle" />
                    <el-option key="mssql" label="SQL Server" value="mssql" />
                    <el-option key="sqlite" label="SQLite" value="sqlite" />
                  </el-select>
                </el-form-item>

                <el-form-item label="数据库名称">
                  <el-input v-model="form.dbName" placeholder="请输入数据库名称" />
                </el-form-item>

                <template v-if="form.dbType !== 'sqlite'">
                  <el-form-item label="主机地址">
                    <el-input v-model="form.host" placeholder="例如 127.0.0.1" />
                  </el-form-item>
                  <el-form-item label="端口">
                    <el-input v-model="form.port" placeholder="请输入数据库端口" />
                  </el-form-item>
                  <el-form-item label="用户名">
                    <el-input v-model="form.userName" placeholder="请输入数据库用户名" />
                  </el-form-item>
                  <el-form-item label="数据库密码">
                    <el-input
                      v-model="form.password"
                      type="password"
                      show-password
                      autocomplete="off"
                      placeholder="没有密码可留空"
                    />
                  </el-form-item>
                </template>

                <el-form-item v-if="form.dbType === 'sqlite'" label="数据库文件路径" class="is-full">
                  <el-input v-model="form.dbPath" placeholder="请输入 SQLite 数据库文件存放路径" />
                </el-form-item>

                <el-form-item v-if="form.dbType === 'pgsql'" label="Template" class="is-full">
                  <el-input v-model="form.template" placeholder="例如 template0" />
                </el-form-item>
              </div>
            </el-form>

            <div class="database-tip">
              <span></span>
              当前选择 {{ databaseTypeLabel }}，请确认地址是后端服务实际可访问的地址。
            </div>
          </div>

          <div class="setup-actions form-actions">
            <el-button size="large" @click="currentStep = 0">上一步</el-button>
            <el-button type="primary" size="large" @click="onSubmit">
              创建并初始化
            </el-button>
          </div>
        </section>
      </div>
    </section>
  </main>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElLoading, ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { initDB } from '@/api/initdb'
import AuthBrandPanel from '@/components/auth/AuthBrandPanel.vue'

defineOptions({
  name: 'Init'
})

const router = useRouter()
const formRef = ref(null)
const currentStep = ref(0)

const form = reactive({
  adminPassword: '123456',
  dbType: 'mysql',
  host: '127.0.0.1',
  port: '3306',
  userName: 'root',
  password: '',
  dbName: 'gva',
  dbPath: ''
})

const databaseTypeLabel = computed(
  () =>
    ({
      mysql: 'MySQL',
      pgsql: 'PostgreSQL',
      oracle: 'Oracle',
      mssql: 'SQL Server',
      sqlite: 'SQLite'
    })[form.dbType] || form.dbType
)

const backToLogin = () => {
  router.push({ name: 'Login' })
}

const goDoc = () => {
  window.open('https://www.gin-vue-admin.com/guide/start-quickly/env.html')
}

const changeDB = (value) => {
  const common = {
    adminPassword: form.adminPassword || '123456',
    password: '',
    dbName: 'gva',
    dbPath: ''
  }

  const presets = {
    mysql: {
      dbType: 'mysql',
      host: '127.0.0.1',
      port: '3306',
      userName: 'root'
    },
    pgsql: {
      dbType: 'pgsql',
      host: '127.0.0.1',
      port: '5432',
      userName: 'postgres',
      template: 'template0'
    },
    oracle: {
      dbType: 'oracle',
      host: '127.0.0.1',
      port: '1521',
      userName: 'oracle'
    },
    mssql: {
      dbType: 'mssql',
      host: '127.0.0.1',
      port: '1433',
      userName: 'mssql'
    },
    sqlite: {
      dbType: 'sqlite',
      host: '',
      port: '',
      userName: ''
    }
  }

  Object.assign(form, common, presets[value] || presets.mysql)
  if (value !== 'pgsql') {
    delete form.template
  }
}

const onSubmit = async () => {
  if (form.adminPassword.length < 6) {
    ElMessage({
      type: 'error',
      message: '管理员密码长度不能少于 6 位'
    })
    return
  }

  const loading = ElLoading.service({
    lock: true,
    text: '正在创建数据库结构与基础数据，请稍候',
    spinner: 'loading',
    background: 'rgba(12, 31, 56, 0.72)'
  })

  try {
    const result = await initDB(form)
    if (result.code !== 0) return

    ElMessage({
      type: 'success',
      message: result.msg
    })

    try {
      await ElMessageBox.confirm(
        '基础数据库和管理员账号已经创建完成，现在可以登录后台。',
        '初始化完成',
        {
          confirmButtonText: '前往登录',
          cancelButtonText: '留在当前页',
          type: 'success',
          center: true
        }
      )
      await router.push({ name: 'Login' })
    } catch (_) {
      // 用户选择留在当前页，保留当前结果状态。
    }
  } finally {
    loading.close()
  }
}
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
  padding: 28px;
  background: var(--auth-bg);
  color: var(--auth-text);
  font-family: "Avenir Next", "HarmonyOS Sans SC", "Microsoft YaHei", sans-serif;
}

.auth-shell {
  position: relative;
  z-index: 1;
  display: grid;
  width: min(1040px, 100%);
  min-height: 680px;
  grid-template-columns: 300px minmax(520px, 1fr);
  overflow: hidden;
  border: 1px solid rgb(173 190 211 / 34%);
  border-radius: 14px;
  background: var(--auth-panel);
  box-shadow: 0 14px 38px rgb(31 57 90 / 10%);
}

.auth-shell :deep(.auth-brand-panel) {
  min-height: 680px;
  border-radius: 0;
}

.auth-main {
  display: flex;
  min-width: 0;
  max-height: 680px;
  flex-direction: column;
  padding: 26px 34px 30px;
  background: #fff;
}

.setup-header {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding-bottom: 22px;
  border-bottom: 1px solid #edf1f6;
}

.back-login {
  padding: 0;
  border: 0;
  background: transparent;
  color: #708097;
  cursor: pointer;
  font-size: 12px;
}

.back-login:hover {
  color: var(--el-color-primary);
}

.setup-steps {
  display: flex;
  align-items: center;
  gap: 9px;
}

.setup-steps > i {
  width: 26px;
  height: 1px;
  background: #dce5ef;
}

.setup-step {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #9aa6b5;
  font-size: 11px;
  font-style: normal;
}

.setup-step span {
  display: inline-flex;
  width: 21px;
  height: 21px;
  align-items: center;
  justify-content: center;
  border: 1px solid #dce5ef;
  border-radius: 50%;
  font-size: 10px;
}

.setup-step.is-active,
.setup-step.is-done {
  color: var(--el-color-primary);
  font-weight: 600;
}

.setup-step.is-active span,
.setup-step.is-done span {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.setup-welcome,
.setup-form-section {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
}

.setup-heading {
  flex: none;
  margin-top: 24px;
}

.setup-heading h2 {
  margin: 9px 0 0;
  color: var(--auth-text);
  font-size: 25px;
  font-weight: 650;
  letter-spacing: -0.04em;
  line-height: 1.3;
}

.setup-heading p {
  margin: 8px 0 0;
  color: var(--auth-muted);
  font-size: 12px;
  line-height: 1.65;
}

.readiness-grid {
  display: grid;
  margin-top: 22px;
  overflow: hidden;
  border: 1px solid #e4eaf2;
  border-radius: 9px;
  background: #fff;
}

.readiness-grid article {
  display: flex;
  min-height: 74px;
  gap: 12px;
  padding: 13px 14px;
  border-bottom: 1px solid #edf1f6;
}

.readiness-grid article:last-child {
  border-bottom: 0;
}

.readiness-number {
  display: inline-flex;
  width: 29px;
  height: 29px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 10px;
  font-weight: 700;
}

.readiness-grid strong {
  color: #293850;
  font-size: 13px;
  font-weight: 650;
}

.readiness-grid p {
  margin: 6px 0 0;
  color: #8492a5;
  font-size: 11px;
  line-height: 1.55;
}

.setup-notice {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 16px;
  padding: 12px 13px;
  border: 1px solid #dceafb;
  border-radius: 9px;
  background: #f4f9ff;
}

.setup-notice > span {
  display: inline-flex;
  width: 18px;
  height: 18px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--el-color-primary-light-8);
  color: var(--el-color-primary);
  font-size: 11px;
  font-weight: 700;
}

.setup-notice p {
  margin: 0;
  color: #60728a;
  font-size: 11px;
  line-height: 1.65;
}

.setup-notice button {
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--el-color-primary);
  cursor: pointer;
  font: inherit;
}

.setup-actions {
  display: flex;
  flex: none;
  justify-content: flex-end;
  gap: 10px;
  margin-top: auto;
  padding-top: 24px;
}

.setup-actions .el-button {
  min-width: 112px;
  height: 44px;
  margin-left: 0;
  border-radius: 9px;
}

.setup-actions .el-button:not(.el-button--primary) {
  border-color: #d5deea;
  background: #fff;
  color: #53657b;
}
.button-arrow {
  margin-left: 8px;
  font-size: 16px;
}

.setup-form-heading {
  margin-top: 23px;
}

.setup-form-scroll {
  min-height: 0;
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  margin: 19px -9px 0 0;
  padding-right: 9px;
  scrollbar-color: #cbd8e8 transparent;
  scrollbar-width: thin;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 13px;
}

.form-grid .is-full {
  grid-column: 1 / -1;
}

.field-full {
  width: 100%;
}

.setup-form :deep(.el-form-item) {
  margin-bottom: 15px;
}

.setup-form :deep(.el-form-item__label) {
  height: auto;
  margin-bottom: 7px;
  color: #3c4c62;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
}

.setup-form :deep(.el-input__wrapper),
.setup-form :deep(.el-select__wrapper) {
  min-height: 43px;
  border: 1px solid #dfe7f1;
  border-radius: 9px;
  background: #f9fbfe;
  box-shadow: none;
}

.setup-form :deep(.el-input__wrapper.is-focus),
.setup-form :deep(.el-select__wrapper.is-focused) {
  border-color: var(--el-color-primary);
  background: #fff;
  box-shadow: 0 0 0 3px var(--el-color-primary-light-9);
}
.setup-form :deep(.el-input__inner) {
  color: #24344b;
  -webkit-text-fill-color: #24344b;
}

.setup-form :deep(.el-input__inner::placeholder) {
  color: #9aa6b5;
  -webkit-text-fill-color: #9aa6b5;
}

.database-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 11px;
  border-radius: 8px;
  background: #f6f8fb;
  color: #7b899c;
  font-size: 10px;
  line-height: 1.5;
}

.database-tip span {
  width: 7px;
  height: 7px;
  flex: none;
  border-radius: 50%;
  background: #20b486;
  box-shadow: 0 0 0 4px rgb(32 180 134 / 10%);
}

.form-actions {
  margin-top: 0;
  border-top: 1px solid #edf1f6;
}

@media (max-width: 920px) {
  .auth-page {
    align-items: flex-start;
    overflow: auto;
    padding: 18px 16px;
  }

  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-shell :deep(.auth-brand-panel) {
    min-height: 210px;
  }

  .auth-main {
    max-height: none;
    min-height: 690px;
    padding: 24px 28px 30px;
  }
}

@media (max-width: 560px) {
  .auth-main {
    padding: 22px 20px 28px;
  }

  .setup-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .form-grid,
  .readiness-grid {
    grid-template-columns: 1fr;
  }

  .form-grid .is-full {
    grid-column: auto;
  }

  .setup-actions {
    width: 100%;
  }

  .setup-actions .el-button {
    min-width: 0;
    flex: 1;
  }
}
</style>