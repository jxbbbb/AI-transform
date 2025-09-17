<script setup lang="ts">
import Menu from './components/Menu.vue'
import TranslateRecords from './components/TranslateRecords/TranslateRecords.vue'
</script>

<template>
    <div class="common-layout;" style="width: 100%;height:100%;">
      <el-container style="width:100%;height:100%;">
        <el-aside >
          <Menu></Menu>
        </el-aside>
        <el-main >
          <TranslateRecords></TranslateRecords>
        </el-main>
      </el-container>
    </div>
</template>

<style scoped>
.el-aside {
    box-sizing: border-box;
    flex-shrink: 0;
    overflow: auto;
    width:auto;
    min-width: 4rem;
    max-width: 12.5rem;
}
</style>
