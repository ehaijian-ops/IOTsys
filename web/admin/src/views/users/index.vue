<template>
  <div class="users-page">
    <div class="page-header">
      <h3>用户管理</h3>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon> 添加用户
      </el-button>
    </div>

    <!-- 用户列表 -->
    <el-card>
      <el-table :data="users" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="username" label="用户名" width="130" />
        <el-table-column prop="nickname" label="昵称" width="130" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" width="160">
          <template #default="{ row }">
            {{ row.last_login_at || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ row.created_at?.slice(0, 19).replace('T', ' ') }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="openResetPwd(row)">重置密码</el-button>
            <el-popconfirm
              title="确认删除该用户？"
              confirm-button-text="删除"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button type="danger" link size="small"
                  :disabled="row.role === 'super_admin'"
                >删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="fetchUsers"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户弹窗 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingUser ? '编辑用户' : '添加用户'"
      width="500px"
      :close-on-click-modal="false"
      @close="resetForm"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!editingUser" placeholder="英文/数字，2-64字符" />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="显示名称" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" style="width:100%">
            <el-option
              v-for="r in availableRoles"
              :key="r.value"
              :label="r.label"
              :value="r.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="选填" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" placeholder="选填" />
        </el-form-item>
        <el-form-item v-if="editingUser" label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ editingUser ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码弹窗 -->
    <el-dialog v-model="showResetPwdDialog" title="重置密码" width="400px" :close-on-click-modal="false">
      <el-form :model="resetPwdForm" label-width="80px">
        <el-form-item label="新密码">
          <el-input v-model="resetPwdForm.new_password" type="password" show-password placeholder="至少6位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetPwdDialog = false">取消</el-button>
        <el-button type="primary" @click="handleResetPwd">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import http from '@/api/index'

interface UserItem {
  id: number
  username: string
  nickname: string
  role: string
  email?: string
  phone?: string
  enabled: boolean
  last_login_at?: string
  created_at: string
}

interface RoleOption {
  value: string
  label: string
}

const loading = ref(false)
const submitting = ref(false)
const users = ref<UserItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const showCreateDialog = ref(false)
const showResetPwdDialog = ref(false)
const editingUser = ref<UserItem | null>(null)
const resetTargetUser = ref<UserItem | null>(null)
const formRef = ref()
const availableRoles = ref<RoleOption[]>([])

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  role: 'viewer',
  email: '',
  phone: '',
  enabled: true,
})

const resetPwdForm = reactive({
  new_password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur', min: 2, max: 64 }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur', min: 6, max: 64 }],
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

function roleLabel(role: string): string {
  const map: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    operator: '运维人员',
    viewer: '查看者',
  }
  return map[role] || role
}

function roleTagType(role: string): string {
  const map: Record<string, string> = {
    super_admin: 'danger',
    admin: 'warning',
    operator: 'success',
    viewer: 'info',
  }
  return map[role] || 'info'
}

async function fetchUsers() {
  loading.value = true
  try {
    const resp = await http.get('/users', { params: { page: page.value, page_size: pageSize.value } }) as any
    users.value = resp.data || []
    total.value = resp.total || 0
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function fetchRoles() {
  try {
    const resp = await http.get('/users/roles') as any
    availableRoles.value = resp.data || []
  } catch {
    // fallback
  }
}

function resetForm() {
  editingUser.value = null
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.role = 'viewer'
  form.email = ''
  form.phone = ''
  form.enabled = true
}

function openEdit(user: UserItem) {
  editingUser.value = user
  form.username = user.username
  form.nickname = user.nickname
  form.role = user.role
  form.email = user.email || ''
  form.phone = user.phone || ''
  form.enabled = user.enabled
  showCreateDialog.value = true
}

function openResetPwd(user: UserItem) {
  resetTargetUser.value = user
  resetPwdForm.new_password = ''
  showResetPwdDialog.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (editingUser.value) {
      await http.put(`/users/${editingUser.value.id}`, {
        nickname: form.nickname,
        role: form.role,
        email: form.email || undefined,
        phone: form.phone || undefined,
        enabled: form.enabled,
      })
      ElMessage.success('用户更新成功')
    } else {
      await http.post('/users', {
        username: form.username,
        password: form.password,
        nickname: form.nickname,
        role: form.role,
        email: form.email || undefined,
        phone: form.phone || undefined,
      })
      ElMessage.success('用户创建成功')
    }
    showCreateDialog.value = false
    fetchUsers()
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || '操作失败'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await http.delete(`/users/${id}`)
    ElMessage.success('用户删除成功')
    fetchUsers()
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || '删除失败'
    ElMessage.error(msg)
  }
}

async function handleResetPwd() {
  if (resetPwdForm.new_password.length < 6) {
    ElMessage.error('密码长度不能少于6位')
    return
  }
  try {
    await http.put(`/users/${resetTargetUser.value?.id}/reset-password`, {
      new_password: resetPwdForm.new_password,
    })
    ElMessage.success('密码重置成功')
    showResetPwdDialog.value = false
  } catch (e: any) {
    const msg = e?.response?.data?.message || e?.message || '重置失败'
    ElMessage.error(msg)
  }
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<style scoped>
.users-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.page-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
