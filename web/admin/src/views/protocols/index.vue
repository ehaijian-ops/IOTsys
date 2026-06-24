<template>
  <div class="page-container">
    <div class="page-header">
      <h2>协议管理</h2>
      <el-button type="primary">
        <el-icon><Plus /></el-icon> 注册新协议
      </el-button>
    </div>

    <el-row :gutter="20">
      <el-col :span="8" v-for="p in protocols" :key="p.name">
        <el-card class="dashboard-card protocol-card">
          <div class="protocol-header">
            <el-tag :type="p.device_type === 'ev_charger' ? 'success' : 'warning'">
              {{ p.device_type === 'ev_charger' ? '汽车充电桩' : '电单车充电桩' }}
            </el-tag>
            <el-tag type="info" size="small">{{ p.status }}</el-tag>
          </div>
          <h3 class="protocol-name">{{ p.name }}</h3>
          <p class="protocol-version">版本: {{ p.version }}</p>
          <p class="protocol-desc">{{ p.description }}</p>
          <div class="protocol-meta">
            <span>通信方式: {{ p.comm_type }}</span>
            <span>在线设备: {{ p.online_count }}</span>
          </div>
          <div class="protocol-actions">
            <el-button size="small">查看详情</el-button>
            <el-button size="small" type="primary">配置</el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 添加新协议 -->
      <el-col :span="8">
        <el-card class="dashboard-card protocol-card add-card">
          <div class="add-protocol">
            <el-icon :size="48" color="#c0c4cc"><Plus /></el-icon>
            <p>添加新协议适配器</p>
            <span>支持自定义设备协议扩展</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const protocols = ref([
  {
    name: 'AP3000_v2',
    version: '8.6',
    device_type: 'ebike_charger',
    description: '安平科技电单车充电桩通信协议，支持TCP长连接，二进制帧格式(DNY)',
    comm_type: 'TCP',
    online_count: 85,
    status: '已注册',
  },
  {
    name: 'WSD_v1',
    version: '1.0',
    device_type: 'ebike_charger',
    description: '微小电12路电单车充电桩通信协议，支持TCP长连接，二进制帧格式(0xEE)，含34个主指令',
    comm_type: 'TCP',
    online_count: 12,
    status: '已注册',
  },
  {
    name: 'TF100_v1',
    version: '1.0',
    device_type: 'ev_charger',
    description: '特来电汽车充电桩通信协议，支持TCP连接，参照国标GB/T 27930',
    comm_type: 'TCP',
    online_count: 42,
    status: '已注册',
  },
])
</script>

<style scoped>
.protocol-card {
  min-height: 260px;
}

.protocol-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.protocol-name {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.protocol-version {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.protocol-desc {
  font-size: 13px;
  color: #606266;
  margin-bottom: 12px;
  line-height: 1.5;
}

.protocol-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.protocol-actions {
  display: flex;
  gap: 8px;
}

.add-card {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px dashed #dcdfe6;
  background: #fafafa;
  cursor: pointer;
}

.add-card:hover {
  border-color: #409eff;
  background: #f0f7ff;
}

.add-protocol {
  text-align: center;
}

.add-protocol p {
  margin-top: 12px;
  color: #606266;
  font-size: 15px;
}

.add-protocol span {
  color: #c0c4cc;
  font-size: 12px;
}
</style>
