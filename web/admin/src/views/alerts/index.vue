<template>
  <div class="page-container">
    <div class="page-header">
      <h2>告警管理</h2>
    </div>

    <el-card>
      <!-- 筛选 -->
      <el-form :inline="true" style="margin-bottom: 16px">
        <el-form-item label="严重程度">
          <el-select v-model="severity" placeholder="全部" clearable style="width: 130px">
            <el-option label="严重" value="critical" />
            <el-option label="警告" value="warning" />
            <el-option label="提示" value="info" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="status" placeholder="全部" clearable style="width: 130px">
            <el-option label="待处理" value="pending" />
            <el-option label="已确认" value="confirmed" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="alerts" stripe>
        <el-table-column prop="device_sn" label="设备" width="160" />
        <el-table-column prop="alert_type" label="告警类型" width="140">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">
              {{ row.alert_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="severity" label="严重程度" width="100">
          <template #default="{ row }">
            <el-tag :type="severityTag(row.severity)" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="告警内容" min-width="250" />
        <el-table-column prop="created_at" label="触发时间" width="170" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'pending' ? 'danger' : 'success'" size="small">
              {{ row.status === 'pending' ? '待处理' : row.status === 'confirmed' ? '已确认' : '已解决' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending'" link type="primary" @click="handleAck(row)">确认</el-button>
            <el-button v-if="row.status !== 'resolved'" link type="success" @click="handleResolve(row)">解决</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; text-align: right">
        <el-pagination :total="5" :page-size="20" layout="total, prev, pager, next" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const severity = ref('')
const status = ref('')

const alerts = ref([
  { id: '1', device_sn: 'TF100-002', alert_type: '通信故障', severity: 'critical', message: '设备离线超过5分钟，请检查网络连接', created_at: '2026-06-19 00:15', status: 'pending' },
  { id: '2', device_sn: 'AP3000-005', alert_type: '过温保护', severity: 'warning', message: '设备温度达到87℃，超过阈值85℃', created_at: '2026-06-19 00:10', status: 'pending' },
  { id: '3', device_sn: 'AP3000-001', alert_type: '过流保护', severity: 'warning', message: '端口3电流异常，当前15.2A，阈值12A', created_at: '2026-06-18 23:50', status: 'resolved' },
])

function severityTag(severity: string) {
  const map: Record<string, string> = { critical: 'danger', warning: 'warning', info: 'info' }
  return map[severity] || 'info'
}

function handleAck(alert: any) {
  ElMessage.success(`告警 [${alert.alert_type}] 已确认`)
  alert.status = 'confirmed'
}

function handleResolve(alert: any) {
  ElMessage.success(`告警 [${alert.alert_type}] 已解决`)
  alert.status = 'resolved'
}
</script>
