<template>
  <div class="page-container">
    <div class="page-header">
      <h2>代理商与运营商</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 代理商 -->
      <el-tab-pane label="代理商管理" name="agents">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openAgentDialog()">
            <el-icon><Plus /></el-icon> 新增代理商
          </el-button>
        </div>
        <el-card>
          <el-table :data="agents" stripe v-loading="agentLoading">
            <el-table-column prop="name" label="代理商名称" width="180" />
            <el-table-column prop="contact_name" label="联系人" width="120" />
            <el-table-column prop="contact_phone" label="联系电话" width="140" />
            <el-table-column prop="commission_rate" label="分佣比例(%)" width="110" align="center">
              <template #default="{ row }">{{ row.commission_rate || 0 }}%</template>
            </el-table-column>
            <el-table-column prop="total_revenue" label="累计收益(元)" width="130" align="right">
              <template #default="{ row }">¥{{ (row.total_revenue || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="withdrawable" label="可提现(元)" width="120" align="right">
              <template #default="{ row }">¥{{ (row.withdrawable || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openAgentDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteAgent(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 运营商 -->
      <el-tab-pane label="运营商管理" name="operators">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openOperatorDialog()">
            <el-icon><Plus /></el-icon> 新增运营商
          </el-button>
        </div>
        <el-card>
          <el-table :data="operators" stripe v-loading="operatorLoading">
            <el-table-column prop="name" label="运营商名称" width="180" />
            <el-table-column prop="contact_name" label="联系人" width="120" />
            <el-table-column prop="contact_phone" label="联系电话" width="140" />
            <el-table-column prop="site_name" label="关联站点" width="150" />
            <el-table-column prop="commission_rate" label="分佣比例(%)" width="110" align="center">
              <template #default="{ row }">{{ row.commission_rate || 0 }}%</template>
            </el-table-column>
            <el-table-column prop="total_revenue" label="累计收益(元)" width="130" align="right">
              <template #default="{ row }">¥{{ (row.total_revenue || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openOperatorDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteOperator(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 代理商弹窗 -->
    <el-dialog v-model="agentDialog.visible" :title="agentDialog.isEdit ? '编辑代理商' : '新增代理商'" width="500px">
      <el-form :model="agentDialog.form" label-width="100px">
        <el-form-item label="代理商名称">
          <el-input v-model="agentDialog.form.name" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="agentDialog.form.contact_name" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="agentDialog.form.contact_phone" />
        </el-form-item>
        <el-form-item label="分佣比例(%)">
          <el-input-number v-model="agentDialog.form.commission_rate" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="agentDialog.form.address" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="agentDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveAgent" :loading="agentSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 运营商弹窗 -->
    <el-dialog v-model="operatorDialog.visible" :title="operatorDialog.isEdit ? '编辑运营商' : '新增运营商'" width="500px">
      <el-form :model="operatorDialog.form" label-width="100px">
        <el-form-item label="运营商名称">
          <el-input v-model="operatorDialog.form.name" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="operatorDialog.form.contact_name" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="operatorDialog.form.contact_phone" />
        </el-form-item>
        <el-form-item label="关联站点ID">
          <el-input v-model="operatorDialog.form.site_id" placeholder="关联的站点ID" />
        </el-form-item>
        <el-form-item label="分佣比例(%)">
          <el-input-number v-model="operatorDialog.form.commission_rate" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="operatorDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveOperator" :loading="operatorSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAgents, createAgent, updateAgent, deleteAgent } from '@/api/agent'
import { getOperators, createOperator, updateOperator, deleteOperator } from '@/api/agent'

const activeTab = ref('agents')

// 代理商
const agents = ref<any[]>([])
const agentLoading = ref(false)
const agentDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const agentSaving = ref(false)

// 运营商
const operators = ref<any[]>([])
const operatorLoading = ref(false)
const operatorDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const operatorSaving = ref(false)

// 代理商 CRUD
async function fetchAgents() {
  agentLoading.value = true
  try {
    const res: any = await getAgents()
    agents.value = res?.data || res?.list || []
  } catch {} finally {
    agentLoading.value = false
  }
}

function openAgentDialog(row?: any) {
  agentDialog.isEdit = !!row
  agentDialog.form = row ? { ...row } : { name: '', contact_name: '', contact_phone: '', commission_rate: 10, address: '' }
  agentDialog.visible = true
}

async function doSaveAgent() {
  agentSaving.value = true
  try {
    if (agentDialog.isEdit) {
      await updateAgent(agentDialog.form.id, agentDialog.form)
      ElMessage.success('代理商已更新')
    } else {
      await createAgent(agentDialog.form)
      ElMessage.success('代理商已创建')
    }
    agentDialog.visible = false
    fetchAgents()
  } catch {} finally {
    agentSaving.value = false
  }
}

function handleDeleteAgent(row: any) {
  ElMessageBox.confirm('确认删除该代理商？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteAgent(row.id)
    ElMessage.success('已删除')
    fetchAgents()
  }).catch(() => {})
}

// 运营商 CRUD
async function fetchOperators() {
  operatorLoading.value = true
  try {
    const res: any = await getOperators()
    operators.value = res?.data || res?.list || []
  } catch {} finally {
    operatorLoading.value = false
  }
}

function openOperatorDialog(row?: any) {
  operatorDialog.isEdit = !!row
  operatorDialog.form = row ? { ...row } : { name: '', contact_name: '', contact_phone: '', site_id: '', commission_rate: 5 }
  operatorDialog.visible = true
}

async function doSaveOperator() {
  operatorSaving.value = true
  try {
    if (operatorDialog.isEdit) {
      await updateOperator(operatorDialog.form.id, operatorDialog.form)
      ElMessage.success('运营商已更新')
    } else {
      await createOperator(operatorDialog.form)
      ElMessage.success('运营商已创建')
    }
    operatorDialog.visible = false
    fetchOperators()
  } catch {} finally {
    operatorSaving.value = false
  }
}

function handleDeleteOperator(row: any) {
  ElMessageBox.confirm('确认删除该运营商？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteOperator(row.id)
    ElMessage.success('已删除')
    fetchOperators()
  }).catch(() => {})
}

onMounted(() => {
  fetchAgents()
  fetchOperators()
})
</script>
