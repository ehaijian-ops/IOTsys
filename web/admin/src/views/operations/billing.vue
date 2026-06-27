<template>
  <div class="page-container">
    <div class="page-header">
      <h2>收费方案管理</h2>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon> 新增方案
      </el-button>
    </div>

    <!-- 收费方案列表 -->
    <el-card>
      <el-tabs v-model="schemeTab" @tab-change="onSchemeTabChange">
        <el-tab-pane label="汽车收费方案" name="car">
          <el-table :data="schemes.filter((s: any) => s.scheme_type === 'car')" stripe>
            <el-table-column prop="name" label="方案名称" width="200" />
            <el-table-column prop="service_fee" label="服务费(元)" width="100" align="right" />
            <el-table-column prop="base_price" label="基础电价(元/kWh)" width="130" align="right" />
            <el-table-column label="尖峰平谷" min-width="300">
              <template #default="{ row }">
                <el-tag v-for="p in row.periods" :key="p.id" size="small" style="margin: 2px">
                  {{ p.period }} ¥{{ p.price }}/kWh
                </el-tag>
                <span v-if="!row.periods?.length" style="color: #909399">未配置</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
                <el-button type="success" link size="small" @click="openPeriod(row)">时段</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="时间收费方案" name="time">
          <el-table :data="schemes.filter((s: any) => s.scheme_type === 'time')" stripe>
            <el-table-column prop="name" label="方案名称" width="200" />
            <el-table-column prop="price_per_minute" label="每分钟价格(元)" width="140" align="right" />
            <el-table-column prop="max_duration_min" label="最大时长(分钟)" width="140" align="right" />
            <el-table-column prop="auto_stop_full" label="充满自停" width="100" align="center">
              <template #default="{ row }"><el-switch :model-value="row.auto_stop_full" disabled /></template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="电量收费方案" name="energy">
          <el-table :data="schemes.filter((s: any) => s.scheme_type === 'energy')" stripe>
            <el-table-column prop="name" label="方案名称" width="200" />
            <el-table-column prop="price_per_kwh" label="每度价格(元)" width="140" align="right" />
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="功率收费方案" name="power">
          <el-table :data="schemes.filter((s: any) => s.scheme_type === 'power')" stripe>
            <el-table-column prop="name" label="方案名称" width="200" />
            <el-table-column prop="price_per_kw" label="每千瓦价格(元)" width="150" align="right" />
            <el-table-column prop="prepaid" label="预收费" width="100" align="center">
              <template #default="{ row }"><el-switch :model-value="row.prepaid" disabled /></template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 收费方案弹窗 -->
    <el-dialog v-model="dialog.visible" :title="dialog.isEdit ? '编辑方案' : '新增方案'" width="520px">
      <el-form :model="dialog.form" label-width="110px">
        <el-form-item label="方案名称">
          <el-input v-model="dialog.form.name" placeholder="如：标准充电方案" />
        </el-form-item>
        <el-form-item label="方案类型">
          <el-select v-model="dialog.form.scheme_type" :disabled="dialog.isEdit" style="width: 100%">
            <el-option label="汽车收费方案" value="car" />
            <el-option label="时间收费方案" value="time" />
            <el-option label="电量收费方案" value="energy" />
            <el-option label="功率收费方案" value="power" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'car'" label="服务费(元)">
          <el-input-number v-model="dialog.form.service_fee" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'car'" label="基础电价(元/kWh)">
          <el-input-number v-model="dialog.form.base_price" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'time'" label="每分钟价格(元)">
          <el-input-number v-model="dialog.form.price_per_minute" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'time'" label="最大时长(分钟)">
          <el-input-number v-model="dialog.form.max_duration_min" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'time'" label="充满自停">
          <el-switch v-model="dialog.form.auto_stop_full" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'energy'" label="每度价格(元)">
          <el-input-number v-model="dialog.form.price_per_kwh" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'power'" label="每千瓦价格(元)">
          <el-input-number v-model="dialog.form.price_per_kw" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="dialog.form.scheme_type === 'power'" label="预收费">
          <el-switch v-model="dialog.form.prepaid" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 时段配置弹窗 -->
    <el-dialog v-model="periodDialog.visible" title="尖峰平谷时段配置" width="600px">
      <el-table :data="periodDialog.periods" stripe>
        <el-table-column prop="period" label="时段类型" width="120">
          <template #default="{ row }">
            <el-select v-model="row.period" size="small" style="width: 100px">
              <el-option label="尖峰" value="peak" />
              <el-option label="高峰" value="high" />
              <el-option label="平段" value="flat" />
              <el-option label="低谷" value="valley" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="130">
          <template #default="{ row }"><el-time-picker v-model="row.start_time" size="small" format="HH:mm" value-format="HH:mm" /></template>
        </el-table-column>
        <el-table-column prop="end_time" label="结束时间" width="130">
          <template #default="{ row }"><el-time-picker v-model="row.end_time" size="small" format="HH:mm" value-format="HH:mm" /></template>
        </el-table-column>
        <el-table-column prop="price" label="电价(元/kWh)" width="130">
          <template #default="{ row }"><el-input-number v-model="row.price" size="small" :min="0" :precision="2" /></template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ $index }">
            <el-button type="danger" link size="small" @click="periodDialog.periods.splice($index, 1)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="dashed" style="margin-top: 12px; width: 100%" @click="periodDialog.periods.push({ period: 'flat', start_time: '', end_time: '', price: 0 })">
        + 添加时段
      </el-button>
      <template #footer>
        <el-button @click="periodDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSavePeriods" :loading="savingPeriods">保存时段</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getBillingSchemes, createBillingScheme, updateBillingScheme, deleteBillingScheme,
  getBillingPeriods, setBillingPeriods,
} from '@/api/billing'

const schemeTab = ref('car')
const schemes = ref<any[]>([])
const saving = ref(false)
const savingPeriods = ref(false)

const dialog = reactive({
  visible: false, isEdit: false,
  form: {} as any,
})

const periodDialog = reactive({
  visible: false, schemeId: '',
  periods: [] as any[],
})

async function fetchSchemes() {
  try {
    const res: any = await getBillingSchemes()
    schemes.value = res?.data || res?.list || []
  } catch {}
}

function openCreate() {
  dialog.isEdit = false
  dialog.form = { scheme_type: schemeTab.value }
  dialog.visible = true
}

function openEdit(row: any) {
  dialog.isEdit = true
  dialog.form = { ...row }
  dialog.visible = true
}

async function doSave() {
  saving.value = true
  try {
    if (dialog.isEdit) {
      await updateBillingScheme(dialog.form.id, dialog.form)
      ElMessage.success('方案已更新')
    } else {
      await createBillingScheme(dialog.form)
      ElMessage.success('方案已创建')
    }
    dialog.visible = false
    fetchSchemes()
  } catch {} finally {
    saving.value = false
  }
}

function handleDelete(row: any) {
  ElMessageBox.confirm('确认删除该收费方案？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteBillingScheme(row.id)
    ElMessage.success('已删除')
    fetchSchemes()
  }).catch(() => {})
}

async function openPeriod(row: any) {
  periodDialog.schemeId = row.id
  periodDialog.periods = []
  try {
    const res: any = await getBillingPeriods(row.id)
    const periods = res?.data || res?.list || []
    periodDialog.periods = periods.length > 0 ? periods.map((p: any) => ({ ...p })) : [{ period: 'flat', start_time: '', end_time: '', price: 0 }]
  } catch {
    periodDialog.periods = [{ period: 'flat', start_time: '', end_time: '', price: 0 }]
  }
  periodDialog.visible = true
}

async function doSavePeriods() {
  savingPeriods.value = true
  try {
    await setBillingPeriods(periodDialog.schemeId, periodDialog.periods)
    ElMessage.success('时段已保存')
    periodDialog.visible = false
    fetchSchemes()
  } catch {} finally {
    savingPeriods.value = false
  }
}

function onSchemeTabChange() {
  if (!dialog.visible) dialog.form.scheme_type = schemeTab.value
}

onMounted(fetchSchemes)
</script>
