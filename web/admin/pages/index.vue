<template>
  <div>
    <!-- Welcome Section -->
    <a-row :gutter="24" style="margin-bottom: 24px">
      <a-col :span="24">
        <a-card>
          <div class="welcome-content">
            <div class="welcome-text">
              <h1>Welcome to your Solid Pod</h1>
              <p>Your personal data space where you control your information</p>
            </div>
            <div class="welcome-actions">
              <a-button type="primary" size="large" @click="router.push('/all-resources')">
                <FileTextOutlined />
                Browse My Data
              </a-button>
              <a-button size="large" @click="router.push('/containers')">
                <FolderOutlined />
                Manage Containers
              </a-button>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- Statistics Section -->
    <a-row :gutter="24" style="margin-bottom: 24px">
      <a-col :xs="24" :sm="6">
        <a-card>
          <a-statistic
            title="Total Resources"
            :value="stats.totalResources"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <DatabaseOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="6">
        <a-card>
          <a-statistic
            title="Containers"
            :value="stats.containers"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <FolderOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="6">
        <a-card>
          <a-statistic
            title="Shared Resources"
            :value="stats.sharedResources"
            :value-style="{ color: '#722ed1' }"
          >
            <template #prefix>
              <ShareAltOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="6">
        <a-card>
          <a-statistic
            title="Storage Used"
            :value="stats.storageUsed"
            suffix="MB"
            :value-style="{ color: '#fa8c16' }"
          >
            <template #prefix>
              <PlusOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <!-- Data Types Overview -->
    <a-row :gutter="24" style="margin-bottom: 24px">
      <a-col :span="24">
        <a-card title="Data Types in Your Pod">
          <a-row :gutter="16">
            <a-col 
              v-for="dataType in dataTypes" 
              :key="dataType.name"
              :xs="24" 
              :sm="12" 
              :md="8" 
              :lg="6"
            >
              <a-card 
                hoverable
                class="data-type-card"
                @click="router.push(`/${dataType.route}`)"
              >
                <div class="data-type-icon">
                  <component :is="dataType.icon" />
                </div>
                <h3>{{ dataType.name }}</h3>
                <p>{{ dataType.count }} items</p>
              </a-card>
            </a-col>
          </a-row>
        </a-card>
      </a-col>
    </a-row>

    <!-- Recent Activity & Quick Actions -->
    <a-row :gutter="24">
      <a-col :xs="24" :lg="12">
        <a-card title="Recent Activity">
          <a-list 
            :data-source="recentActivity" 
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta
                  :description="item.description"
                >
                  <template #title>{{ item.title }}</template>
                  <template #avatar>
                    <a-avatar :style="{ backgroundColor: item.color }">
                      <component :is="item.icon" />
                    </a-avatar>
                  </template>
                </a-list-item-meta>
                <div>{{ item.time }}</div>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>
      
      <a-col :xs="24" :lg="12">
        <a-card title="Quick Actions">
          <a-space direction="vertical" style="width: 100%">
            <a-button 
              v-for="action in quickActions" 
              :key="action.title"
              type="default" 
              size="large" 
              block
              @click="action.action"
            >
              <component :is="action.icon" />
              {{ action.title }}
            </a-button>
          </a-space>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { 
  FileTextOutlined, 
  FolderOutlined, 
  ShareAltOutlined, 
  DatabaseOutlined,
  PlusOutlined,
  UploadOutlined,
  UserAddOutlined,
  SettingOutlined,
  BookOutlined,
  ContactsOutlined,
  FileOutlined,
  PictureOutlined
} from '@ant-design/icons-vue'

const router = useRouter()

// Mock data - replace with actual API calls
const stats = ref({
  totalResources: 1247,
  containers: 23,
  sharedResources: 156,
  storageUsed: 2.4
})

const dataTypes = ref([
  { name: 'Recipes', count: 42, icon: BookOutlined, route: 'recipes' },
  { name: 'Contacts', count: 128, icon: ContactsOutlined, route: 'contacts' },
  { name: 'Documents', count: 89, icon: FileOutlined, route: 'documents' },
  { name: 'Photos', count: 234, icon: PictureOutlined, route: 'photos' },
  { name: 'Files', count: 567, icon: FileTextOutlined, route: 'files' },
  { name: 'Folders', count: 23, icon: FolderOutlined, route: 'containers' }
])

const recentActivity = ref([
  {
    title: 'New recipe added',
    description: 'Chocolate Chip Cookies recipe was created',
    time: '2 hours ago',
    icon: BookOutlined,
    color: '#52c41a'
  },
  {
    title: 'Contact updated',
    description: 'John Doe contact information was modified',
    time: '5 hours ago', 
    icon: ContactsOutlined,
    color: '#1890ff'
  },
  {
    title: 'Document shared',
    description: 'Project proposal shared with team',
    time: '1 day ago',
    icon: ShareAltOutlined,
    color: '#722ed1'
  },
  {
    title: 'New container created',
    description: 'Work Documents container was created',
    time: '2 days ago',
    icon: FolderOutlined,
    color: '#fa8c16'
  }
])

const quickActions = ref([
  {
    title: 'Upload File',
    icon: UploadOutlined,
    action: () => console.log('Upload file')
  },
  {
    title: 'Create Container',
    icon: FolderOutlined,
    action: () => router.push('/containers')
  },
  {
    title: 'Invite User',
    icon: UserAddOutlined,
    action: () => router.push('/permissions')
  },
  {
    title: 'Settings',
    icon: SettingOutlined,
    action: () => router.push('/settings')
  }
])
</script>

<style scoped>
.welcome-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 24px;
}

.welcome-text h1 {
  margin: 0 0 8px 0;
  font-size: 32px;
  font-weight: 700;
}

.welcome-text p {
  margin: 0;
  font-size: 16px;
}

.welcome-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.data-type-icon {
  font-size: 32px;
  margin-bottom: 12px;
  color: #1890ff;
}

.data-type-card {
  text-align: center;
  padding: 24px;
  cursor: pointer;
}

.data-type-card h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
}

.data-type-card p {
  margin: 0;
  font-size: 14px;
}

/* Responsive design */
@media (max-width: 768px) {
  .welcome-content {
    text-align: center;
    flex-direction: column;
  }
  
  .welcome-actions {
    justify-content: center;
  }
}
</style>