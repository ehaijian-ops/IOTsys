<template>
  <div class="page-container">
    <div class="page-header">
      <h2>卡片管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- IC充电卡 -->
      <el-tab-pane label="IC充电卡" name="ic">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openICDialog()"><el-icon><Plus /></el-icon> 新增IC卡</el-button>
          <el-upload
            :show-file-list="false"
            :before-upload="handleICImport"
            accept=".csv,.xlsx"
            style="display: inline-block; margin-left: 12px"
          >
            <el-button><el-icon><Upload /></el-icon> 批量导入</el-button>
          </el-upload>
        </div>
        <el-card>
          <el-table :data="icCards" stripe v-loading="icLoading">
            <el-table-column prop="card_no" label="卡号" width="200" />
            <el-table-column prop="card_type" label="卡片类型" width="120">
              <template #default="{ row }">{{ { charge: '充电卡', member: '会员卡' }[row.card_type] || row.card_type }}</template>
            </el-table-column>
            <el-table-column prop="balance" label="余额(元)" width="120" align="right">
              <template #default="{ row }">¥{{ (row.balance || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="user_name" label="绑定用户" width="140" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : row.status === 'lost' ? 'danger' : 'info'" size="small">
                  {{ { active: '正常', lost: '已挂失', disabled: '已禁用' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openRecharge(row)">充值</el-button>
                <el-button type="warning" link size="small" @click="handleLost(row)" v-if="row.status === 'active'">挂失</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteIC(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 流量卡 -->
      <el-tab-pane label="流量卡" name="traffic">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openTCDialog()"><el-icon><Plus /></el-icon> 新增流量卡</el-button>
        </div>
        <el-card>
          <el-table :data="trafficCards" stripe>
            <el-table-column prop="iccid" label="ICCID" width="220" />
            <el-table-column prop="imsi" label="IMSI" width="180" />
            <el-table-column prop="phone_no" label="手机号" width="140" />
            <el-table-column prop="operator" label="运营商" width="100">
              <template #default="{ row }">{{ { cmcc: '移动', cucc: '联通', ctcc: '电信' }[row.operator] || row.operator }}</template>
            </el-table-column>
            <el-table-column prop="total_flow_mb" label="总流量(MB)" width="120" align="right" />
            <el-table-column prop="used_flow_mb" label="已用(MB)" width="100" align="right" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '在用' : '停机' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openTCDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteTC(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 月卡记录 -->
      <el-tab-pane label="月卡记录" name="monthly">
        <el-card>
          <el-table :data="monthlyCards" stripe>
            <el-table-column prop="user_name" label="用户" width="140" />
            <el-table-column prop="scheme_name" label="套餐方案" width="200" />
            <el-table-column prop="start_date" label="开始日期" width="120" />
            <el-table-column prop="end_date" label="到期日期" width="120" />
            <el-table-column prop="used_charges" label="已充电次数" width="120" align="center" />
            <el-table-column prop="used_energy_kwh" label="已用电量(kWh)" width="140" align="right" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : row.status === 'expired' ? 'info' : 'danger'" size="small">
                  {{ { active: '生效中', expired: '已过期', exhausted: '已用完', cancelled: '已取消' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- IC卡弹窗 -->
    <el-dialog v-model="icDialog.visible" :title="icDialog.isEdit ? '编辑IC卡' : '新增IC卡'" width="480px">
      <el-form :model="icDialog.form" label-width="80px">
        <el-form-item label="卡号">
          <el-input v-model="icDialog.form.card_no" placeholder="IC卡物理卡号" :disabled="icDialog.isEdit" />
        </el-form-item>
        <el-form-item label="卡片类型">
          <el-select v-model="icDialog.form.card_type" style="width: 100%">
            <el-option label="充电卡" value="charge" />
            <el-option label="会员卡" value="member" />
          </el-select>
        </el-form-item>
        <el-form-item label="初始余额">
          <el-input-number v-model="icDialog.form.balance" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="icDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveIC" :loading="icSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- IC卡充值弹窗 -->
    <el-dialog v-model="rechargeDialog.visible" title="IC卡充值" width="420px">
      <el-form label-width="80px">
        <el-form-item label="卡号">
          <span>{{ rechargeDialog.card.card_no }}</span>
        </el-form-item>
        <el-form-item label="当前余额">
          <span>¥{{ (rechargeDialog.card.balance || 0).toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="充值金额">
          <el-input-number v-model="rechargeDialog.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="rechargeDialog.remark" placeholder="充值备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doRechargeIC" :loading="recharging">确认充值</el-button>
      </template>
    </el-dialog>

    <!-- 流量卡弹窗 -->
    <el-dialog v-model="tcDialog.visible" :title="tcDialog.isEdit ? '编辑流量卡' : '新增流量卡'" width="480px">
      <el-form :model="tcDialog.form" label-width="80px">
        <el-form-item label="ICCID">
          <el-input v-model="tcDialog.form.iccid" :disabled="tcDialog.isEdit" />
        </el-form-item>
        <el-form-item label="IMSI">
          <el-input v-model="tcDialog.form.imsi" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="tcDialog.form.phone_no" />
        </el-form-item>
        <el-form-item label="运营商">
          <el-select v-model="tcDialog.form.operator" style="width: 100%">
            <el-option label="移动" value="cmcc" />
            <el-option label="联通" value="cucc" />
            <el-option label="电信" value="ctcc" />
          </el-select>
        </el-form-item>
        <el-form-item label="总流量(MB)">
          <el-input-number v-model="tcDialog.form.total_flow_mb" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tcDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveTC" :loading="tcSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getICCards, createICCard, updateICCard, deleteICCard, rechargeICCard,
  getTrafficCards, createTrafficCard, updateTrafficCard, deleteTrafficCard,
  getMonthlyCards,
} from '@/api/card'

const activeTab = ref('ic')

// IC卡
const icCards = ref<any[]>([])
const icLoading = ref(false)
const icDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const icSaving = ref(false)
const rechargeDialog = reactive({ visible: false, card: {} as any, amount: 0, remark: '' })
const recharging = ref(false)

// 流量卡
const trafficCards = ref<any[]>([])
const tcDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const tcSaving = ref(false)

// 月卡
const monthlyCards = ref<any[]>([])

async function fetchIC() {
  icLoading.value = true
  try {
    const res: any = await getICCards()
    icCards.value = res?.data || res?.list || []
  } catch {} finally {
    icLoading.value = false
  }
}

function openICDialog(row?: any) {
  icDialog.isEdit = !!row
  icDialog.form = row ? { ...row } : { card_no: '', card_type: 'charge', balance: 0 }
  icDialog.visible = true
}

async function doSaveIC() {
  icSaving.value = true
  try {
    if (icDialog.isEdit) {
      await updateICCard(icDialog.form.id, icDialog.form)
      ElMessage.success('IC卡已更新')
    } else {
      await createICCard(icDialog.form)
      ElMessage.success('IC卡已创建')
    }
    icDialog.visible = false
    fetchIC()
  } catch {} finally {
    icSaving.value = false
  }
}

function openRecharge(row: any) {
  rechargeDialog.card = row
  rechargeDialog.amount = 0
  rechargeDialog.remark = ''
  rechargeDialog.visible = true
}

async function doRechargeIC() {
  if (rechargeDialog.amount <= 0) {
    ElMessage.warning('请输入充值金额')
    return
  }
  recharging.value = true
  try {
    await rechargeICCard(rechargeDialog.card.id, { amount: rechargeDialog.amount, remark: rechargeDialog.remark })
    ElMessage.success('充值成功')
    rechargeDialog.visible = false
    fetchIC()
  } catch {} finally {
    recharging.value = false
  }
}

async function handleLost(row: any) {
  try {
    await updateICCard(row.id, { ...row, status: 'lost' })
    ElMessage.success('已挂失')
    fetchIC()
  } catch {}
}

function handleDeleteIC(row: any) {
  ElMessageBox.confirm('确认删除该IC卡？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteICCard(row.id)
    ElMessage.success('已删除')
    fetchIC()
  }).catch(() => {})
}

function handleICImport(file: File) {
  ElMessage.info(`文件导入功能: ${file.name}, 请通过API批量导入`)
  return false
}

// 流量卡
async function fetchTC() {
  try {
    const res: any = await getTrafficCards()
    trafficCards.value = res?.data || res?.list || []
  } catch {}
}

function openTCDialog(row?: any) {
  tcDialog.isEdit = !!row
  tcDialog.form = row ? { ...row } : { iccid: '', imsi: '', phone_no: '', operator: 'cmcc', total_flow_mb: 1024 }
  tcDialog.visible = true
}

async function doSaveTC() {
  tcSaving.value = true
  try {
    if (tcDialog.isEdit) {
      await updateTrafficCard(tcDialog.form.id, tcDialog.form)
    } else {
      await createTrafficCard(tcDialog.form)
    }
    ElMessage.success('流量卡已保存')
    tcDialog.visible = false
    fetchTC()
  } catch {} finally {
    tcSaving.value = false
  }
}

function handleDeleteTC(row: any) {
  ElMessageBox.confirm('确认删除该流量卡？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteTrafficCard(row.id)
    ElMessage.success('已删除')
    fetchTC()
  }).catch(() => {})
}

// 月卡
async function fetchMonthly() {
  try {
    const res: any = await getMonthlyCards()
    monthlyCards.value = res?.data || res?.list || []
  } catch {}
}

onMounted(() => {
  fetchIC()
  fetchTC()
  fetchMonthly()
})
</script>
