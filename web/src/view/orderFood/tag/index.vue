<template>
  <CatalogManager
    resource="tag"
    title="标签"
    permission-prefix="orderfood:tag"
  />
</template>

<script setup>
import CatalogManager from '@/view/orderFood/components/CatalogManager.vue'

defineOptions({
  name: 'OrderFoodTags'
})
</script>
