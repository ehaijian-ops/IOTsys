<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '220px'" class="layout-aside">
      <div class="logo">
        <el-icon :size="24"><Monitor /></el-icon>
        <span v-show="!isCollapse" class="logo-text">IoT 充电桩平台</span>
      </div>

      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        background-color="#001529"
        text-color="#ffffffb3"
        active-text-color="#ffffff"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>

        <el-sub-menu index="device-group">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>设备管理</span>
          </template>
          <el-menu-item index="/devices">设备列表</el-menu-item>
        </el-sub-menu>

        <el-menu-item index="/commands">
          <el-icon><Operation /></el-icon>
          <span>指令管理</span>
        </el-menu-item>

        <el-menu-item index="/alerts">
          <el-icon><Bell /></el-icon>
          <span>告警管理</span>
        </el-menu-item>

        <el-menu-item index="/sites">
          <el-icon><Location /></el-icon>
          <span>站点管理</span>
        </el-menu-item>

        <el-menu-item index="/protocols">
          <el-icon><Connection /></el-icon>
          <span>协议管理</span>
        </el-menu-item>

        <el-menu-item index="/reports">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据报表</span>
        </el-menu-item>

        <el-menu-item index="/logs">
          <el-icon><Tickets /></el-icon>
          <span>实时日志</span>
        </el-menu-item>

        <el-menu-item index="/system">
          <el-icon><Setting /></el-icon>
          <span>系统监控</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <el-header class="layout-header">
        <div class="header-left">
          <el-button
            type="text"
            @click="isCollapse = !isCollapse"
            class="collapse-btn"
          >
            <el-icon :size="20">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </el-button>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute">{{ currentRoute }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-badge :value="3" class="alert-badge">
            <el-icon :size="20"><Bell /></el-icon>
          </el-badge>
          <el-dropdown>
            <span class="user-info">
              <el-icon><UserFilled /></el-icon>
              <span>管理员</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人设置</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
const currentRoute = computed(() => route.meta.title as string)

function handleLogout() {
  localStorage.removeItem('access_token')
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background-color: #001529;
  overflow: hidden;
  transition: width 0.3s;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo-text {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
}

.layout-header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  padding: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.alert-badge {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  color: #606266;
}

.layout-main {
  background: #f0f2f5;
  min-height: calc(100vh - 60px);
}
</style>
