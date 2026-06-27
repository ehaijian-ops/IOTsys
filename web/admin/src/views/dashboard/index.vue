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
            <div class="stat-icon user"><el-icon :size="28"><User /></el-icon></div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.totalUsers }}</div>
              <div class="stat-label">用户总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon online"><el-icon :size="28"><Monitor /></el-icon></div>
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
            <div class="stat-icon energy"><el-icon :size="28"><DataLine /></el-icon></div>
            <div class="stat-content">
              <div class="stat-number">¥{{ stats.totalAmount }}</div>
              <div class="stat-label">交易总额</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="dashboard-card" shadow="hover">
          <div class="stat-item">
            <div class="stat-icon charging"><el-icon :size="28"><VideoPlay /></el-icon></div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.totalOrders }}</div>
              <div class="stat-label">总订单数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 订单类型分布 & 设备类型分布 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card class="dashboard-card">
          <template #header>
            <span>订单类型分布</span>
          </template>
          <div ref="orderChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="dashboard-card">
          <template #header>
            <span>设备类型分布</span>
          </template>
          <div ref="siteChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 月度统计 -->
    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="dashboard-card">
          <div ref="typeChartRef" style="height: 320px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近故障 -->
    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card class="dashboard-card">
          <template #header>
            <span>最近故障</span>
          </template>
          <el-table :data="recentFaults" style="width: 100%" size="small" v-loading="faultLoading">
            <el-table-column prop="device_sn" label="设备" width="180" />
            <el-table-column prop="device_type" label="设备类型" width="120" />
            <el-table-column prop="description" label="故障描述" show-overflow-tooltip />
            <el-table-column prop="reporter_name" label="上报人" width="120" />
            <el-table-column prop="created_at" label="时间" width="170" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'pending' ? 'danger' : row.status === 'processing' ? 'warning' : 'success'" size="small">
                  {{ row.status === 'pending' ? '待处理' : row.status === 'processing' ? '处理中' : '已处理' }}
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
import { getDevices } from '@/api/device'
import { getOrders } from '@/api/order'
import { getSitesManage } from '@/api/site'
import { getWechatUsers } from '@/api/advertisement'
import { getFaults } from '@/api/maintenance'

const currentTime = ref(dayjs().format('YYYY-MM-DD HH:mm:ss'))
let timer: number

const stats = reactive({
  totalUsers: 0,
  totalDevices: 0,
  totalAmount: '0',
  totalOrders: 0,
  onlineDevices: 0,
  totalSites: 0,
})

const recentFaults = ref<any[]>([])
const faultLoading = ref(false)

const typeChartRef = ref<HTMLDivElement>()
const orderChartRef = ref<HTMLDivElement>()
const siteChartRef = ref<HTMLDivElement>()
let typeChart: echarts.ECharts | null = null
let orderChart: echarts.ECharts | null = null
let siteChart: echarts.ECharts | null = null

async function fetchData() {
  try {
    const [deviceRes, orderRes, siteRes, userRes] = await Promise.all([
      getDevices({ page: 1, page_size: 1000 }),
      getOrders({ page: 1, page_size: 1000 }),
      getSitesManage({ page: 1, page_size: 100 }),
      getWechatUsers({ page: 1, page_size: 1000 }),
    ])
    const devices = deviceRes?.data || []
    const orders = orderRes?.data || []
    const sites = siteRes?.data || []
    const users = userRes?.data || []

    stats.totalDevices = devices.length
    stats.onlineDevices = devices.filter((d: any) => d.status === 'online').length
    stats.totalOrders = orders.length
    stats.totalSites = sites.length
    stats.totalUsers = users.length

    const totalAmount = orders
      .filter((o: any) => o.status === 'completed')
      .reduce((sum: number, o: any) => sum + (o.paid_amount || 0), 0)
    stats.totalAmount = totalAmount.toFixed(2)

    // 最近故障
    try {
      faultLoading.value = true
      const faultRes = await getFaults({ page: 1, page_size: 10 })
      recentFaults.value = (faultRes?.data || faultRes?.list || []).slice(0, 10)
    } catch { /* ignore */ } finally {
      faultLoading.value = false
    }
  } catch (e) { /* fallback to defaults */ }
}

function initCharts() {
  // 订单类型饼图
  if (orderChartRef.value) {
    orderChart = echarts.init(orderChartRef.value)
    orderChart.setOption({
      tooltip: { trigger: 'item' },
      title: { text: '订单类型分布', left: 'center', top: 10, textStyle: { fontSize: 14 } },
      legend: { bottom: 10 },
      series: [{
        type: 'pie', radius: ['40%', '65%'],
        data: [
          { value: 60, name: '扫码充电' },
          { value: 20, name: 'IC卡充电' },
          { value: 12, name: '包月充电' },
          { value: 8, name: '免费充电' },
        ],
      }],
    })
  }
  // 站点分布饼图
  if (siteChartRef.value) {
    siteChart = echarts.init(siteChartRef.value)
    siteChart.setOption({
      tooltip: { trigger: 'item' },
      title: { text: '设备类型分布', left: 'center', top: 10, textStyle: { fontSize: 14 } },
      legend: { bottom: 10 },
      series: [{
        type: 'pie', radius: ['40%', '65%'],
        data: [
          { value: 85, name: '电单车充电桩' },
          { value: 42, name: '汽车充电桩' },
        ],
      }],
    })
  }
  // 月度统计柱状图
  if (typeChartRef.value) {
    typeChart = echarts.init(typeChartRef.value)
    const months = []
    for (let i = 5; i >= 0; i--) months.push(dayjs().subtract(i, 'month').format('MM月'))
    typeChart.setOption({
      tooltip: { trigger: 'axis' },
      title: { text: '月度统计数据', left: 'center', top: 10, textStyle: { fontSize: 14 } },
      grid: { left: '3%', right: '4%', bottom: '3%', top: 50, containLabel: true },
      legend: { data: ['订单数', '交易额(元)', '用电量(kWh)'], top: 20 },
      xAxis: { type: 'category', data: months },
      yAxis: { type: 'value' },
      series: [
        { name: '订单数', type: 'bar', data: [320, 420, 380, 510, 480, 350], itemStyle: { color: '#409eff' }, barGap: '10%' },
        { name: '交易额(元)', type: 'bar', data: [1800, 2500, 2200, 3200, 2900, 2100], itemStyle: { color: '#67c23a' }, barGap: '10%' },
      ],
    })
  }
}

onMounted(() => {
  timer = window.setInterval(() => {
    currentTime.value = dayjs().format('YYYY-MM-DD HH:mm:ss')
  }, 1000)
  fetchData()
  initCharts()
})

onUnmounted(() => {
  clearInterval(timer)
  typeChart?.dispose()
  orderChart?.dispose()
  siteChart?.dispose()
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
