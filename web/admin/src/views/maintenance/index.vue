<template>
  <div class="page-container">
    <div class="page-header">
      <h2>运维管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 故障反馈 -->
      <el-tab-pane label="故障反馈" name="faults">
        <el-card>
          <el-form :inline="true" :model="faultQuery">
            <el-form-item label="状态">
              <el-select v-model="faultQuery.status" placeholder="全部" clearable style="width: 140px">
                <el-option label="待处理" value="pending" />
                <el-option label="处理中" value="processing" />
                <el-option label="已处理" value="resolved" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchFaults">查询</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="faults" stripe v-loading="faultLoading" @row-click="expandFault">
            <el-table-column type="index" width="50" />
            <el-table-column prop="device_sn" label="设备SN" width="180" />
            <el-table-column prop="device_type" label="设备类型" width="120" />
            <el-table-column prop="description" label="故障描述" min-width="260" show-overflow-tooltip />
            <el-table-column prop="reporter_name" label="上报人" width="120" />
            <el-table-column prop="reporter_phone" label="联系电话" width="140" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'pending' ? 'danger' : row.status === 'processing' ? 'warning' : 'success'" size="small">
                  {{ { pending: '待处理', processing: '处理中', resolved: '已处理' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="上报时间" width="170" />
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button v-if="row.status === 'pending'" type="primary" link size="small" @click.stop="handleFaultAction(row, 'processing')">受理</el-button>
                <el-button v-if="row.status === 'processing'" type="success" link size="small" @click.stop="openFaultResult(row)">处理</el-button>
              </template>
            </el-table-column>
            <template #expanded="{ row }">
              <div class="fault-detail">
                <p><strong>故障详情：</strong>{{ row.description }}</p>
                <p v-if="row.result"><strong>处理结果：</strong>{{ row.result }}</p>
                <p v-if="row.image_urls">
                  <strong>故障图片：</strong>
                  <span v-for="(url, i) in (row.image_urls || '').split(',')" :key="i">
                    <el-image :src="url" style="width: 80px; height: 60px; margin: 4px" fit="cover" preview-teleported />
                  </span>
                </p>
              </div>
            </template>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 定时任务 -->
      <el-tab-pane label="定时任务" name="tasks">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openTaskDialog()">
            <el-icon><Plus /></el-icon> 新增任务
          </el-button>
        </div>
        <el-card>
          <el-table :data="tasks" stripe>
            <el-table-column prop="name" label="任务名称" width="200" />
            <el-table-column prop="cron_expr" label="Cron表达式" width="160" />
            <el-table-column prop="handler" label="处理器" width="200" />
            <el-table-column prop="last_run_at" label="上次执行" width="170" />
            <el-table-column prop="next_run_at" label="下次执行" width="170" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-switch :model-value="row.status === 'enabled'" @change="() => toggleTask(row)" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="260">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="runTask(row)">立即执行</el-button>
                <el-button type="success" link size="small" @click="viewTaskLogs(row)">日志</el-button>
                <el-button type="warning" link size="small" @click="openTaskDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteTask(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 下载中心 -->
      <el-tab-pane label="下载中心" name="downloads">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openDownloadDialog()">
            <el-icon><Plus /></el-icon> 创建下载任务
          </el-button>
        </div>
        <el-card>
          <el-table :data="downloads" stripe>
            <el-table-column prop="task_name" label="任务名称" min-width="200" />
            <el-table-column prop="task_type" label="类型" width="140">
              <template #default="{ row }">{{ { order_export: '订单导出', device_export: '设备导出', report: '报表下载' }[row.task_type] || row.task_type }}</template>
            </el-table-column>
            <el-table-column prop="file_size" label="文件大小" width="100" align="right">
              <template #default="{ row }">{{ row.file_size ? (row.file_size / 1024).toFixed(1) + ' KB' : '-' }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'completed' ? 'success' : row.status === 'processing' ? 'warning' : 'info'" size="small">
                  {{ { pending: '等待中', processing: '处理中', completed: '已完成', failed: '失败' }[row.status] || row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="170" />
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button v-if="row.status === 'completed'" type="primary" link size="small" @click="downloadFile(row)">下载</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteDownload(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 故障处理弹窗 -->
    <el-dialog v-model="faultResultDialog.visible" title="故障处理" width="480px">
      <el-form label-width="80px">
        <el-form-item label="故障描述">
          <span>{{ faultResultDialog.row.description }}</span>
        </el-form-item>
        <el-form-item label="处理结果">
          <el-input v-model="faultResultDialog.result" type="textarea" :rows="4" placeholder="请填写处理结果" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="faultResultDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doResolveFault">提交处理</el-button>
      </template>
    </el-dialog>

    <!-- 任务弹窗 -->
    <el-dialog v-model="taskDialog.visible" :title="taskDialog.isEdit ? '编辑任务' : '新增任务'" width="540px">
      <el-form :model="taskDialog.form" label-width="100px">
        <el-form-item label="任务名称">
          <el-input v-model="taskDialog.form.name" placeholder="如：每日站点统计" />
        </el-form-item>
        <el-form-item label="Cron表达式">
          <el-input v-model="taskDialog.form.cron_expr" placeholder="如：0 2 * * *" />
        </el-form-item>
        <el-form-item label="处理器">
          <el-input v-model="taskDialog.form.handler" placeholder="如：site_daily_stats" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="taskDialog.form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="taskDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveTask">保存</el-button>
      </template>
    </el-dialog>

    <!-- 下载任务弹窗 -->
    <el-dialog v-model="downloadDialog.visible" title="创建下载任务" width="520px">
      <el-form :model="downloadDialog.form" label-width="80px">
        <el-form-item label="任务名称">
          <el-input v-model="downloadDialog.form.task_name" placeholder="如：2024年6月订单报表" />
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select v-model="downloadDialog.form.task_type" style="width: 100%">
            <el-option label="订单导出" value="order_export" />
            <el-option label="设备导出" value="device_export" />
            <el-option label="报表下载" value="report" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="downloadDialog.form.task_type === 'order_export'" label="时间范围">
          <el-date-picker v-model="downloadDialog.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="downloadDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doCreateDownload">创建</el-button>
      </template>
    </el-dialog>

    <!-- 任务日志弹窗 -->
    <el-dialog v-model="taskLogsDialog.visible" title="任务执行日志" width="700px">
      <el-table :data="taskLogsDialog.logs" stripe size="small">
        <el-table-column type="index" width="50" />
        <el-table-column prop="start_time" label="开始时间" width="170" />
        <el-table-column prop="end_time" label="结束时间" width="170" />
        <el-table-column prop="status" label="结果" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">{{ row.status === 'success' ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="输出" min-width="200" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getFaults, handleFault } from '@/api/maintenance'
import { getTasks, createTask, updateTask, deleteTask, getTaskLogs } from '@/api/maintenance'
import { getDownloads, createDownload } from '@/api/maintenance'

const activeTab = ref('faults')

// 故障
const faults = ref<any[]>([])
const faultLoading = ref(false)
const faultQuery = reactive({ status: '' })
const faultResultDialog = reactive({ visible: false, row: {} as any, result: '' })

async function fetchFaults() {
  faultLoading.value = true
  try {
    const params: any = {}
    if (faultQuery.status) params.status = faultQuery.status
    const res: any = await getFaults(params)
    faults.value = res?.data || res?.list || []
  } catch {} finally {
    faultLoading.value = false
  }
}

function expandFault(row: any) {
  row._expanded = !row._expanded
}

async function handleFaultAction(row: any, status: string) {
  try {
    await handleFault(row.id, { result: status === 'processing' ? '故障已受理，正在处理中' : '' })
    ElMessage.success(status === 'processing' ? '已受理' : '已处理')
    fetchFaults()
  } catch {}
}

function openFaultResult(row: any) {
  faultResultDialog.row = row
  faultResultDialog.result = ''
  faultResultDialog.visible = true
}

async function doResolveFault() {
  try {
    await handleFault(faultResultDialog.row.id, { result: faultResultDialog.result })
    ElMessage.success('故障已处理')
    faultResultDialog.visible = false
    fetchFaults()
  } catch {}
}

// 定时任务
const tasks = ref<any[]>([])
const taskDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const taskLogsDialog = reactive({ visible: false, logs: [] as any[] })

async function fetchTasks() {
  try {
    const res: any = await getTasks()
    tasks.value = res?.data || res?.list || []
  } catch {}
}

function openTaskDialog(row?: any) {
  taskDialog.isEdit = !!row
  taskDialog.form = row ? { ...row } : { name: '', cron_expr: '', handler: '', description: '' }
  taskDialog.visible = true
}

async function doSaveTask() {
  try {
    if (taskDialog.isEdit) {
      await updateTask(taskDialog.form.id, taskDialog.form)
      ElMessage.success('任务已更新')
    } else {
      await createTask(taskDialog.form)
      ElMessage.success('任务已创建')
    }
    taskDialog.visible = false
    fetchTasks()
  } catch {}
}

function toggleTask(row: any) {
  const newStatus = row.status === 'enabled' ? 'disabled' : 'enabled'
  updateTask(row.id, { ...row, status: newStatus }).then(() => {
    row.status = newStatus
    ElMessage.success(newStatus === 'enabled' ? '任务已启用' : '任务已禁用')
  })
}

function runTask(row: any) {
  ElMessage.info(`正在执行任务: ${row.name}`)
}

function handleDeleteTask(row: any) {
  ElMessageBox.confirm('确认删除该任务？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteTask(row.id)
    ElMessage.success('已删除')
    fetchTasks()
  }).catch(() => {})
}

async function viewTaskLogs(row: any) {
  try {
    const res: any = await getTaskLogs(row.id)
    taskLogsDialog.logs = res?.data || res?.list || []
  } catch {
    taskLogsDialog.logs = []
  }
  taskLogsDialog.visible = true
}

// 下载中心
const downloads = ref<any[]>([])
const downloadDialog = reactive({ visible: false, form: {} as any, dateRange: null as [string, string] | null })

async function fetchDownloads() {
  try {
    const res: any = await getDownloads()
    downloads.value = res?.data || res?.list || []
  } catch {}
}

function openDownloadDialog() {
  downloadDialog.form = { task_name: '', task_type: 'order_export' }
  downloadDialog.dateRange = null
  downloadDialog.visible = true
}

async function doCreateDownload() {
  try {
    await createDownload(downloadDialog.form)
    ElMessage.success('下载任务已创建')
    downloadDialog.visible = false
    fetchDownloads()
  } catch {}
}

function downloadFile(row: any) {
  ElMessage.info(`开始下载: ${row.file_name || row.task_name}`)
}

function handleDeleteDownload(row: any) {
  ElMessageBox.confirm('确认删除该下载记录？', '删除确认', { type: 'warning' }).then(() => {
    downloads.value = downloads.value.filter((d: any) => d.id !== row.id)
    ElMessage.success('已删除')
  }).catch(() => {})
}

onMounted(() => {
  fetchFaults()
  fetchTasks()
  fetchDownloads()
})
</script>

<style scoped>
.fault-detail {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
}
.fault-detail p {
  margin: 6px 0;
}
</style>
