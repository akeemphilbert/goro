<template>
  <a-layout>
    <!-- Header with Navigation -->
    <a-layout-header>
      <div class="header-content">
        <div class="logo-section">
          <a-avatar 
            :size="40" 
            :style="{ backgroundColor: '#1890ff' }"
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
            style="width: 300px"
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
          >
            <GoogleOutlined />
            Login with Google
          </a-button>
        </div>
      </div>
    </a-layout-header>

    <a-layout>
      <!-- Sidebar Navigation -->
      <a-layout-sider 
        v-if="isAuthenticated"
        :width="280"
        :collapsed="sidebarCollapsed"
        @collapse="sidebarCollapsed = $event"
      >
        <div class="sidebar-content">
          <a-menu
            v-model:selectedKeys="selectedKeys"
            mode="inline"
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
      <a-layout-content>
        <slot />
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

const handleMenuClick = ({ key }: { key: string | number }) => {
  selectedKeys.value = [String(key)]
  router.push(`/${key}`)
}
</script>

<style scoped>


</style>