<template>
  <aside class="auth-brand-panel">
    <header class="auth-brand-header">
      <div class="auth-brand-logo">
        <img src="/logo.png" :alt="$GIN_VUE_ADMIN.appName" />
      </div>
      <div>
        <strong>{{ $GIN_VUE_ADMIN.appName }}</strong>
        <span>后台管理系统</span>
      </div>
    </header>

    <div class="auth-brand-content">
      <h1>{{ content.title }}</h1>
      <p>{{ content.description }}</p>
    </div>

    <div class="auth-brand-list">
      <div v-for="item in content.items" :key="item">
        <span></span>
        {{ item }}
      </div>
    </div>

    <footer>OrderFood Mini App</footer>
  </aside>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  mode: {
    type: String,
    default: 'login'
  }
})

const content = computed(() =>
  props.mode === 'init'
    ? {
        title: '系统初始化',
        description: '配置数据库连接并创建管理员账号。',
        items: ['确认运行环境', '填写数据库信息', '完成基础数据初始化']
      }
    : {
        title: '欢迎使用',
        description: '登录后台管理用户、内容和运营配置。',
        items: ['内容与菜谱管理', '用户与积分管理', '平台能力配置']
      }
)
</script>

<style scoped>
.auth-brand-panel {
  display: flex;
  min-height: 560px;
  flex-direction: column;
  padding: 38px 34px 30px;
  background: #1f5fae;
  color: #fff;
}

.auth-brand-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auth-brand-logo {
  display: flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: rgb(255 255 255 / 14%);
}

.auth-brand-logo img {
  width: 30px;
  height: 30px;
  object-fit: contain;
}

.auth-brand-header strong,
.auth-brand-header span {
  display: block;
}

.auth-brand-header strong {
  font-size: 19px;
  font-weight: 650;
}

.auth-brand-header span {
  margin-top: 3px;
  color: rgb(255 255 255 / 66%);
  font-size: 11px;
}

.auth-brand-content {
  margin-top: 94px;
}

.auth-brand-content h1 {
  margin: 0;
  font-size: 30px;
  font-weight: 600;
  line-height: 1.35;
}

.auth-brand-content p {
  max-width: 270px;
  margin: 12px 0 0;
  color: rgb(255 255 255 / 72%);
  font-size: 13px;
  line-height: 1.7;
}

.auth-brand-list {
  display: grid;
  gap: 13px;
  margin-top: 40px;
}

.auth-brand-list > div {
  display: flex;
  align-items: center;
  gap: 9px;
  color: rgb(255 255 255 / 78%);
  font-size: 12px;
}

.auth-brand-list span {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #9fd0ff;
}

.auth-brand-panel footer {
  margin-top: auto;
  color: rgb(255 255 255 / 48%);
  font-size: 10px;
  letter-spacing: 0.08em;
}

@media (max-width: 900px) {
  .auth-brand-panel {
    min-height: 160px;
    padding: 24px;
  }

  .auth-brand-content {
    margin-top: 24px;
  }

  .auth-brand-content h1 {
    font-size: 24px;
  }

  .auth-brand-content p,
  .auth-brand-list,
  .auth-brand-panel footer {
    display: none;
  }
}
</style>