<template>
  <div class="page-container">
    <div class="page-header">
      <h2>指令管理</h2>
    </div>

    <el-card>
      <el-table :data="commands" stripe v-loading="loading">
        <el-table-column prop="id" label="指令ID" width="280" />
        <el-table-column prop="device_sn" label="目标设备" width="160" />
        <el-table-column prop="cmd_type" label="指令类型" width="130">
          <template #default="{ row }">
            <el-tag size="small">{{ cmdTypeText(row.cmd_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="params" label="参数" min-width="200">
          <template #default="{ row }">
            <span style="font-size: 12px; color: #909399">{{ JSON.stringify(row.params) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_by" label="下发人" width="120" />
        <el-table-column prop="created_at" label="下发时间" width="170" />
        <el-table-column prop="responded_at" label="响应时间" width="170" />
      </el-table>

      <div style="margin-top: 16px; text-align: right">
        <el-pagination
          v-model:current-page="page"
          :total="total"
          :page-size="20"
          layout="total, prev, pager, next"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const loading = ref(false)
const page = ref(1)
const total = ref(3)

const commands = ref([
  { id: 'cmd-20260619-001', device_sn: 'AP3000-001', cmd_type: 'start_charge', params: { port: 1, duration: 120 }, status: 'success', created_by: 'admin', created_at: '2026-06-19 00:30', responded_at: '2026-06-19 00:30:02' },
  { id: 'cmd-20260619-002', device_sn: 'TF100-001', cmd_type: 'stop_charge', params: { gun: 1 }, status: 'success', created_by: 'admin', created_at: '2026-06-19 00:25', responded_at: '2026-06-19 00:25:01' },
  { id: 'cmd-20260619-003', device_sn: 'AP3000-005', cmd_type: 'reboot', params: {}, status: 'pending', created_by: 'admin', created_at: '2026-06-19 00:20', responded_at: null },
])

function cmdTypeText(type: string) {
  const map: Record<string, string> = {
    start_charge: '启动充电',
    stop_charge: '停止充电',
    config: '参数配置',
    reboot: '远程重启',
  }
  return map[type] || type
}

function statusTag(status: string) {
  const map: Record<string, string> = {
    success: 'success',
    failed: 'danger',
    pending: 'warning',
    sent: 'info',
    timeout: 'danger',
  }
  return map[status] || 'info'
}
</script>
