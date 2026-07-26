<template>
  <CatalogManager
    resource="unit"
    title="单位"
    permission-prefix="orderfood:unit"
  />
</template>

<script setup>
import CatalogManager from '@/view/orderFood/components/CatalogManager.vue'

defineOptions({
  name: 'OrderFoodUnits'
})
</script>
