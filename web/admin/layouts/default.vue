<template>
  <a-layout class="main-layout">
    <!-- Header with Navigation -->
    <a-layout-header class="app-header">
      <div class="header-content">
        <div class="logo-section">
          <a-avatar 
            :size="40" 
            :style="{ background: 'linear-gradient(135deg, #667eea, #764ba2)' }"
          >
            <template #icon>
              <DatabaseOutlined />
            </template>
          </a-avatar>
          <h1 class="app-title">Solid Pod</h1>
        </div>
        
        <div class="header-actions">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search your pod..."
            class="search-input"
            @search="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input-search>
          
          <a-dropdown v-if="isAuthenticated" placement="bottomRight">
            <a-button type="text" class="user-menu-trigger">
              <a-avatar :src="user?.avatar" :size="32">
                {{ user?.name?.charAt(0) }}
              </a-avatar>
              <span class="user-name">{{ user?.name }}</span>
              <DownOutlined />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile" @click="router.push('/profile')">
                  <UserOutlined />
                  Profile
                </a-menu-item>
                <a-menu-item key="settings" @click="router.push('/settings')">
                  <SettingOutlined />
                  Settings
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  Logout
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
          
          <a-button 
            v-else 
            type="primary" 
            @click="handleLogin"
            class="login-btn"
          >
            <GoogleOutlined />
            Login with Google
          </a-button>
        </div>
      </div>
    </a-layout-header>

    <a-layout class="main-content">
      <!-- Sidebar Navigation -->
      <a-layout-sider 
        v-if="isAuthenticated"
        :width="280" 
        class="sidebar"
        :collapsed="sidebarCollapsed"
        @collapse="sidebarCollapsed = $event"
      >
        <div class="sidebar-content">
          <a-menu
            v-model:selectedKeys="selectedKeys"
            mode="inline"
            class="sidebar-menu"
            @click="handleMenuClick"
          >
            <a-menu-item key="dashboard">
              <template #icon>
                <DashboardOutlined />
              </template>
              Dashboard
            </a-menu-item>
            
            <a-menu-item-group title="My Data">
              <a-menu-item key="all-resources">
                <template #icon>
                  <FileTextOutlined />
                </template>
                All Resources
              </a-menu-item>
              <a-menu-item key="recipes">
                <template #icon>
                  <BookOutlined />
                </template>
                Recipes
              </a-menu-item>
              <a-menu-item key="contacts">
                <template #icon>
                  <ContactsOutlined />
                </template>
                Contacts
              </a-menu-item>
              <a-menu-item key="documents">
                <template #icon>
                  <FileOutlined />
                </template>
                Documents
              </a-menu-item>
              <a-menu-item key="photos">
                <template #icon>
                  <PictureOutlined />
                </template>
                Photos
              </a-menu-item>
            </a-menu-item-group>
            
            <a-menu-item-group title="Management">
              <a-menu-item key="containers">
                <template #icon>
                  <FolderOutlined />
                </template>
                Containers
              </a-menu-item>
              <a-menu-item key="permissions">
                <template #icon>
                  <SafetyOutlined />
                </template>
                Permissions
              </a-menu-item>
              <a-menu-item key="invitations">
                <template #icon>
                  <UserAddOutlined />
                </template>
                Invitations
              </a-menu-item>
            </a-menu-item-group>
          </a-menu>
        </div>
      </a-layout-sider>

      <!-- Main Content Area -->
      <a-layout-content class="content-area">
        <div class="content-wrapper">
          <slot />
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  DatabaseOutlined, 
  SearchOutlined, 
  DownOutlined, 
  UserOutlined, 
  SettingOutlined, 
  LogoutOutlined,
  GoogleOutlined,
  DashboardOutlined,
  FileTextOutlined,
  BookOutlined,
  ContactsOutlined,
  FileOutlined,
  PictureOutlined,
  FolderOutlined,
  SafetyOutlined,
  UserAddOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const sidebarCollapsed = ref(false)
const selectedKeys = ref(['dashboard'])

// Mock user data - replace with actual auth state
const user = ref<{
  name: string
  email: string
  avatar: string
} | null>({
  name: 'John Doe',
  email: 'john@example.com',
  avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=John'
})

const isAuthenticated = computed(() => !!user.value)

// Methods
const handleSearch = (value: string) => {
  console.log('Searching for:', value)
  // Implement search functionality
}

const handleLogin = () => {
  console.log('Initiating Google login')
  // Implement Google OAuth
}

const handleLogout = () => {
  console.log('Logging out')
  user.value = null
  // Implement logout
}

const router = useRouter()

const handleMenuClick = ({ key }: { key: string }) => {
  selectedKeys.value = [key]
  router.push(`/${key}`)
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.app-header {
  padding: 0 24px;
  height: 64px;
  line-height: 64px;
  position: sticky;
  top: 0;
  z-index: 1000;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1400px;
  margin: 0 auto;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.search-input {
  width: 300px;
  border-radius: 20px;
}

.user-menu-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-color);
}

.user-menu-trigger:hover {
  background: rgba(255, 255, 255, 0.2);
  color: var(--text-color);
}

.user-name {
  font-weight: 500;
}

.login-btn {
  border-radius: 20px;
  height: 40px;
  padding: 0 20px;
  font-weight: 500;
}

.main-content {
  max-width: 1400px;
  margin: 0 auto;
  background: transparent;
}

.sidebar {
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(10px);
  border-radius: 0 0 0 16px;
  box-shadow: var(--shadow);
  margin: 0 0 0 24px;
}

.sidebar-content {
  padding: 24px 0;
  height: 100%;
}

.sidebar-menu {
  border: none !important;
  background: transparent !important;
  padding: 0 16px;
}

.sidebar-menu .ant-menu-item {
  margin: 2px 0;
  border-radius: var(--border-radius);
  height: 40px;
  line-height: 40px;
}

.sidebar-menu .ant-menu-item-group-title {
  color: var(--text-color-secondary);
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 24px 0 8px 0;
  padding: 0 12px;
}

.content-area {
  background: transparent;
  padding: 0;
}

.content-wrapper {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px 0 0 0;
  min-height: calc(100vh - 64px);
  padding: 24px;
  box-shadow: var(--shadow);
}

/* Responsive design */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .search-input {
    width: 200px;
  }
  
  .user-name {
    display: none;
  }
  
  .sidebar {
    margin: 0;
    border-radius: 0;
  }
  
  .content-wrapper {
    border-radius: 0;
    padding: 16px;
  }
}
</style>
