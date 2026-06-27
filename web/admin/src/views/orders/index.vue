<template>
  <div class="page-container">
    <div class="page-header">
      <h2>订单管理</h2>
    </div>

    <!-- 搜索筛选 -->
    <el-card style="margin-bottom: 16px">
      <el-form :inline="true" :model="query" size="default">
        <el-form-item label="订单编号">
          <el-input v-model="query.order_sn" placeholder="订单号" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="订单类型">
          <el-select v-model="query.order_type" placeholder="全部" clearable style="width: 140px">
            <el-option label="扫码充电" value="scan" />
            <el-option label="IC卡充电" value="ic_card" />
            <el-option label="免费充电" value="free" />
            <el-option label="包月充电" value="monthly" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable style="width: 140px">
            <el-option label="待支付" value="pending" />
            <el-option label="充电中" value="charging" />
            <el-option label="已完成" value="completed" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已退款" value="refunded" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">
            <el-icon><Search /></el-icon> 查询
          </el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 订单列表 -->
    <el-card>
      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="order_sn" label="订单编号" width="200" />
        <el-table-column prop="order_type" label="订单类型" width="120">
          <template #default="{ row }">
            <el-tag :type="orderTypeTag(row.order_type)" size="small">
              {{ orderTypeLabel(row.order_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="device_sn" label="设备SN" width="160" />
        <el-table-column prop="username" label="用户" width="120" />
        <el-table-column prop="amount" label="金额(元)" width="100" align="right">
          <template #default="{ row }">¥{{ (row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="energy_kwh" label="电量(kWh)" width="100" align="right">
          <template #default="{ row }">{{ (row.energy_kwh || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="duration_min" label="时长(分钟)" width="100" align="right" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="viewDetail(row)">
              <el-icon><View /></el-icon> 详情
            </el-button>
            <el-button v-if="row.status === 'charging'" type="warning" link size="small" @click="handleEnd(row)">
              <el-icon><VideoPause /></el-icon> 结束
            </el-button>
            <el-button v-if="row.status === 'completed'" type="danger" link size="small" @click="handleRefund(row)">
              <el-icon><Money /></el-icon> 退款
            </el-button>
            <el-button v-if="row.status === 'cancelled'" type="danger" link size="small" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon> 删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; text-align: right">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="search"
          @current-change="search"
        />
      </div>
    </el-card>

    <!-- 退款弹窗 -->
    <el-dialog v-model="refundDialog.visible" title="确认退款" width="420px">
      <el-form :model="refundDialog" label-width="80px">
        <el-form-item label="订单编号">
          <span>{{ refundDialog.order.order_sn }}</span>
        </el-form-item>
        <el-form-item label="已付金额">
          <span>¥{{ (refundDialog.order.amount || 0).toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="退款金额">
          <el-input-number v-model="refundDialog.amount" :min="0" :max="refundDialog.order.amount || 0" :precision="2" />
        </el-form-item>
        <el-form-item label="退款原因">
          <el-input v-model="refundDialog.reason" type="textarea" :rows="2" placeholder="请输入退款原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialog.visible = false">取消</el-button>
        <el-button type="danger" @click="doRefund" :loading="refunding">确认退款</el-button>
      </template>
    </el-dialog>

    <!-- 结束充电弹窗 -->
    <el-dialog v-model="endDialog.visible" title="结束充电" width="420px">
      <el-form label-width="80px">
        <el-form-item label="订单编号">
          <span>{{ endDialog.order.order_sn }}</span>
        </el-form-item>
        <el-form-item label="结算金额">
          <el-input-number v-model="endDialog.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="充电电量">
          <el-input-number v-model="endDialog.energy_kwh" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="endDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doEndOrder" :loading="ending">确认结束</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrders, endOrder, refundOrder, deleteOrder } from '@/api/order'

const router = useRouter()

const query = reactive<any>({
  order_sn: '',
  order_type: '',
  status: '',
  page: 1,
  page_size: 20,
})
const dateRange = ref<[string, string] | null>(null)
const list = ref<any[]>([])
const total = ref(0)
const loading = ref(false)

function orderTypeLabel(t: string) {
  const map: Record<string, string> = { scan: '扫码充电', ic_card: 'IC卡充电', free: '免费充电', monthly: '包月充电' }
  return map[t] || t
}
function orderTypeTag(t: string) {
  const map: Record<string, string> = { scan: '', ic_card: 'warning', free: 'success', monthly: 'info' }
  return map[t] || ''
}
function statusLabel(s: string) {
  const map: Record<string, string> = { pending: '待支付', charging: '充电中', completed: '已完成', cancelled: '已取消', refunded: '已退款' }
  return map[s] || s
}
function statusTag(s: string) {
  const map: Record<string, string> = { pending: 'info', charging: 'warning', completed: 'success', cancelled: 'info', refunded: 'danger' }
  return map[s] || ''
}

async function search() {
  loading.value = true
  try {
    const params: any = { ...query }
    if (dateRange.value) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }
    const res: any = await getOrders(params)
    list.value = res?.data || res?.list || []
    total.value = res?.total || 0
  } catch { /* handled by interceptor */ } finally {
    loading.value = false
  }
}

function reset() {
  query.order_sn = ''
  query.order_type = ''
  query.status = ''
  dateRange.value = null
  query.page = 1
  search()
}

function viewDetail(row: any) {
  router.push(`/orders/${row.id}`)
}

// 结束充电
const endDialog = reactive({ visible: false, order: {} as any, amount: 0, energy_kwh: 0 })
const ending = ref(false)
function handleEnd(row: any) {
  endDialog.order = row
  endDialog.amount = row.amount || 0
  endDialog.energy_kwh = row.energy_kwh || 0
  endDialog.visible = true
}
async function doEndOrder() {
  ending.value = true
  try {
    await endOrder(endDialog.order.id, { amount: endDialog.amount, energy_kwh: endDialog.energy_kwh })
    ElMessage.success('充电已结束')
    endDialog.visible = false
    search()
  } catch { /* handled */ } finally {
    ending.value = false
  }
}

// 退款
const refundDialog = reactive({ visible: false, order: {} as any, amount: 0, reason: '' })
const refunding = ref(false)
function handleRefund(row: any) {
  refundDialog.order = row
  refundDialog.amount = row.amount || 0
  refundDialog.reason = ''
  refundDialog.visible = true
}
async function doRefund() {
  if (!refundDialog.reason.trim()) {
    ElMessage.warning('请输入退款原因')
    return
  }
  refunding.value = true
  try {
    await refundOrder(refundDialog.order.id, { amount: refundDialog.amount, reason: refundDialog.reason })
    ElMessage.success('退款申请已提交')
    refundDialog.visible = false
    search()
  } catch { /* handled */ } finally {
    refunding.value = false
  }
}

// 删除
function handleDelete(row: any) {
  ElMessageBox.confirm('确认删除该订单？此操作不可恢复！', '删除确认', {
    confirmButtonText: '确认删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    await deleteOrder(String(row.id))
    ElMessage.success('订单已删除')
    search()
  }).catch(() => {})
}

// 挂载时查询
search()
</script>
