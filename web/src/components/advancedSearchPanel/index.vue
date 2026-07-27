<template>
  <div
    ref="panelRef"
    class="gva-search-box advanced-search-panel"
    :class="{ 'advanced-search-panel--expanded': expanded }"
  >
    <slot />

    <Teleport v-if="hasAdvancedFilters && actionTarget" :to="actionTarget">
      <el-button
        link
        type="primary"
        class="advanced-search-panel__toggle"
        @click="expanded = !expanded"
      >
        {{ expanded ? '收起筛选' : '高级筛选' }}
        <el-icon class="advanced-search-panel__toggle-icon">
          <ArrowDown />
        </el-icon>
      </el-button>
    </Teleport>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'

defineOptions({
  name: 'AdvancedSearchPanel'
})

const props = defineProps({
  basicCount: {
    type: Number,
    default: 3
  }
})

const panelRef = ref(null)
const expanded = ref(false)
const hasAdvancedFilters = ref(false)
const actionTarget = ref(null)
let observer

const classifyFormItems = () => {
  const form = panelRef.value?.querySelector(':scope > .el-form')
  if (!form) return

  const items = Array.from(form.children).filter((element) =>
    element.classList.contains('el-form-item')
  )
  const actionItem = items.at(-1)
  const filterItems = items.slice(0, -1)
  const advancedItems = filterItems.slice(props.basicCount)

  filterItems.forEach((item) => {
    item.classList.remove('advanced-search-panel__advanced-item')
  })
  advancedItems.forEach((item) => {
    item.classList.add('advanced-search-panel__advanced-item')
  })

  hasAdvancedFilters.value = advancedItems.length > 0
  actionTarget.value = hasAdvancedFilters.value ? actionItem : null
}

onMounted(async () => {
  await nextTick()
  classifyFormItems()

  const form = panelRef.value?.querySelector(':scope > .el-form')
  if (form) {
    observer = new MutationObserver(classifyFormItems)
    observer.observe(form, { childList: true })
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

<style scoped>
.advanced-search-panel:not(.advanced-search-panel--expanded)
  :deep(.advanced-search-panel__advanced-item) {
  display: none;
}

.advanced-search-panel__toggle-icon {
  margin-left: 4px;
}

.advanced-search-panel__toggle {
  margin-left: 12px !important;
}

.advanced-search-panel :deep(.el-date-editor--daterange),
.advanced-search-panel :deep(.el-date-editor--datetimerange),
.advanced-search-panel :deep(.el-date-editor--monthrange) {
  width: 22rem !important;
  min-width: 22rem;
}

.advanced-search-panel--expanded .advanced-search-panel__toggle-icon {
  transform: rotate(180deg);
}

</style>
