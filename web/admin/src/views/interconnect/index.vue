<template>
  <div class="page-container">
    <div class="page-header">
      <h2>互联互通</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 机构管理 -->
      <el-tab-pane label="机构管理" name="orgs">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openOrgDialog()">
            <el-icon><Plus /></el-icon> 新增机构
          </el-button>
        </div>
        <el-card>
          <el-table :data="orgs" stripe v-loading="orgLoading">
            <el-table-column prop="name" label="机构名称" width="200" />
            <el-table-column prop="code" label="机构编码" width="150" />
            <el-table-column prop="contact_name" label="联系人" width="120" />
            <el-table-column prop="contact_phone" label="联系电话" width="140" />
            <el-table-column prop="push_url" label="推送URL" min-width="240" show-overflow-tooltip />
            <el-table-column prop="reconcile_url" label="对账URL" min-width="240" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '已激活' : '已禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openOrgDialog(row)">编辑</el-button>
                <el-button type="success" link size="small" @click="testPush(row)">测试推送</el-button>
                <el-button type="warning" link size="small" @click="syncReconcile(row)">对账</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteOrg(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 密钥管理 -->
      <el-tab-pane label="密钥管理" name="keys">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openKeyDialog()">
            <el-icon><Plus /></el-icon> 新增密钥
          </el-button>
        </div>
        <el-card>
          <el-table :data="keys" stripe v-loading="keyLoading">
            <el-table-column prop="org_name" label="所属机构" width="180" />
            <el-table-column prop="key_type" label="密钥类型" width="140">
              <template #default="{ row }">
                <el-tag size="small">{{ { org_key: '机构密钥', sign_key: '签名密钥', msg_key: '消息密钥' }[row.key_type] || row.key_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="key_name" label="密钥名称" width="180" />
            <el-table-column prop="key_value" label="密钥值" min-width="300" show-overflow-tooltip>
              <template #default="{ row }">
                <el-input :model-value="row.key_value" readonly size="small" style="width: 100%">
                  <template #suffix>
                    <el-button link size="small" @click="copyKey(row.key_value)">复制</el-button>
                  </template>
                </el-input>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'active' ? '有效' : '已失效' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openKeyDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteKey(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 机构弹窗 -->
    <el-dialog v-model="orgDialog.visible" :title="orgDialog.isEdit ? '编辑机构' : '新增机构'" width="560px">
      <el-form :model="orgDialog.form" label-width="100px">
        <el-form-item label="机构名称">
          <el-input v-model="orgDialog.form.name" placeholder="如：星星充电" />
        </el-form-item>
        <el-form-item label="机构编码">
          <el-input v-model="orgDialog.form.code" placeholder="唯一编码" :disabled="orgDialog.isEdit" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="orgDialog.form.contact_name" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="orgDialog.form.contact_phone" />
        </el-form-item>
        <el-form-item label="推送URL">
          <el-input v-model="orgDialog.form.push_url" placeholder="https://partner.example.com/api/charge/push" />
        </el-form-item>
        <el-form-item label="对账URL">
          <el-input v-model="orgDialog.form.reconcile_url" placeholder="https://partner.example.com/api/charge/reconcile" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="orgDialog.form.status" active-value="active" inactive-value="disabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="orgDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveOrg" :loading="orgSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 密钥弹窗 -->
    <el-dialog v-model="keyDialog.visible" :title="keyDialog.isEdit ? '编辑密钥' : '新增密钥'" width="520px">
      <el-form :model="keyDialog.form" label-width="100px">
        <el-form-item label="所属机构">
          <el-select v-model="keyDialog.form.org_id" placeholder="选择机构" style="width: 100%" :disabled="keyDialog.isEdit">
            <el-option v-for="org in orgs" :key="org.id" :label="org.name" :value="org.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="密钥类型">
          <el-select v-model="keyDialog.form.key_type" style="width: 100%" :disabled="keyDialog.isEdit">
            <el-option label="机构密钥" value="org_key" />
            <el-option label="签名密钥" value="sign_key" />
            <el-option label="消息密钥" value="msg_key" />
          </el-select>
        </el-form-item>
        <el-form-item label="密钥名称">
          <el-input v-model="keyDialog.form.key_name" />
        </el-form-item>
        <el-form-item label="密钥值">
          <el-input v-model="keyDialog.form.key_value" :disabled="keyDialog.isEdit">
            <template #append>
              <el-button @click="keyDialog.form.key_value = generateRandomKey()">生成</el-button>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="keyDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveKey" :loading="keySaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getInterconnectOrgs, createInterconnectOrg, updateInterconnectOrg, deleteInterconnectOrg } from '@/api/interconnect'
import { getInterconnectKeys, createInterconnectKey, updateInterconnectKey, deleteInterconnectKey } from '@/api/interconnect'

const activeTab = ref('orgs')

// 机构
const orgs = ref<any[]>([])
const orgLoading = ref(false)
const orgDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const orgSaving = ref(false)

async function fetchOrgs() {
  orgLoading.value = true
  try {
    const res: any = await getInterconnectOrgs()
    orgs.value = res?.data || res?.list || []
  } catch {} finally {
    orgLoading.value = false
  }
}

function openOrgDialog(row?: any) {
  orgDialog.isEdit = !!row
  orgDialog.form = row ? { ...row } : { name: '', code: '', contact_name: '', contact_phone: '', push_url: '', reconcile_url: '', status: 'active' }
  orgDialog.visible = true
}

async function doSaveOrg() {
  orgSaving.value = true
  try {
    if (orgDialog.isEdit) {
      await updateInterconnectOrg(orgDialog.form.id, orgDialog.form)
      ElMessage.success('机构已更新')
    } else {
      await createInterconnectOrg(orgDialog.form)
      ElMessage.success('机构已创建')
    }
    orgDialog.visible = false
    fetchOrgs()
  } catch {} finally {
    orgSaving.value = false
  }
}

function handleDeleteOrg(row: any) {
  ElMessageBox.confirm('确认删除该机构？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteInterconnectOrg(row.id)
    ElMessage.success('已删除')
    fetchOrgs()
  }).catch(() => {})
}

function testPush(row: any) {
  ElMessage.info(`正在测试推送至: ${row.name}`)
}

function syncReconcile(row: any) {
  ElMessage.info(`正在与 ${row.name} 进行对账...`)
}

// 密钥
const keys = ref<any[]>([])
const keyLoading = ref(false)
const keyDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const keySaving = ref(false)

async function fetchKeys() {
  keyLoading.value = true
  try {
    const res: any = await getInterconnectKeys()
    keys.value = res?.data || res?.list || []
  } catch {} finally {
    keyLoading.value = false
  }
}

function openKeyDialog(row?: any) {
  keyDialog.isEdit = !!row
  keyDialog.form = row ? { ...row } : { org_id: '', key_type: 'sign_key', key_name: '', key_value: '' }
  keyDialog.visible = true
}

async function doSaveKey() {
  keySaving.value = true
  try {
    if (keyDialog.isEdit) {
      await updateInterconnectKey(keyDialog.form.id, keyDialog.form)
      ElMessage.success('密钥已更新')
    } else {
      await createInterconnectKey(keyDialog.form)
      ElMessage.success('密钥已创建')
    }
    keyDialog.visible = false
    fetchKeys()
  } catch {} finally {
    keySaving.value = false
  }
}

function handleDeleteKey(row: any) {
  ElMessageBox.confirm('确认删除该密钥？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteInterconnectKey(row.id)
    ElMessage.success('已删除')
    fetchKeys()
  }).catch(() => {})
}

function generateRandomKey() {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  return Array.from({ length: 32 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
}

function copyKey(value: string) {
  navigator.clipboard.writeText(value).then(() => ElMessage.success('已复制'))
}

onMounted(() => {
  fetchOrgs()
  fetchKeys()
})
</script>
