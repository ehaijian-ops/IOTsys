<template>
  <div class="page-container">
    <div class="page-header">
      <h2>数据统计</h2>
    </div>

    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <!-- 经营统计 -->
      <el-tab-pane label="经营统计" name="business">
        <el-row :gutter="20" style="margin-bottom: 16px">
          <el-col :span="6" v-for="card in bizCards" :key="card.label">
            <el-card shadow="hover">
              <div class="stat-card">
                <div class="stat-card-value">{{ card.value }}</div>
                <div class="stat-card-label">{{ card.label }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-card>
          <div ref="bizChartRef" style="height: 350px"></div>
        </el-card>
      </el-tab-pane>

      <!-- 设备统计 -->
      <el-tab-pane label="设备统计" name="device">
        <el-row :gutter="20" style="margin-bottom: 16px">
          <el-col :span="6" v-for="card in devCards" :key="card.label">
            <el-card shadow="hover">
              <div class="stat-card">
                <div class="stat-card-value">{{ card.value }}</div>
                <div class="stat-card-label">{{ card.label }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card>
              <template #header><span>设备状态分布</span></template>
              <div ref="devStatusChartRef" style="height: 300px"></div>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card>
              <template #header><span>站点设备数 Top10</span></template>
              <div ref="devSiteChartRef" style="height: 300px"></div>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- 充电汇总 -->
      <el-tab-pane label="充电汇总" name="charge">
        <el-row :gutter="20" style="margin-bottom: 16px">
          <el-col :span="6" v-for="card in chargeCards" :key="card.label">
            <el-card shadow="hover">
              <div class="stat-card">
                <div class="stat-card-value">{{ card.value }}</div>
                <div class="stat-card-label">{{ card.label }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-card>
          <div style="margin-bottom: 16px">
            <el-date-picker
              v-model="chargeDateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 260px; margin-right: 12px"
            />
            <el-button type="primary" @click="loadChargeSummary">查询</el-button>
          </div>
          <div ref="chargeChartRef" style="height: 350px"></div>
        </el-card>
        <el-card style="margin-top: 16px">
          <template #header><span>充电记录明细</span></template>
          <el-table :data="chargeRecords" v-loading="chargeLoading" stripe>
            <el-table-column prop="order_sn" label="订单编号" width="200" />
            <el-table-column prop="device_sn" label="设备SN" width="160" />
            <el-table-column prop="order_type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ { scan: '扫码', ic_card: 'IC卡', free: '免费', monthly: '包月' }[row.order_type] || row.order_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="amount" label="金额(元)" width="100" align="right">
              <template #default="{ row }">¥{{ (row.amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="energy_kwh" label="电量(kWh)" width="100" align="right" />
            <el-table-column prop="start_time" label="开始时间" width="170" />
            <el-table-column prop="end_time" label="结束时间" width="170" />
            <el-table-column prop="duration_min" label="时长(分钟)" width="100" align="right" />
          </el-table>
          <div style="margin-top: 16px; text-align: right">
            <el-pagination
              v-model:current-page="chargePage"
              v-model:page-size="chargePageSize"
              :page-sizes="[10, 20, 50]"
              :total="chargeTotal"
              layout="total, sizes, prev, pager, next"
              @size-change="loadChargeSummary"
              @current-change="loadChargeSummary"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 交易汇总 -->
      <el-tab-pane label="交易汇总" name="trade">
        <el-row :gutter="20" style="margin-bottom: 16px">
          <el-col :span="6" v-for="card in tradeCards" :key="card.label">
            <el-card shadow="hover">
              <div class="stat-card">
                <div class="stat-card-value">{{ card.value }}</div>
                <div class="stat-card-label">{{ card.label }}</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-card>
          <div ref="tradeChartRef" style="height: 350px"></div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import dayjs from 'dayjs'
import { getDevices } from '@/api/device'
import { getOrders } from '@/api/order'
import { getSitesManage } from '@/api/site'
import { getWechatUsers } from '@/api/advertisement'

const activeTab = ref('business')

// 经营统计
const bizCards = reactive([
  { label: '用户总数', value: 0 },
  { label: '设备总数', value: 0 },
  { label: '站点总数', value: 0 },
  { label: '本月营收(元)', value: '0' },
])
const bizChartRef = ref<HTMLDivElement>()
let bizChart: echarts.ECharts | null = null

// 设备统计
const devCards = reactive([
  { label: '总设备数', value: 0 },
  { label: '在线设备', value: 0 },
  { label: '离线设备', value: 0 },
  { label: '故障设备', value: 0 },
])
const devStatusChartRef = ref<HTMLDivElement>()
const devSiteChartRef = ref<HTMLDivElement>()
let devStatusChart: echarts.ECharts | null = null
let devSiteChart: echarts.ECharts | null = null

// 充电汇总
const chargeCards = reactive([
  { label: '充电总次数', value: 0 },
  { label: '充电总时长(h)', value: '0' },
  { label: '充电总电量(kWh)', value: '0' },
  { label: '充电总金额(元)', value: '0' },
])
const chargeDateRange = ref<[string, string]>([
  dayjs().startOf('month').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
])
const chargeChartRef = ref<HTMLDivElement>()
let chargeChart: echarts.ECharts | null = null
const chargeRecords = ref<any[]>([])
const chargeLoading = ref(false)
const chargePage = ref(1)
const chargePageSize = ref(20)
const chargeTotal = ref(0)

// 交易汇总
const tradeCards = reactive([
  { label: '交易总笔数', value: 0 },
  { label: '交易总额(元)', value: '0' },
  { label: '退款笔数', value: 0 },
  { label: '退款总额(元)', value: '0' },
])
const tradeChartRef = ref<HTMLDivElement>()
let tradeChart: echarts.ECharts | null = null

// 缓存
let allDevices: any[] = []
let allOrders: any[] = []

async function initData() {
  try {
    const [dRes, oRes, sRes, uRes]: any[] = await Promise.all([
      getDevices({ page: 1, page_size: 10000 }),
      getOrders({ page: 1, page_size: 10000 }),
      getSitesManage({ page: 1, page_size: 1000 }),
      getWechatUsers({ page: 1, page_size: 10000 }),
    ])
    allDevices = dRes?.data || dRes?.list || []
    allOrders = oRes?.data || oRes?.list || []
    const sites = sRes?.data || sRes?.list || []
    const users = uRes?.data || []

    // 经营统计
    bizCards[0].value = users.length
    bizCards[1].value = allDevices.length
    bizCards[2].value = sites.length
    const now = dayjs()
    const monthOrders = allOrders.filter((o: any) => o.start_time && dayjs(o.start_time).isAfter(now.startOf('month')))
    const monthAmount = monthOrders.filter((o: any) => o.status === 'completed').reduce((s: number, o: any) => s + (o.paid_amount || 0), 0)
    bizCards[3].value = monthAmount.toFixed(2)

    // 设备统计
    devCards[0].value = allDevices.length
    devCards[1].value = allDevices.filter((d: any) => d.status === 'online').length
    devCards[2].value = allDevices.filter((d: any) => d.status === 'offline').length
    devCards[3].value = allDevices.filter((d: any) => d.status === 'fault').length

    // 充电汇总
    const completed = allOrders.filter((o: any) => o.status === 'completed' || o.status === 'refunded')
    chargeCards[0].value = completed.length
    chargeCards[1].value = (completed.reduce((s: number, o: any) => s + (o.duration_min || 0), 0) / 60).toFixed(1)
    chargeCards[2].value = completed.reduce((s: number, o: any) => s + (o.energy_kwh || 0), 0).toFixed(1)
    chargeCards[3].value = completed.reduce((s: number, o: any) => s + (o.amount || 0), 0).toFixed(2)

    // 交易汇总
    tradeCards[0].value = allOrders.length
    tradeCards[1].value = completed.reduce((s: number, o: any) => s + (o.amount || 0), 0).toFixed(2)
    const refunded = allOrders.filter((o: any) => o.status === 'refunded')
    tradeCards[2].value = refunded.length
    tradeCards[3].value = refunded.reduce((s: number, o: any) => s + (o.refund_amount || 0), 0).toFixed(2)

    await nextTick()
    renderBizChart()
  } catch {}
}

function renderBizChart() {
  if (!bizChartRef.value) return
  bizChart?.dispose()
  bizChart = echarts.init(bizChartRef.value)
  const months: string[] = []
  for (let i = 11; i >= 0; i--) months.push(dayjs().subtract(i, 'month').format('YYYY-MM'))
  const monthData = months.map(m => {
    const mOrders = allOrders.filter((o: any) => o.start_time && o.start_time.startsWith(m))
    const completed = mOrders.filter((o: any) => o.status === 'completed')
    return {
      orders: mOrders.length,
      amount: completed.reduce((s: number, o: any) => s + (o.amount || 0), 0),
      energy: completed.reduce((s: number, o: any) => s + (o.energy_kwh || 0), 0),
    }
  })
  bizChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['订单数', '交易额(元)', '用电量(kWh)'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: months.map(m => m.slice(5)) },
    yAxis: { type: 'value' },
    series: [
      { name: '订单数', type: 'bar', data: monthData.map(d => d.orders), itemStyle: { color: '#409eff' } },
      { name: '交易额(元)', type: 'bar', data: monthData.map(d => d.amount), itemStyle: { color: '#67c23a' } },
      { name: '用电量(kWh)', type: 'line', data: monthData.map(d => d.energy), itemStyle: { color: '#e6a23c' }, yAxisIndex: 1 },
    ],
  })
}

function renderDevCharts() {
  // 设备状态饼图
  if (devStatusChartRef.value) {
    devStatusChart?.dispose()
    devStatusChart = echarts.init(devStatusChartRef.value)
    devStatusChart.setOption({
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie', radius: ['45%', '70%'],
        data: [
          { value: devCards[1].value, name: '在线' },
          { value: devCards[2].value, name: '离线' },
          { value: devCards[3].value, name: '故障' },
        ],
      }],
    })
  }
  // 站点设备柱状图
  if (devSiteChartRef.value) {
    devSiteChart?.dispose()
    devSiteChart = echarts.init(devSiteChartRef.value)
    const siteMap: Record<string, number> = {}
    allDevices.forEach((d: any) => {
      const s = d.site_name || d.site_id || '未分配'
      siteMap[s] = (siteMap[s] || 0) + 1
    })
    const sites = Object.entries(siteMap).sort((a, b) => b[1] - a[1]).slice(0, 10)
    devSiteChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '10%', bottom: '3%', containLabel: true },
      xAxis: { type: 'value' },
      yAxis: { type: 'category', data: sites.map(s => s[0]).reverse(), axisLabel: { width: 100, overflow: 'truncate' } },
      series: [{ type: 'bar', data: sites.map(s => s[1]).reverse(), itemStyle: { color: '#409eff' } }],
    })
  }
}

async function loadChargeSummary() {
  chargeLoading.value = true
  try {
    const params: any = { page: chargePage.value, page_size: chargePageSize.value, status: 'completed' }
    if (chargeDateRange.value) {
      params.start_date = chargeDateRange.value[0]
      params.end_date = chargeDateRange.value[1]
    }
    const res: any = await getOrders(params)
    chargeRecords.value = res?.data || res?.list || []
    chargeTotal.value = res?.total || 0

    // 汇总
    const completed = res?.data || res?.list || []
    chargeCards[0].value = completed.length
    chargeCards[1].value = (completed.reduce((s: number, o: any) => s + (o.duration_min || 0), 0) / 60).toFixed(1)
    chargeCards[2].value = completed.reduce((s: number, o: any) => s + (o.energy_kwh || 0), 0).toFixed(1)
    chargeCards[3].value = completed.reduce((s: number, o: any) => s + (o.amount || 0), 0).toFixed(2)

    // 趋势
    if (chargeChartRef.value) {
      chargeChart?.dispose()
      chargeChart = echarts.init(chargeChartRef.value)
      const dailyMap: Record<string, { count: number; amount: number }> = {}
      completed.forEach((o: any) => {
        const day = (o.start_time || '').slice(0, 10)
        if (!day) return
        if (!dailyMap[day]) dailyMap[day] = { count: 0, amount: 0 }
        dailyMap[day].count++
        dailyMap[day].amount += o.amount || 0
      })
      const days = Object.keys(dailyMap).sort()
      chargeChart.setOption({
        tooltip: { trigger: 'axis' },
        legend: { data: ['充电次数', '充电金额(元)'] },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: days },
        yAxis: { type: 'value' },
        series: [
          { name: '充电次数', type: 'line', data: days.map(d => dailyMap[d].count), smooth: true, itemStyle: { color: '#409eff' } },
          { name: '充电金额(元)', type: 'line', data: days.map(d => dailyMap[d].amount), smooth: true, itemStyle: { color: '#67c23a' } },
        ],
      })
    }
  } catch {} finally {
    chargeLoading.value = false
  }
}

function renderTradeChart() {
  if (!tradeChartRef.value) return
  tradeChart?.dispose()
  tradeChart = echarts.init(tradeChartRef.value)
  const months: string[] = []
  for (let i = 5; i >= 0; i--) months.push(dayjs().subtract(i, 'month').format('YYYY-MM'))
  const mData = months.map(m => {
    const mOrders = allOrders.filter((o: any) => o.created_at && o.created_at.startsWith(m))
    const refunded = mOrders.filter((o: any) => o.status === 'refunded')
    return {
      total: mOrders.reduce((s: number, o: any) => s + (o.amount || 0), 0),
      refund: refunded.reduce((s: number, o: any) => s + (o.refund_amount || 0), 0),
    }
  })
  tradeChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['交易额', '退款额'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: months.map(m => m.slice(5)) },
    yAxis: { type: 'value' },
    series: [
      { name: '交易额', type: 'bar', data: mData.map(d => d.total), itemStyle: { color: '#409eff' } },
      { name: '退款额', type: 'bar', data: mData.map(d => d.refund), itemStyle: { color: '#f56c6c' } },
    ],
  })
}

function onTabChange(tab: string) {
  nextTick(() => {
    if (tab === 'device') renderDevCharts()
    else if (tab === 'charge') loadChargeSummary()
    else if (tab === 'trade') renderTradeChart()
  })
}

onMounted(initData)
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 8px 0;
}
.stat-card-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}
.stat-card-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}
</style>
