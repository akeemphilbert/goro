<template>
  <div class="dashboard">
    <!-- Welcome Section -->
    <div class="welcome-section">
      <a-row :gutter="24">
        <a-col :span="24">
          <a-card class="welcome-card">
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
    </div>

    <!-- Quick Stats -->
    <div class="stats-section">
      <a-row :gutter="24">
        <a-col :xs="24" :sm="12" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Total Resources"
              :value="stats.totalResources"
              :value-style="{ color: '#667eea' }"
            >
              <template #prefix>
                <FileTextOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Containers"
              :value="stats.containers"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <FolderOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Shared Resources"
              :value="stats.sharedResources"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <ShareAltOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Storage Used"
              :value="stats.storageUsed"
              suffix="MB"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <DatabaseOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- Recent Activity -->
    <div class="activity-section">
      <a-row :gutter="24">
        <a-col :span="16">
          <a-card title="Recent Activity" class="activity-card">
            <a-timeline>
              <a-timeline-item
                v-for="activity in recentActivity"
                :key="activity.id"
                :color="activity.color"
              >
                <template #dot>
                  <component :is="activity.icon" />
                </template>
                <div class="activity-item">
                  <div class="activity-content">
                    <p class="activity-text">{{ activity.description }}</p>
                    <span class="activity-time">{{ activity.time }}</span>
                  </div>
                </div>
              </a-timeline-item>
            </a-timeline>
          </a-card>
        </a-col>
        <a-col :span="8">
          <a-card title="Quick Actions" class="quick-actions-card">
            <div class="quick-actions">
              <a-button 
                v-for="action in quickActions"
                :key="action.key"
                :type="action.type"
                block
                class="quick-action-btn"
                @click="handleQuickAction(action.key)"
              >
                <component :is="action.icon" />
                {{ action.label }}
              </a-button>
            </div>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- Data Type Overview -->
    <div class="data-types-section">
      <a-card title="Your Data Types" class="data-types-card">
        <a-row :gutter="16">
          <a-col 
            v-for="dataType in dataTypes"
            :key="dataType.type"
            :xs="24" 
            :sm="12" 
            :md="8" 
            :lg="6"
          >
            <div 
              class="data-type-card"
              @click="router.push(`/${dataType.route}`)"
            >
              <div class="data-type-icon">
                <component :is="dataType.icon" />
              </div>
              <h3>{{ dataType.name }}</h3>
              <p>{{ dataType.count }} items</p>
            </div>
          </a-col>
        </a-row>
      </a-card>
    </div>
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

// Mock data - replace with actual API calls
const stats = ref({
  totalResources: 1247,
  containers: 23,
  sharedResources: 89,
  storageUsed: 245
})

const recentActivity = ref([
  {
    id: 1,
    description: 'Added new recipe: "Chocolate Chip Cookies"',
    time: '2 hours ago',
    color: 'green',
    icon: 'BookOutlined'
  },
  {
    id: 2,
    description: 'Shared container "Work Documents" with team',
    time: '4 hours ago',
    color: 'blue',
    icon: 'ShareAltOutlined'
  },
  {
    id: 3,
    description: 'Updated contact information for John Smith',
    time: '1 day ago',
    color: 'orange',
    icon: 'ContactsOutlined'
  },
  {
    id: 4,
    description: 'Created new container "Personal Photos"',
    time: '2 days ago',
    color: 'purple',
    icon: 'FolderOutlined'
  }
])

const quickActions = ref([
  {
    key: 'upload',
    label: 'Upload File',
    icon: 'UploadOutlined',
    type: 'primary'
  },
  {
    key: 'create-container',
    label: 'Create Container',
    icon: 'PlusOutlined',
    type: 'default'
  },
  {
    key: 'invite-user',
    label: 'Invite User',
    icon: 'UserAddOutlined',
    type: 'default'
  },
  {
    key: 'settings',
    label: 'Settings',
    icon: 'SettingOutlined',
    type: 'default'
  }
])

const dataTypes = ref([
  {
    type: 'recipes',
    name: 'Recipes',
    count: 45,
    route: 'recipes',
    icon: 'BookOutlined'
  },
  {
    type: 'contacts',
    name: 'Contacts',
    count: 123,
    route: 'contacts',
    icon: 'ContactsOutlined'
  },
  {
    type: 'documents',
    name: 'Documents',
    count: 89,
    route: 'documents',
    icon: 'FileOutlined'
  },
  {
    type: 'photos',
    name: 'Photos',
    count: 234,
    route: 'photos',
    icon: 'PictureOutlined'
  }
])

const router = useRouter()

const handleQuickAction = (action: string) => {
  console.log('Quick action:', action)
  // Implement quick actions
  if (action === 'upload') {
    router.push('/upload')
  } else if (action === 'create-container') {
    router.push('/containers')
  } else if (action === 'invite-user') {
    router.push('/permissions')
  } else if (action === 'settings') {
    router.push('/settings')
  }
}
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.welcome-section {
  margin-bottom: 24px;
}

.welcome-card {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
  border: 1px solid rgba(102, 126, 234, 0.2);
}

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
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.welcome-text p {
  margin: 0;
  font-size: 16px;
  color: var(--text-color-secondary);
}

.welcome-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.stats-section {
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
}

.activity-section {
  margin-bottom: 24px;
}

.activity-card, .quick-actions-card {
  height: 400px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.activity-content {
  flex: 1;
}

.activity-text {
  margin: 0 0 4px 0;
  font-weight: 500;
}

.activity-time {
  color: var(--text-color-secondary);
  font-size: 12px;
}

.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.quick-action-btn {
  height: 48px;
  border-radius: var(--border-radius);
  font-weight: 500;
}

.data-types-section {
  margin-bottom: 24px;
}

.data-type-card {
  background: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 140px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.data-type-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover);
  border-color: var(--primary-color);
}

.data-type-icon {
  font-size: 32px;
  color: var(--primary-color);
  margin-bottom: 12px;
}

.data-type-card h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.data-type-card p {
  margin: 0;
  color: var(--text-color-secondary);
  font-size: 14px;
}

/* Responsive design */
@media (max-width: 768px) {
  .welcome-content {
    flex-direction: column;
    text-align: center;
  }
  
  .welcome-actions {
    justify-content: center;
  }
  
  .activity-card, .quick-actions-card {
    height: auto;
    margin-bottom: 16px;
  }
}
</style>
