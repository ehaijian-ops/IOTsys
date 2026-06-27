<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 角色管理 -->
      <el-tab-pane label="角色管理" name="roles">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openRoleDialog()">
            <el-icon><Plus /></el-icon> 新增角色
          </el-button>
        </div>
        <el-card>
          <el-table :data="roles" stripe v-loading="roleLoading">
            <el-table-column prop="name" label="角色名称" width="180" />
            <el-table-column prop="code" label="角色编码" width="160" />
            <el-table-column prop="data_scope" label="数据权限" width="140">
              <template #default="{ row }">
                <el-tag size="small">{{ { all: '全部数据', site: '站点级', self: '本人数据' }[row.data_scope] || row.data_scope }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="菜单权限" min-width="300">
              <template #default="{ row }">
                <el-tag v-for="m in (row.permissions || [])" :key="m" size="small" style="margin: 2px">{{ m }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openRoleDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteRole(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 菜单管理 -->
      <el-tab-pane label="菜单管理" name="menus">
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openMenuDialog()">
            <el-icon><Plus /></el-icon> 新增菜单
          </el-button>
        </div>
        <el-card>
          <el-table :data="menus" stripe row-key="id" v-loading="menuLoading">
            <el-table-column prop="title" label="菜单名称" width="200" />
            <el-table-column prop="icon" label="图标" width="120" />
            <el-table-column prop="path" label="路由路径" width="220" />
            <el-table-column prop="component" label="组件路径" min-width="260" show-overflow-tooltip />
            <el-table-column prop="sort_order" label="排序" width="80" align="center" />
            <el-table-column prop="hidden" label="隐藏" width="80" align="center">
              <template #default="{ row }">
                <el-switch :model-value="row.hidden" disabled size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openMenuDialog(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteMenu(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 登陆日志 -->
      <el-tab-pane label="登陆日志" name="login-logs">
        <el-card>
          <el-form :inline="true" :model="loginLogQuery">
            <el-form-item label="用户名">
              <el-input v-model="loginLogQuery.username" placeholder="用户名" clearable style="width: 160px" />
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="loginLogQuery.status" placeholder="全部" clearable style="width: 120px">
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchLoginLogs">查询</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="loginLogs" stripe v-loading="loginLogLoading">
            <el-table-column prop="username" label="用户名" width="140" />
            <el-table-column prop="ip_address" label="IP地址" width="160" />
            <el-table-column prop="user_agent" label="浏览器" min-width="240" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message" label="备注" width="160" show-overflow-tooltip />
            <el-table-column prop="created_at" label="登录时间" width="170" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 系统日志 -->
      <el-tab-pane label="系统日志" name="sys-logs">
        <el-card>
          <el-form :inline="true" :model="sysLogQuery">
            <el-form-item label="操作类型">
              <el-select v-model="sysLogQuery.action_type" placeholder="全部" clearable style="width: 140px">
                <el-option label="创建" value="create" />
                <el-option label="更新" value="update" />
                <el-option label="删除" value="delete" />
                <el-option label="查询" value="query" />
              </el-select>
            </el-form-item>
            <el-form-item label="操作人">
              <el-input v-model="sysLogQuery.username" placeholder="用户名" clearable style="width: 160px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchSysLogs">查询</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="sysLogs" stripe v-loading="sysLogLoading">
            <el-table-column prop="username" label="操作人" width="140" />
            <el-table-column prop="action_type" label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ { create: '创建', update: '更新', delete: '删除', query: '查询', login: '登录' }[row.action_type] || row.action_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="target" label="操作对象" width="160" />
            <el-table-column prop="details" label="操作详情" min-width="260" show-overflow-tooltip />
            <el-table-column prop="ip_address" label="IP地址" width="160" />
            <el-table-column prop="created_at" label="操作时间" width="170" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 角色弹窗 -->
    <el-dialog v-model="roleDialog.visible" :title="roleDialog.isEdit ? '编辑角色' : '新增角色'" width="560px">
      <el-form :model="roleDialog.form" label-width="100px">
        <el-form-item label="角色名称">
          <el-input v-model="roleDialog.form.name" placeholder="如：运维人员" />
        </el-form-item>
        <el-form-item label="角色编码">
          <el-input v-model="roleDialog.form.code" placeholder="如：operator" :disabled="roleDialog.isEdit" />
        </el-form-item>
        <el-form-item label="数据权限">
          <el-select v-model="roleDialog.form.data_scope" style="width: 100%">
            <el-option label="全部数据" value="all" />
            <el-option label="站点级" value="site" />
            <el-option label="本人数据" value="self" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜单权限">
          <el-tree
            :data="availableMenus"
            show-checkbox
            node-key="key"
            :default-checked-keys="roleDialog.form.permissions || []"
            ref="roleTreeRef"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveRole" :loading="roleSaving">保存</el-button>
      </template>
    </el-dialog>

    <!-- 菜单弹窗 -->
    <el-dialog v-model="menuDialog.visible" :title="menuDialog.isEdit ? '编辑菜单' : '新增菜单'" width="560px">
      <el-form :model="menuDialog.form" label-width="100px">
        <el-form-item label="上级菜单">
          <el-select v-model="menuDialog.form.parent_id" clearable placeholder="顶级菜单" style="width: 100%">
            <el-option v-for="m in menus" :key="m.id" :label="m.title" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜单名称">
          <el-input v-model="menuDialog.form.title" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="menuDialog.form.icon" placeholder="如：Monitor" />
        </el-form-item>
        <el-form-item label="路由路径">
          <el-input v-model="menuDialog.form.path" placeholder="如：/devices" />
        </el-form-item>
        <el-form-item label="组件路径">
          <el-input v-model="menuDialog.form.component" placeholder="如：views/devices/index.vue" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="menuDialog.form.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="是否隐藏">
          <el-switch v-model="menuDialog.form.hidden" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="menuDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doSaveMenu" :loading="menuSaving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRoles, createRole, updateRole, deleteRole } from '@/api/system2'
import { getMenus, createMenu, updateMenu, deleteMenu } from '@/api/system2'
import { getLoginLogs, getSystemLogs } from '@/api/system2'

const activeTab = ref('roles')

// 角色
const roles = ref<any[]>([])
const roleLoading = ref(false)
const roleDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const roleSaving = ref(false)
const roleTreeRef = ref()

const availableMenus = [
  { key: 'dashboard', label: '仪表盘', children: [] },
  { key: 'devices', label: '设备管理', children: [
    { key: 'devices.list', label: '设备列表' },
    { key: 'devices.unregistered', label: '未注册设备' },
    { key: 'devices.advanced', label: '设备扩展' },
  ]},
  { key: 'orders', label: '订单管理', children: [] },
  { key: 'operations', label: '运营管理', children: [
    { key: 'operations.billing', label: '收费方案' },
    { key: 'operations.ads', label: '广告运营' },
  ]},
  { key: 'finance', label: '财务管理', children: [
    { key: 'finance.recharges', label: '财务中心' },
    { key: 'finance.agents', label: '代理运营' },
  ]},
  { key: 'cards', label: '卡片管理', children: [] },
  { key: 'maintenance', label: '运维管理', children: [] },
  { key: 'interconnect', label: '互联互通', children: [] },
  { key: 'system', label: '系统管理', children: [] },
]

async function fetchRoles() {
  roleLoading.value = true
  try {
    const res: any = await getRoles()
    roles.value = res?.data || res?.list || []
  } catch {} finally {
    roleLoading.value = false
  }
}

function openRoleDialog(row?: any) {
  roleDialog.isEdit = !!row
  roleDialog.form = row ? { ...row, permissions: row.permissions || [] } : { name: '', code: '', data_scope: 'all', permissions: [] }
  roleDialog.visible = true
}

async function doSaveRole() {
  roleSaving.value = true
  try {
    const checkedKeys: string[] = roleTreeRef.value?.getCheckedKeys?.() || []
    const form = { ...roleDialog.form, permissions: checkedKeys }
    if (roleDialog.isEdit) {
      await updateRole(form.id, form)
      ElMessage.success('角色已更新')
    } else {
      await createRole(form)
      ElMessage.success('角色已创建')
    }
    roleDialog.visible = false
    fetchRoles()
  } catch {} finally {
    roleSaving.value = false
  }
}

function handleDeleteRole(row: any) {
  ElMessageBox.confirm('确认删除该角色？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteRole(row.id)
    ElMessage.success('已删除')
    fetchRoles()
  }).catch(() => {})
}

// 菜单
const menus = ref<any[]>([])
const menuLoading = ref(false)
const menuDialog = reactive({ visible: false, isEdit: false, form: {} as any })
const menuSaving = ref(false)

async function fetchMenus() {
  menuLoading.value = true
  try {
    const res: any = await getMenus()
    menus.value = res?.data || res?.list || []
  } catch {} finally {
    menuLoading.value = false
  }
}

function openMenuDialog(row?: any) {
  menuDialog.isEdit = !!row
  menuDialog.form = row ? { ...row } : { parent_id: 0, title: '', icon: '', path: '', component: '', sort_order: 0, hidden: false }
  menuDialog.visible = true
}

async function doSaveMenu() {
  menuSaving.value = true
  try {
    if (menuDialog.isEdit) {
      await updateMenu(menuDialog.form.id, menuDialog.form)
      ElMessage.success('菜单已更新')
    } else {
      await createMenu(menuDialog.form)
      ElMessage.success('菜单已创建')
    }
    menuDialog.visible = false
    fetchMenus()
  } catch {} finally {
    menuSaving.value = false
  }
}

function handleDeleteMenu(row: any) {
  ElMessageBox.confirm('确认删除该菜单？', '删除确认', { type: 'warning' }).then(async () => {
    await deleteMenu(row.id)
    ElMessage.success('已删除')
    fetchMenus()
  }).catch(() => {})
}

// 登陆日志
const loginLogs = ref<any[]>([])
const loginLogQuery = reactive({ username: '', status: '' })
const loginLogLoading = ref(false)

async function fetchLoginLogs() {
  loginLogLoading.value = true
  try {
    const params: any = {}
    if (loginLogQuery.username) params.username = loginLogQuery.username
    if (loginLogQuery.status) params.status = loginLogQuery.status
    const res: any = await getLoginLogs(params)
    loginLogs.value = res?.data || res?.list || []
  } catch {} finally {
    loginLogLoading.value = false
  }
}

// 系统日志
const sysLogs = ref<any[]>([])
const sysLogQuery = reactive({ action_type: '', username: '' })
const sysLogLoading = ref(false)

async function fetchSysLogs() {
  sysLogLoading.value = true
  try {
    const params: any = {}
    if (sysLogQuery.action_type) params.action_type = sysLogQuery.action_type
    if (sysLogQuery.username) params.username = sysLogQuery.username
    const res: any = await getSystemLogs(params)
    sysLogs.value = res?.data || res?.list || []
  } catch {} finally {
    sysLogLoading.value = false
  }
}

onMounted(() => {
  fetchRoles()
  fetchMenus()
})
</script>
