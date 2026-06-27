<template>
  <div class="page-container">
    <div class="page-header">
      <h2>财务管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 充值记录 -->
      <el-tab-pane label="充值记录" name="recharge">
        <el-card>
          <el-form :inline="true" :model="rechargeQuery">
            <el-form-item label="用户ID">
              <el-input v-model="rechargeQuery.user_id" placeholder="用户ID" clearable style="width: 140px" />
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker v-model="rechargeDateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 260px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchRecharges">查询</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="recharges" stripe v-loading="rechargeLoading">
            <el-table-column prop="user_id" label="用户ID" width="100" />
            <el-table-column prop="user_name" label="用户名" width="140" />
            <el-table-column prop="amount" label="充值金额(元)" width="130" align="right">
              <template #default="{ row }">¥{{ (row.amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="bonus_amount" label="赠送金额(元)" width="130" align="right">
              <template #default="{ row }">¥{{ (row.bonus_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="pay_method" label="支付方式" width="120" />
            <el-table-column prop="trade_no" label="交易单号" width="200" show-overflow-tooltip />
            <el-table-column prop="created_at" label="充值时间" width="170" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 提现管理 -->
      <el-tab-pane label="提现管理" name="withdraw">
        <el-card>
          <el-table :data="withdraws" stripe v-loading="withdrawLoading">
            <el-table-column prop="user_id" label="用户ID" width="100" />
            <el-table-column prop="user_name" label="用户名" width="140" />
            <el-table-column prop="amount" label="提现金额(元)" width="130" align="right">
              <template #default="{ row }">¥{{ (row.amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="bank_name" label="银行" width="120" />
            <el-table-column prop="bank_card_no" label="银行卡号" width="180" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'" size="small">
                  {{ { pending: '待审核', approved: '已打款', rejected: '已驳回' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
            <el-table-column prop="created_at" label="申请时间" width="170" />
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button v-if="row.status === 'pending'" type="success" link size="small" @click="handleWithdraw(row, 'approved')">打款</el-button>
                <el-button v-if="row.status === 'pending'" type="danger" link size="small" @click="handleWithdraw(row, 'rejected')">驳回</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 分成记录 -->
      <el-tab-pane label="分成记录" name="splits">
        <el-card>
          <el-table :data="splits" stripe v-loading="splitsLoading">
            <el-table-column prop="order_sn" label="订单编号" width="200" />
            <el-table-column prop="total_amount" label="订单金额(元)" width="120" align="right">
              <template #default="{ row }">¥{{ (row.total_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="platform_amount" label="平台(元)" width="100" align="right">
              <template #default="{ row }">¥{{ (row.platform_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="agent_amount" label="代理(元)" width="100" align="right">
              <template #default="{ row }">¥{{ (row.agent_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="operator_amount" label="运营商(元)" width="110" align="right">
              <template #default="{ row }">¥{{ (row.operator_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'settled' ? 'success' : 'warning'" size="small">
                  {{ row.status === 'settled' ? '已结算' : '待结算' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="settled_at" label="结算时间" width="170" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getRecharges, getWithdraws, processWithdraw, getSplits } from '@/api/finance'

const activeTab = ref('recharge')

// 充值记录
const recharges = ref<any[]>([])
const rechargeLoading = ref(false)
const rechargeQuery = reactive({ user_id: '' })
const rechargeDateRange = ref<[string, string] | null>(null)

// 提现管理
const withdraws = ref<any[]>([])
const withdrawLoading = ref(false)

// 分成
const splits = ref<any[]>([])
const splitsLoading = ref(false)

async function fetchRecharges() {
  rechargeLoading.value = true
  try {
    const params: any = { ...rechargeQuery }
    if (rechargeDateRange.value) {
      params.start_date = rechargeDateRange.value[0]
      params.end_date = rechargeDateRange.value[1]
    }
    const res: any = await getRecharges(params)
    recharges.value = res?.data || res?.list || []
  } catch {} finally {
    rechargeLoading.value = false
  }
}

async function fetchWithdraws() {
  withdrawLoading.value = true
  try {
    const res: any = await getWithdraws()
    withdraws.value = res?.data || res?.list || []
  } catch {} finally {
    withdrawLoading.value = false
  }
}

async function handleWithdraw(row: any, status: string) {
  try {
    await processWithdraw(row.id, { status })
    ElMessage.success(status === 'approved' ? '已打款' : '已驳回')
    fetchWithdraws()
  } catch {}
}

async function fetchSplits() {
  splitsLoading.value = true
  try {
    const res: any = await getSplits()
    splits.value = res?.data || res?.list || []
  } catch {} finally {
    splitsLoading.value = false
  }
}

onMounted(fetchRecharges)
</script>
