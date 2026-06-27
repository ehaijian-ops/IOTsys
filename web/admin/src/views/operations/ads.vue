<template>
  <div class="page-container">
    <div class="page-header">
      <h2>广告与运营</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 广告列表 -->
      <el-tab-pane label="广告列表" name="ads">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openAdDialog()">
            <el-icon><Plus /></el-icon> 新增广告
          </el-button>
        </div>
        <el-card>
          <el-table :data="ads" stripe>
            <el-table-column prop="title" label="标题" width="200" />
            <el-table-column prop="image_url" label="图片" width="150">
              <template #default="{ row }">
                <el-image :src="row.image_url" fit="cover" style="width: 80px; height: 50px" preview-teleported />
              </template>
            </el-table-column>
            <el-table-column prop="link_url" label="跳转链接" min-width="200" show-overflow-tooltip />
            <el-table-column prop="position" label="位置" width="100">
              <template #default="{ row }">{{ { home_top: '首页顶部', home_middle: '首页中部', charge_page: '充电页' }[row.position] || row.position }}</template>
            </el-table-column>
            <el-table-column prop="sort_order" label="排序" width="80" align="center" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-switch :model-value="row.status === 'active'" @change="() => toggleAd(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openAdDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteAd(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 加盟合作 -->
      <el-tab-pane label="加盟合作" name="franchise">
        <el-card>
          <el-table :data="franchises" stripe>
            <el-table-column prop="company_name" label="公司名称" width="200" />
            <el-table-column prop="contact_name" label="联系人" width="120" />
            <el-table-column prop="contact_phone" label="联系电话" width="140" />
            <el-table-column prop="city" label="城市" width="120" />
            <el-table-column prop="remark" label="备注" min-width="200" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'" size="small">
                  {{ { pending: '待审核', approved: '已通过', rejected: '已驳回' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button v-if="row.status === 'pending'" type="success" link size="small" @click="processFq(row, 'approved')">通过</el-button>
                <el-button v-if="row.status === 'pending'" type="danger" link size="small" @click="processFq(row, 'rejected')">驳回</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 月卡套餐 -->
      <el-tab-pane label="月卡方案" name="monthly">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openMonthlyDialog()">
            <el-icon><Plus /></el-icon> 新增月卡方案
          </el-button>
        </div>
        <el-card>
          <el-table :data="monthlySchemes" stripe>
            <el-table-column prop="name" label="方案名称" width="200" />
            <el-table-column prop="price" label="价格(元)" width="120" align="right">
              <template #default="{ row }">¥{{ (row.price || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="duration_days" label="有效天数" width="100" align="center" />
            <el-table-column prop="max_energy_kwh" label="最大电量(kWh)" width="130" align="right" />
            <el-table-column prop="max_charges" label="最大充电次数" width="120" align="center" />
            <el-table-column prop="device_type" label="适用设备" width="130" />
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openMonthlyDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteMonthly(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 微信用户 -->
      <el-tab-pane label="微信用户" name="wechat">
        <el-card>
          <el-table :data="wechatUsers" stripe>
            <el-table-column prop="nickname" label="昵称" width="160" />
            <el-table-column prop="phone" label="手机号" width="140" />
            <el-table-column prop="balance" label="余额(元)" width="120" align="right">
              <template #default="{ row }">¥{{ (row.balance || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="total_recharge" label="累计充值" width="120" align="right">
              <template #default="{ row }">¥{{ (row.total_recharge || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column prop="total_charge" label="累计充电次数" width="120" align="center" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'frozen' ? 'danger' : 'success'" size="small">
                  {{ row.status === 'frozen' ? '已冻结' : '正常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button v-if="row.status !== 'frozen'" type="danger" link size="small" @click="freezeUser(row)">冻结</el-button>
                <el-button v-else type="success" link size="small" @click="freezeUser(row)">解冻</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 广告弹窗 -->
    <el-dialog v-model="adDialog.visible" :title="adDialog.isEdit ? '编辑广告' : '新增广告'" width="520px">
      <el-form :model="adDialog.form" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="adDialog.form.title" placeholder="广告标题" />
        </el-form-item>
        <el-form-item label="图片地址">
          <el-input v-model="adDialog.form.image_url" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="adDialog.form.link_url" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="位置">
          <el-select v-model="adDialog.form.position" style="width: 100%">
            <el-option label="首页顶部" value="home_top" />
            <el-option label="首页中部" value="home_middle" />
            <el-option label="充电页" value="charge_page" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="adDialog.form.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="adDialog.form.status" active-value="active" inactive-value="inactive" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveAd" :loading="adSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 月卡方案弹窗 -->
    <el-dialog v-model="monthlyDialog.visible" :title="monthlyDialog.isEdit ? '编辑月卡方案' : '新增月卡方案'" width="480px">
      <el-form :model="monthlyDialog.form" label-width="120px">
        <el-form-item label="方案名称">
          <el-input v-model="monthlyDialog.form.name" />
        </el-form-item>
        <el-form-item label="价格(元)">
          <el-input-number v-model="monthlyDialog.form.price" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="有效天数">
          <el-input-number v-model="monthlyDialog.form.duration_days" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大电量(kWh)">
          <el-input-number v-model="monthlyDialog.form.max_energy_kwh" :min="0" :precision="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="最大充电次数">
          <el-input-number v-model="monthlyDialog.form.max_charges" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="适用设备类型">
          <el-input v-model="monthlyDialog.form.device_type" placeholder="如: ebike_charger" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="monthlyDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveMonthly" :loading="monthlySaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAds, createAd, updateAd, deleteAd } from '@/api/advertisement'
import { getFranchises, processFranchise } from '@/api/advertisement'
import { getWechatUsers, freezeWechatUser } from '@/api/advertisement'
import { getMonthlyCardSchemes, createMonthlyCardScheme, updateMonthlyCardScheme, deleteMonthlyCardScheme } from '@/api/billing'

const activeTab = ref('ads')

// 广告
const ads = ref<any[]>([])
const adDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const adSaving = ref(false)

// 加盟
const franchises = ref<any[]>([])

// 月卡方案
const monthlySchemes = ref<any[]>([])
const monthlyDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const monthlySaving = ref(false)

// 微信用户
const wechatUsers = ref<any[]>([])

async function fetchAds() {
  try {
    const res: any = await getAds()
    ads.value = res?.data || res?.list || []
  } catch {}
}

function openAdDialog(row?: any) {
  adDialog.isEdit = !!row
  adDialog.form = row ? { ...row } : { title: '', image_url: '', link_url: '', position: 'home_top', sort_order: 0, status: 'active' }
  adDialog.visible = true
}

async function doSaveAd() {
  adSaving.value = true
  try {
    if (adDialog.isEdit) {
      await updateAd(adDialog.form.id, adDialog.form)
      ElMessage.success('广告已更新')
    } else {
      await createAd(adDialog.form)
      ElMessage.success('广告已创建')
    }
    adDialog.visible = false
    fetchAds()
  } catch {} finally {
    adSaving.value = false
  }
}

async function toggleAd(row: any) {
  const newStatus = row.status === 'active' ? 'inactive' : 'active'
  try {
    await updateAd(row.id, { ...row, status: newStatus })
    row.status = newStatus
  } catch {}
}

function handleDeleteAd(row: any) {
  ElMessageBox.confirm('确认删除该广告？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteAd(row.id)
    ElMessage.success('已删除')
    fetchAds()
  }).catch(() => {})
}

async function fetchFranchises() {
  try {
    const res: any = await getFranchises()
    franchises.value = res?.data || res?.list || []
  } catch {}
}

async function processFq(row: any, status: string) {
  try {
    await processFranchise(row.id, { status })
    ElMessage.success(status === 'approved' ? '已通过' : '已驳回')
    fetchFranchises()
  } catch {}
}

// 月卡
async function fetchMonthly() {
  try {
    const res: any = await getMonthlyCardSchemes()
    monthlySchemes.value = res?.data || res?.list || []
  } catch {}
}

function openMonthlyDialog(row?: any) {
  monthlyDialog.isEdit = !!row
  monthlyDialog.form = row ? { ...row } : { name: '', price: 0, duration_days: 30, max_energy_kwh: 0, max_charges: 0, device_type: '' }
  monthlyDialog.visible = true
}

async function doSaveMonthly() {
  monthlySaving.value = true
  try {
    if (monthlyDialog.isEdit) {
      await updateMonthlyCardScheme(monthlyDialog.form.id, monthlyDialog.form)
      ElMessage.success('方案已更新')
    } else {
      await createMonthlyCardScheme(monthlyDialog.form)
      ElMessage.success('方案已创建')
    }
    monthlyDialog.visible = false
    fetchMonthly()
  } catch {} finally {
    monthlySaving.value = false
  }
}

function handleDeleteMonthly(row: any) {
  ElMessageBox.confirm('确认删除该月卡方案？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteMonthlyCardScheme(row.id)
    ElMessage.success('已删除')
    fetchMonthly()
  }).catch(() => {})
}

// 微信用户
async function fetchWechat() {
  try {
    const res: any = await getWechatUsers({ page: 1, page_size: 1000 })
    wechatUsers.value = res?.data || res?.list || []
  } catch {}
}

async function freezeUser(row: any) {
  try {
    await freezeWechatUser(row.id)
    ElMessage.success(row.status === 'frozen' ? '用户已解冻' : '用户已冻结')
    fetchWechat()
  } catch {}
}

onMounted(() => {
  fetchAds()
  fetchFranchises()
  fetchMonthly()
  fetchWechat()
})
</script>
