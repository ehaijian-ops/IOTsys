<template>
  <div class="page-container">
    <!-- 页头 -->
    <div class="page-header">
      <div class="header-left">
        <h2>站点管理</h2>
        <span class="header-count">共 {{ total }} 个站点</span>
      </div>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon> 新增站点
      </el-button>
    </div>

    <!-- 搜索筛选栏 -->
    <el-card class="search-card" shadow="never">
      <el-form :model="searchParams" inline :show-label="false" class="search-form">
        <el-form-item>
          <el-input
            v-model="searchParams.keyword"
            placeholder="站点名称/地址"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-select v-model="searchParams.status" placeholder="运营状态" clearable style="width: 130px">
            <el-option label="运营中" value="active" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input v-model="searchParams.province" placeholder="省份" clearable style="width: 120px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch"><el-icon><Search /></el-icon> 搜索</el-button>
        </el-form-item>
        <el-form-item>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 数据表格 -->
    <el-card shadow="never" class="table-card">
      <el-table
        ref="tableRef"
        :data="sites"
        v-loading="loading"
        stripe
        border
        highlight-current-row
        row-key="id"
        empty-text="暂无站点数据，点击上方按钮新增"
        style="width: 100%"
      >
        <el-table-column prop="name" label="站点名称" min-width="140" show-overflow-tooltip fixed />
        <el-table-column prop="address" label="详细地址" min-width="180" show-overflow-tooltip />
        <el-table-column label="省市" width="120">
          <template #default="{ row }">{{ row.province }}{{ row.city ? ` / ${row.city}` : '' }}</template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" min-width="100" show-overflow-tooltip />
        <el-table-column label="联系电话" width="130">
          <template #default="{ row }">{{ row.phone || '-' }}</template>
        </el-table-column>
        <el-table-column label="设备数" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.device_count > 0 ? 'primary' : 'info'" size="small" round>{{ row.device_count }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分佣比例" width="90" align="center">
          <template #default="{ row }">{{ row.commission_rate }}%</template>
        </el-table-column>
        <el-table-column label="收费方案" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ row.billing_scheme_name || '未设置' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'active' ? 'success' : 'info'"
              effect="dark"
              size="small"
              disable-transitions
            >
              {{ statusMap[row.status] || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-popconfirm
              title="确定删除该站点吗？此操作不可撤销。"
              :title="`确定删除「${row.name}」？${row.device_count > 0 ? '⚠️ 该站点下有设备' : ''}`"
              confirm-button-text="确认"
              cancel-button-text="取消"
              confirm-button-type="danger"
              icon-color="#F56C6C"
              :disabled="row.device_count > 0"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button size="small" link type="danger" :disabled="row.device_count > 0">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="fetchSites"
          @current-change="fetchSites"
        />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗（分区折叠表单） -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑站点' : '新增站点'"
      width="680px"
      :close-on-click-modal="false"
      destroy-on-close
      top="5vh"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="110px"
        label-position="right"
        class="site-form"
      >
        <el-collapse v-model="activeCollapse" accordion>
          <!-- 基本信息 -->
          <el-collapse-item name="basic" title="📋 基本信息">
            <el-row :gutter="20">
              <el-col :span="24">
                <el-form-item label="站点名称" prop="name">
                  <el-input v-model="form.name" placeholder="请输入站点名称" maxlength="100" show-word-limit clearable />
                </el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="详细地址" prop="address">
                  <el-input v-model="form.address" placeholder="请输入详细地址" maxlength="255" clearable />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="运营状态" prop="status">
                  <el-radio-group v-model="form.status">
                    <el-radio value="active">运营中</el-radio>
                    <el-radio value="inactive">停用</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="营业时间" prop="business_hours">
                  <el-time-picker
                    v-model="businessHoursRange"
                    is-range
                    range-separator="-"
                    start-placeholder="开始时间"
                    end-placeholder="结束时间"
                    format="HH:mm"
                    value-format="HH:mm"
                    style="width: 100%"
                    @change="onBusinessHoursChange"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="联系人" prop="contact">
                  <el-input v-model="form.contact" placeholder="联系人姓名" maxlength="50" clearable />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="联系电话" prop="phone">
                  <el-input v-model="form.phone" placeholder="手机或座机号" maxlength="20" clearable />
                </el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="站点图片" prop="image_url">
                  <div class="image-uploader-wrapper">
                    <el-input v-model="form.image_url" placeholder="请输入图片URL" clearable>
                      <template #prepend><el-icon><Picture /></el-icon></template>
                    </el-input>
                    <el-image
                      v-if="form.image_url"
                      :src="form.image_url"
                      fit="cover"
                      class="preview-img"
                      :preview-src-list="[form.image_url]"
                      preview-teleported
                    >
                      <template #error>
                        <div class="image-error"><el-icon><PictureFilled /></el-icon> 加载失败</div>
                      </template>
                    </el-image>
                  </div>
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 位置详情 -->
          <el-collapse-item name="location" title="📍 位置详情">
            <el-row :gutter="20">
              <el-col :xs="24" :sm="12">
                <el-form-item label="省份" prop="province">
                  <el-input v-model="form.province" placeholder="省份/直辖市/自治区" maxlength="30" clearable />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="城市" prop="city">
                  <el-input v-model="form.city" placeholder="城市/地级市" maxlength="30" clearable />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="经度" prop="longitude">
                  <el-input-number
                    v-model="form.longitude"
                    :precision="6"
                    :min="-180"
                    :max="180"
                    controls-position="right"
                    placeholder="经度 -180~180"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="纬度" prop="latitude">
                  <el-input-number
                    v-model="form.latitude"
                    :precision="6"
                    :min="-90"
                    :max="90"
                    controls-position="right"
                    placeholder="纬度 -90~90"
                    style="width: 100%"
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>

          <!-- 运营设置 -->
          <el-collapse-item name="operation" title="⚙️ 运营设置">
            <el-row :gutter="20">
              <el-col :xs="24" :sm="12">
                <el-form-item label="所属代理商" prop="agent_id">
                  <el-select v-model="form.agent_id" placeholder="选择代理商（可选）" clearable filterable allow-create style="width: 100%">
                    <el-option
                      v-for="a in agentList"
                      :key="a.value"
                      :label="a.label"
                      :value="a.value"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="所属运营商" prop="operator_id">
                  <el-select v-model="form.operator_id" placeholder="选择运营商（可选）" clearable filterable allow-create style="width: 100%">
                    <el-option
                      v-for="o in operatorList"
                      :key="o.value"
                      :label="o.label"
                      :value="o.value"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="收费方案" prop="billing_scheme_id">
                  <el-select v-model="form.billing_scheme_id" placeholder="选择收费方案（可选）" clearable filterable style="width: 100%">
                    <el-option
                      v-for="s in billingSchemeList"
                      :key="s.value"
                      :label="s.label"
                      :value="s.value"
                    />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="12">
                <el-form-item label="分佣比例(%)" prop="commission_rate">
                  <el-slider
                    v-model="form.commission_rate"
                    :min="0"
                    :max="100"
                    :step="0.1"
                    show-input
                    :show-input-controls="false"
                    input-size="small"
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>
        </el-collapse>
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
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import {
  getSitesManage,
  createSite,
  updateSite,
  deleteSite,
  type Site,
  type CreateSiteRequest,
  type UpdateSiteRequest,
  type SiteListParams,
} from '@/api/site'

// ─── 状态常量 ───
const statusMap: Record<string, string> = { active: '运营中', inactive: '停用' }

// ─── 列表状态 ───
const loading = ref(false)
const submitting = ref(false)
const sites = ref<Site[]>([])
const total = ref(0)
const tableRef = ref()
const pagination = reactive({ page: 1, page_size: 10 })
const searchParams = reactive<SiteListParams>({
  keyword: '',
  status: '',
  province: '',
})

// ─── 弹窗状态 ───
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref('')
const formRef = ref<FormInstance>()
const activeCollapse = ref('basic')
const businessHoursRange = ref<[string, string] | null>(null)

// 下拉选项（可从 API 获取，此处用静态示例）
interface SelectOption { label: string; value: string | number }
const agentList = ref<SelectOption[]>([])
const operatorList = ref<SelectOption[]>([])
const billingSchemeList = ref<SelectOption[]>([])

// ─── 表单默认值 & 模型 ───
const defaultForm = (): CreateSiteRequest => ({
  name: '',
  address: '',
  province: '',
  city: '',
  contact: '',
  phone: '',
  longitude: 116.397428,
  latitude: 39.90923,
  agent_id: undefined,
  operator_id: undefined,
  commission_rate: 10,
  billing_scheme_id: undefined,
  business_hours: '00:00-23:59',
  image_url: '',
  status: 'active',
})
const form = reactive(defaultForm())

// ─── 表单验证规则 ───
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入站点名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 ~ 100 字符', trigger: 'blur' },
  ],
  address: [
    { required: true, message: '请输入站点地址', trigger: 'blur' },
    { max: 255, message: '最长 255 字符', trigger: 'blur' },
  ],
  phone: [
    { pattern: /^[\d\-+\s()]{6,20}$/, message: '请输入有效的联系电话', trigger: 'blur' },
  ],
  longitude: [{ type: 'number', min: -180, max: 180, message: '经度范围 -180~180', trigger: 'blur' }],
  latitude: [{ type: 'number', min: -90, max: 90, message: '纬度范围 -90~90', trigger: 'blur' }],
  commission_rate: [{ type: 'number', min: 0, max: 100, message: '分佣比例 0~100%', trigger: 'blur' }],
}

// ─── 营业时间转换 ───
function parseBusinessHours(val?: string): [string, string] | null {
  if (!val) return null
  const parts = val.split('-')
  if (parts.length === 2) return parts as [string, string]
  return null
}

function onBusinessHoursChange(val: [string, string] | null) {
  form.business_hours = val ? `${val[0]}-${val[1]}` : ''
}

// ─── 列表 CRUD ───

async function fetchSites() {
  loading.value = true
  try {
    const res = await getSitesManage({
      ...searchParams,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    const data = res.data as any
    sites.value = data?.items ?? (Array.isArray(data) ? data : [])
    total.value = data?.total ?? (Array.isArray(data) ? data.length : 0)
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  fetchSites()
}

function handleReset() {
  searchParams.keyword = ''
  searchParams.status = ''
  searchParams.province = ''
  pagination.page = 1
  fetchSites()
}

// ─── 弹窗逻辑 ───

function openCreateDialog() {
  isEdit.value = false
  editingId.value = ''
  Object.assign(form, defaultForm())
  businessHoursRange.value = null
  activeCollapse.value = 'basic'
  dialogVisible.value = true
}

function openEditDialog(site: Site) {
  isEdit.value = true
  editingId.value = String(site.id)

  // 填充全部字段
  Object.assign(form, {
    name: site.name,
    address: site.address || '',
    province: site.province || '',
    city: site.city || '',
    contact: site.contact || '',
    phone: site.phone || '',
    longitude: site.longitude ?? 0,
    latitude: site.latitude ?? 0,
    agent_id: site.agent_id,
    operator_id: site.operator_id,
    commission_rate: site.commission_rate ?? 10,
    billing_scheme_id: site.billing_scheme_id,
    business_hours: site.business_hours || '00:00-23:59',
    image_url: site.image_url || '',
    status: site.status || 'active',
  })

  businessHoursRange.value = parseBusinessHours(form.business_hours)
  activeCollapse.value = 'basic'
  dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = { ...form }
    if (!payload.agent_id) delete payload.agent_id
    if (!payload.operator_id) delete payload.operator_id
    if (!payload.billing_scheme_id) delete payload.billing_scheme_id

    if (isEdit.value) {
      await updateSite(editingId.value, payload as UpdateSiteRequest)
      ElMessage.success('站点信息已更新')
    } else {
      await createSite(payload)
      ElMessage.success('站点创建成功')
    }

    dialogVisible.value = false
    fetchSites()
  } catch {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function handleDelete(site: Site) {
  try {
    await deleteSite(String(site.id))
    ElMessage.success(`站点「${site.name}」已删除`)
    // 如果当前页删完了且不是第一页，回到上一页
    if (sites.value.length <= 1 && pagination.page > 1) {
      pagination.page--
    }
    fetchSites()
  } catch {
    // error handled by interceptor
  }
}

// ─── 生命周期 ───
onMounted(() => {
  fetchSites()
})
</script>

<style scoped>
.page-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.header-count {
  color: #909399;
  font-size: 13px;
}

/* 搜索栏 */
.search-card :deep(.el-card__body) {
  padding-bottom: 2px;
}
.search-form {
  display: flex;
  flex-wrap: wrap;
}
.search-form .el-form-item {
  margin-bottom: 14px;
  margin-right: 8px;
}

/* 表格 */
.table-card {
  flex: 1;
}
.table-card :deep(.el-card__body) {
  padding: 16px;
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

/* 弹窗表单 */
.site-form .el-collapse {
  border: none;
}
.site-form :deep(.el-collapse-item__header) {
  font-size: 14px;
  font-weight: 600;
  padding-left: 4px;
  background: #fafbfc;
  border-radius: 4px;
  margin-bottom: 8px;
}
.site-form :deep(.el-collapse-item__content) {
  padding: 4px 16px 16px;
}
.site-form .el-form-item {
  margin-bottom: 18px;
}

/* 图片上传区域 */
.image-uploader-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.preview-img {
  width: 60px;
  height: 60px;
  border-radius: 6px;
  border: 1px solid #dcdfe6;
  cursor: pointer;
}
.image-error {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #c0c4cc;
  font-size: 12px;
  flex-direction: column;
  gap: 4px;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .pagination-wrapper {
    justify-content: center;
  }
  .site-form :deep(.el-form-item__label) {
    text-align: left !important;
    float: none;
    display: block;
  }
}
</style>
