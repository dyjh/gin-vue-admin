<template>
  <CatalogManager
    resource="category"
    title="分类"
    permission-prefix="orderfood:category"
  />
</template>

<script setup>
import CatalogManager from '@/view/orderFood/components/CatalogManager.vue'

defineOptions({
  name: 'OrderFoodCategories'
})
</script>
