<template>
  <div class="page-container">
    <div class="page-header">
      <h2>站点管理</h2>
      <el-button type="primary"><el-icon><Plus /></el-icon> 新增站点</el-button>
    </div>

    <el-row :gutter="20">
      <el-col :span="8" v-for="site in sites" :key="site.id" style="margin-bottom: 20px">
        <el-card class="dashboard-card">
          <template #header>
            <div class="site-header">
              <span>{{ site.name }}</span>
              <el-tag :type="site.status === 'active' ? 'success' : 'info'" size="small">
                {{ site.status === 'active' ? '运营中' : '停用' }}
              </el-tag>
            </div>
          </template>
          <div class="site-info">
            <p><el-icon><Location /></el-icon> {{ site.address }}</p>
            <p><el-icon><Monitor /></el-icon> 设备: {{ site.device_count }} 台</p>
            <p><el-icon><User /></el-icon> 联系人: {{ site.contact }}</p>
          </div>
          <div class="site-actions">
            <el-button size="small">详情</el-button>
            <el-button size="small" type="primary">编辑</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const sites = ref([
  { id: '1', name: '城东充电站', address: '城东区创业路100号', device_count: 45, contact: '张工', status: 'active' },
  { id: '2', name: '万达广场充电站', address: '万达广场B1停车场', device_count: 30, contact: '李工', status: 'active' },
  { id: '3', name: '高铁站充电站', address: '高铁站南广场', device_count: 20, contact: '王工', status: 'active' },
])
</script>

<style scoped>
.site-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.site-info p {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.site-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}
</style>
