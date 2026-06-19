<template>
  <div class="page-container">
    <div class="page-header">
      <h2>数据报表</h2>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        style="width: 260px"
      />
    </div>

    <!-- 充电量统计 -->
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card class="dashboard-card">
          <template #header><span>充电量趋势</span></template>
          <div ref="energyChartRef" style="height: 350px"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="dashboard-card">
          <template #header><span>汇总数据</span></template>
          <div class="summary-list">
            <div class="summary-item">
              <span class="summary-label">总充电量</span>
              <span class="summary-value">12,345 kWh</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">总充电次数</span>
              <span class="summary-value">8,562 次</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">设备平均利用率</span>
              <span class="summary-value">76.5%</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">总收入</span>
              <span class="summary-value">¥15,678</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">故障率</span>
              <span class="summary-value">1.2%</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 设备利用率排行 -->
    <el-card style="margin-top: 20px" class="dashboard-card">
      <template #header><span>设备利用率排行</span></template>
      <el-table :data="rankings" stripe size="small">
        <el-table-column type="index" label="排名" width="60" />
        <el-table-column prop="sn" label="设备SN" width="180" />
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="usage_rate" label="利用率" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.usage_rate" :color="progressColor(row.usage_rate)" />
          </template>
        </el-table-column>
        <el-table-column prop="charge_count" label="充电次数" width="120" />
        <el-table-column prop="total_energy" label="充电量(kWh)" width="150" />
        <el-table-column prop="revenue" label="收入(¥)" width="120" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'

const dateRange = ref<[Date, Date]>()

const energyChartRef = ref<HTMLDivElement>()
let energyChart: echarts.ECharts | null = null

const rankings = ref([
  { sn: 'AP3000-001', type: '电单车充电桩', usage_rate: 92, charge_count: 245, total_energy: 456.8, revenue: 685.2 },
  { sn: 'TF100-001', type: '汽车充电桩', usage_rate: 88, charge_count: 120, total_energy: 2340.5, revenue: 3510.75 },
  { sn: 'AP3000-003', type: '电单车充电桩', usage_rate: 85, charge_count: 210, total_energy: 389.2, revenue: 583.8 },
  { sn: 'TF100-003', type: '汽车充电桩', usage_rate: 72, charge_count: 95, total_energy: 1890.3, revenue: 2835.45 },
  { sn: 'AP3000-002', type: '电单车充电桩', usage_rate: 68, charge_count: 180, total_energy: 312.4, revenue: 468.6 },
])

function progressColor(rate: number) {
  if (rate >= 90) return '#67c23a'
  if (rate >= 70) return '#409eff'
  if (rate >= 50) return '#e6a23c'
  return '#f56c6c'
}

onMounted(() => {
  if (energyChartRef.value) {
    energyChart = echarts.init(energyChartRef.value)
    energyChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['电单车充电桩', '汽车充电桩'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: ['6/13', '6/14', '6/15', '6/16', '6/17', '6/18', '6/19'] },
      yAxis: { type: 'value', name: '充电量(kWh)' },
      series: [
        {
          name: '电单车充电桩', type: 'bar', stack: 'total',
          data: [120, 132, 101, 134, 90, 230, 210],
          itemStyle: { color: '#409eff' },
        },
        {
          name: '汽车充电桩', type: 'bar', stack: 'total',
          data: [220, 182, 191, 234, 290, 330, 310],
          itemStyle: { color: '#67c23a' },
        },
      ],
    })
  }
})

onUnmounted(() => {
  energyChart?.dispose()
})
</script>

<style scoped>
.summary-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.summary-label {
  color: #909399;
  font-size: 14px;
}

.summary-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}
</style>
