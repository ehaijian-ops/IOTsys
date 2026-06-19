<template>
  <div class="page-container">
    <div class="page-header">
      <h2>仪表盘</h2>
      <span class="header-time">{{ currentTime }}</span>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon online">
              <el-icon :size="28"><Monitor /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.totalDevices }}</div>
              <div class="stat-label">设备总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon charging">
              <el-icon :size="28"><VideoPlay /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.onlineDevices }}</div>
              <div class="stat-label">在线设备</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon fault">
              <el-icon :size="28"><WarningFilled /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.faultDevices }}</div>
              <div class="stat-label">故障设备</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon energy">
              <el-icon :size="28"><DataLine /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.todayEnergy }}</div>
              <div class="stat-label">今日充电量 (kWh)</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 设备类型分布 & 在线率 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card class="dashboard-card">
          <template #header>
            <span>设备类型分布</span>
          </template>
          <div ref="typeChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="dashboard-card">
          <template #header>
            <span>设备在线率趋势</span>
          </template>
          <div ref="onlineChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近告警 -->
    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="dashboard-card">
          <template #header>
            <span>最近告警</span>
          </template>
          <el-table :data="recentAlerts" style="width: 100%" size="small">
            <el-table-column prop="device_sn" label="设备" width="180" />
            <el-table-column prop="alert_type" label="告警类型" width="150">
              <template #default="{ row }">
                <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">
                  {{ row.alert_type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="告警内容" />
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'pending' ? 'danger' : 'success'" size="small">
                  {{ row.status === 'pending' ? '待处理' : '已处理' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import dayjs from 'dayjs'

const currentTime = ref(dayjs().format('YYYY-MM-DD HH:mm:ss'))
let timer: number

const stats = reactive({
  totalDevices: 0,
  onlineDevices: 0,
  faultDevices: 0,
  todayEnergy: '0',
})

const recentAlerts = ref([
  { device_sn: 'AP3000-001', alert_type: '过温保护', severity: 'warning', message: '设备温度超过85℃', created_at: '2026-06-19 00:30', status: 'pending' },
  { device_sn: 'TF100-002', alert_type: '通信故障', severity: 'critical', message: '设备离线超过5分钟', created_at: '2026-06-19 00:15', status: 'pending' },
  { device_sn: 'AP3000-005', alert_type: '过流保护', severity: 'warning', message: '端口3电流异常', created_at: '2026-06-18 23:50', status: 'resolved' },
])

const typeChartRef = ref<HTMLDivElement>()
const onlineChartRef = ref<HTMLDivElement>()
let typeChart: echarts.ECharts | null = null
let onlineChart: echarts.ECharts | null = null

onMounted(() => {
  timer = window.setInterval(() => {
    currentTime.value = dayjs().format('YYYY-MM-DD HH:mm:ss')
  }, 1000)

  // 设备类型分布图
  if (typeChartRef.value) {
    typeChart = echarts.init(typeChartRef.value)
    typeChart.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: '0%' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}: {c}台' },
        data: [
          { value: 85, name: '电单车充电桩' },
          { value: 42, name: '汽车充电桩' },
        ],
      }],
    })
  }

  // 在线率趋势图
  if (onlineChartRef.value) {
    onlineChart = echarts.init(onlineChartRef.value)
    onlineChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'] },
      yAxis: { type: 'value', min: 80, max: 100 },
      series: [{
        data: [98, 97, 96, 95, 97, 98, 99],
        type: 'line',
        smooth: true,
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64,158,255,0.3)' },
            { offset: 1, color: 'rgba(64,158,255,0.05)' },
          ]),
        },
        itemStyle: { color: '#409eff' },
      }],
    })
  }
})

onUnmounted(() => {
  clearInterval(timer)
  typeChart?.dispose()
  onlineChart?.dispose()
})
</script>

<style scoped>
.header-time {
  color: #909399;
  font-size: 14px;
}

.stat-row {
  margin-bottom: 0;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.stat-icon.online { background: linear-gradient(135deg, #67c23a, #85ce61); }
.stat-icon.charging { background: linear-gradient(135deg, #409eff, #66b1ff); }
.stat-icon.fault { background: linear-gradient(135deg, #f56c6c, #f89898); }
.stat-icon.energy { background: linear-gradient(135deg, #e6a23c, #ebb563); }

.stat-content {
  flex: 1;
}
</style>
