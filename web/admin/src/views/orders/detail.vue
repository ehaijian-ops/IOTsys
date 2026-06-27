<template>
  <div class="page-container">
    <div class="page-header">
      <el-button @click="$router.push('/orders')" style="margin-right: 12px">
        <el-icon><ArrowLeft /></el-icon> 返回
      </el-button>
      <h2>订单详情</h2>
    </div>

    <!-- 订单信息 -->
    <el-card style="margin-bottom: 16px">
      <template #header><span>订单信息</span></template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="订单编号">{{ order.order_sn }}</el-descriptions-item>
        <el-descriptions-item label="订单类型">
          <el-tag :type="orderTypeTag(order.order_type)" size="small">{{ orderTypeLabel(order.order_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTag(order.status)" size="small">{{ statusLabel(order.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="设备SN">{{ order.device_sn || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备类型">{{ order.device_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="站点">{{ order.site_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ order.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ (order.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实际支付">¥{{ (order.paid_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="退款金额">¥{{ (order.refund_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="充电电量">{{ (order.energy_kwh || 0).toFixed(2) }} kWh</el-descriptions-item>
        <el-descriptions-item label="充电时长">{{ order.duration_min || 0 }} 分钟</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ order.start_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ order.end_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ order.created_at || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 充电曲线 -->
    <el-card style="margin-bottom: 16px" v-if="curveData.length > 0">
      <template #header><span>充电功率曲线</span></template>
      <div ref="chartRef" style="height: 300px"></div>
    </el-card>

    <!-- 操作按钮 -->
    <el-card>
      <el-space>
        <el-button v-if="order.status === 'charging'" type="warning" @click="showEndDialog = true">
          结束充电
        </el-button>
        <el-button v-if="order.status === 'completed'" type="danger" @click="showRefundDialog = true">
          退款
        </el-button>
      </el-space>
    </el-card>

    <!-- 结束充电弹窗 -->
    <el-dialog v-model="showEndDialog" title="结束充电" width="420px">
      <el-form label-width="80px">
        <el-form-item label="结算金额">
          <el-input-number v-model="endForm.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="充电电量">
          <el-input-number v-model="endForm.energy_kwh" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEndDialog = false">取消</el-button>
        <el-button type="primary" @click="doEndOrder" :loading="ending">确认结束</el-button>
      </template>
    </el-dialog>

    <!-- 退款弹窗 -->
    <el-dialog v-model="showRefundDialog" title="确认退款" width="420px">
      <el-form label-width="80px">
        <el-form-item label="已付金额">
          <span>¥{{ (order.amount || 0).toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="退款金额">
          <el-input-number v-model="refundForm.amount" :min="0" :max="order.amount || 0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="退款原因">
          <el-input v-model="refundForm.reason" type="textarea" :rows="2" placeholder="请输入退款原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRefundDialog = false">取消</el-button>
        <el-button type="danger" @click="doRefund" :loading="refunding">确认退款</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getOrder, getChargeCurve, endOrder, refundOrder } from '@/api/order'

const route = useRoute()
const order = ref<any>({})
const curveData = ref<any[]>([])
const chartRef = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null

const showEndDialog = ref(false)
const endForm = reactive({ amount: 0, energy_kwh: 0 })
const ending = ref(false)
const showRefundDialog = ref(false)
const refundForm = reactive({ amount: 0, reason: '' })
const refunding = ref(false)

function orderTypeLabel(t: string) {
  const m: Record<string, string> = { scan: '扫码充电', ic_card: 'IC卡充电', free: '免费充电', monthly: '包月充电' }
  return m[t] || t
}
function orderTypeTag(t: string) {
  const m: Record<string, string> = { scan: '', ic_card: 'warning', free: 'success', monthly: 'info' }
  return m[t] || ''
}
function statusLabel(s: string) {
  const m: Record<string, string> = { pending: '待支付', charging: '充电中', completed: '已完成', cancelled: '已取消', refunded: '已退款' }
  return m[s] || s
}
function statusTag(s: string) {
  const m: Record<string, string> = { pending: 'info', charging: 'warning', completed: 'success', cancelled: 'info', refunded: 'danger' }
  return m[s] || ''
}

async function fetchDetail() {
  const id = route.params.id as string
  try {
    const [orderRes, curveRes]: any[] = await Promise.all([
      getOrder(id),
      getChargeCurve(id).catch(() => ({ data: [] })),
    ])
    order.value = orderRes?.data || orderRes || {}
    curveData.value = curveRes?.data || curveRes?.list || []

    // 如果有充电曲线且已完成，渲染图表
    if (curveData.value.length > 0) {
      await nextTick()
      renderChart()
    }

    // 回填结束/退款表单
    endForm.amount = order.value.amount || 0
    endForm.energy_kwh = order.value.energy_kwh || 0
    refundForm.amount = order.value.amount || 0
  } catch {}
}

function renderChart() {
  if (!chartRef.value) return
  chart?.dispose()
  chart = echarts.init(chartRef.value)
  const times = curveData.value.map((r: any) => r.record_time || r.created_at || '')
  const powers = curveData.value.map((r: any) => r.power_kw || r.power || 0)
  const voltages = curveData.value.map((r: any) => r.voltage || r.voltage_v || 0)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['功率(kW)', '电压(V)'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: times.map((t: string) => t.slice(11, 19) || t) },
    yAxis: { type: 'value' },
    series: [
      { name: '功率(kW)', type: 'line', data: powers, smooth: true, itemStyle: { color: '#409eff' } },
      { name: '电压(V)', type: 'line', data: voltages, smooth: true, itemStyle: { color: '#67c23a' } },
    ],
  })
}

async function doEndOrder() {
  ending.value = true
  try {
    await endOrder(order.value.id, { amount: endForm.amount, energy_kwh: endForm.energy_kwh })
    ElMessage.success('充电已结束')
    showEndDialog.value = false
    fetchDetail()
  } catch {} finally {
    ending.value = false
  }
}

async function doRefund() {
  if (!refundForm.reason.trim()) {
    ElMessage.warning('请输入退款原因')
    return
  }
  refunding.value = true
  try {
    await refundOrder(order.value.id, { amount: refundForm.amount, reason: refundForm.reason })
    ElMessage.success('退款申请已提交')
    showRefundDialog.value = false
    fetchDetail()
  } catch {} finally {
    refunding.value = false
  }
}

onMounted(fetchDetail)
</script>
