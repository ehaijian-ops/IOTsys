<template>
  <div class="page-container">
    <div class="page-header">
      <h2>站点管理</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 新增站点
      </el-button>
    </div>

    <!-- 站点卡片列表 -->
    <el-row :gutter="20" v-loading="loading">
      <el-col :span="8" v-for="site in sites" :key="site.id" style="margin-bottom: 20px">
        <el-card class="dashboard-card">
          <template #header>
            <div class="site-header">
              <span>{{ site.name }}</span>
              <el-tag :type="site.status === 'active' ? 'success' : 'info'" size="small">
                {{ site.status === 'active' ? '运营中' : '停用' }}
              </el-tag>
            </div>
          </template>
          <div class="site-info">
            <p v-if="site.address"><el-icon><Location /></el-icon> {{ site.address }}</p>
            <p><el-icon><Monitor /></el-icon> 设备: {{ site.device_count }} 台</p>
            <p v-if="site.contact"><el-icon><User /></el-icon> 联系人: {{ site.contact }}{{ site.phone ? ' · ' + site.phone : '' }}</p>
          </div>
          <div class="site-actions">
            <el-button size="small" @click="openEditDialog(site)" type="primary">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(site)">删除</el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 空状态 -->
      <el-col :span="24" v-if="!loading && sites.length === 0">
        <el-empty description="暂无站点，点击上方按钮新增" />
      </el-col>
    </el-row>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑站点' : '新增站点'"
      width="560px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="90px"
        label-position="right"
      >
        <el-form-item label="站点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入站点名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="站点地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入站点地址" maxlength="255" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact">
          <el-input v-model="form.contact" placeholder="请输入联系人" maxlength="50" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入联系电话" maxlength="20" />
        </el-form-item>
        <el-form-item label="经度" prop="longitude">
          <el-input-number v-model="form.longitude" :precision="6" :min="-180" :max="180" style="width: 100%" />
        </el-form-item>
        <el-form-item label="纬度" prop="latitude">
          <el-input-number v-model="form.latitude" :precision="6" :min="-90" :max="90" style="width: 100%" />
        </el-form-item>
        <el-form-item label="运营状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio value="active">运营中</el-radio>
            <el-radio value="inactive">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false" :disabled="submitting">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存修改' : '确认创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  getSitesManage,
  createSite,
  updateSite,
  deleteSite,
  type Site,
  type CreateSiteRequest,
  type UpdateSiteRequest,
} from '@/api/site'

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref('')
const sites = ref<Site[]>([])
const formRef = ref<FormInstance>()

const defaultForm = () => ({
  name: '',
  address: '',
  contact: '',
  phone: '',
  longitude: 0,
  latitude: 0,
  status: 'active',
})

const form = reactive(defaultForm())

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入站点名称', trigger: 'blur' },
    { min: 2, max: 100, message: '站点名称长度在 2 到 100 个字符', trigger: 'blur' },
  ],
}

onMounted(() => {
  fetchSites()
})

// 获取站点列表
async function fetchSites() {
  loading.value = true
  try {
    const res = await getSitesManage()
    sites.value = res.data || []
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

// 打开新增弹窗
function openCreateDialog() {
  isEdit.value = false
  editingId.value = ''
  Object.assign(form, defaultForm())
  dialogVisible.value = true
}

// 打开编辑弹窗
function openEditDialog(site: Site) {
  isEdit.value = true
  editingId.value = site.id
  form.name = site.name
  form.address = site.address || ''
  form.contact = site.contact || ''
  form.phone = site.phone || ''
  form.longitude = site.longitude || 0
  form.latitude = site.latitude || 0
  form.status = site.status || 'active'
  dialogVisible.value = true
}

// 提交表单
async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit.value) {
      const payload: UpdateSiteRequest = {
        name: form.name,
        address: form.address,
        contact: form.contact,
        phone: form.phone,
        longitude: form.longitude,
        latitude: form.latitude,
        status: form.status,
      }
      await updateSite(editingId.value, payload)
      ElMessage.success('站点信息已更新')
    } else {
      const payload: CreateSiteRequest = {
        name: form.name,
        address: form.address,
        contact: form.contact,
        phone: form.phone,
        longitude: form.longitude,
        latitude: form.latitude,
        status: form.status,
      }
      await createSite(payload)
      ElMessage.success('站点创建成功')
    }

    dialogVisible.value = false
    await fetchSites()
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

// 删除站点
async function handleDelete(site: Site) {
  if (site.device_count > 0) {
    ElMessage.warning('该站点下还有设备，请先将设备迁移后再删除')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除站点「${site.name}」吗？此操作不可撤销。`,
      '删除确认',
      {
        confirmButtonText: '确认删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
  } catch {
    return // 用户取消
  }

  try {
    await deleteSite(site.id)
    ElMessage.success('站点已删除')
    await fetchSites()
  } catch {
    // error handled by interceptor
  }
}
</script>

<style scoped>
.site-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.site-info p {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.site-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}
</style>
