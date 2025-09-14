<template>
  <div class="permissions-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1>
            <SafetyOutlined />
            Permissions & Access
          </h1>
          <p>Manage who can access your data and containers</p>
        </div>
        <div class="header-actions">
          <a-button type="primary" @click="showInviteUserModal">
            <UserAddOutlined />
            Invite User
          </a-button>
        </div>
      </div>
    </div>

    <!-- Permission Overview -->
    <div class="overview-section">
      <a-row :gutter="24">
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="overview-card">
            <a-statistic
              title="Total Users"
              :value="permissionStats.totalUsers"
              :value-style="{ color: '#667eea' }"
            >
              <template #prefix>
                <UserOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="overview-card">
            <a-statistic
              title="Active Permissions"
              :value="permissionStats.activePermissions"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <SafetyOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="overview-card">
            <a-statistic
              title="Pending Invites"
              :value="permissionStats.pendingInvites"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <ClockCircleOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="overview-card">
            <a-statistic
              title="Shared Resources"
              :value="permissionStats.sharedResources"
              :value-style="{ color: '#ff4d4f' }"
            >
              <template #prefix>
                <ShareAltOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- Tabs for different permission views -->
    <div class="tabs-section">
      <a-tabs v-model:activeKey="activeTab" class="permission-tabs">
        <a-tab-pane key="users" tab="Users & Permissions">
          <div class="users-section">
            <a-table
              :columns="userColumns"
              :data-source="users"
              :pagination="{ pageSize: 10 }"
              row-key="id"
              class="users-table"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'avatar'">
                  <a-avatar :src="record.avatar" :size="32">
                    {{ record.name.charAt(0).toUpperCase() }}
                  </a-avatar>
                </template>
                
                <template v-if="column.key === 'permissions'">
                  <a-space wrap>
                    <a-tag 
                      v-for="permission in record.permissions"
                      :key="permission"
                      :color="getPermissionColor(permission)"
                    >
                      {{ permission }}
                    </a-tag>
                  </a-space>
                </template>
                
                <template v-if="column.key === 'status'">
                  <a-tag :color="record.status === 'active' ? 'green' : 'orange'">
                    {{ record.status }}
                  </a-tag>
                </template>
                
                <template v-if="column.key === 'actions'">
                  <a-space>
                    <a-button type="text" @click="editUserPermissions(record)">
                      <EditOutlined />
                    </a-button>
                    <a-button type="text" @click="revokeAccess(record)">
                      <StopOutlined />
                    </a-button>
                    <a-popconfirm
                      title="Are you sure you want to remove this user?"
                      @confirm="removeUser(record.id)"
                    >
                      <a-button type="text" danger>
                        <DeleteOutlined />
                      </a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </template>
            </a-table>
          </div>
        </a-tab-pane>
        
        <a-tab-pane key="containers" tab="Container Permissions">
          <div class="containers-section">
            <a-table
              :columns="containerColumns"
              :data-source="containerPermissions"
              :pagination="{ pageSize: 10 }"
              row-key="id"
              class="containers-table"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'permissions'">
                  <a-space wrap>
                    <a-tag 
                      v-for="permission in record.permissions"
                      :key="permission"
                      :color="getPermissionColor(permission)"
                    >
                      {{ permission }}
                    </a-tag>
                  </a-space>
                </template>
                
                <template v-if="column.key === 'actions'">
                  <a-space>
                    <a-button type="text" @click="editContainerPermissions(record)">
                      <EditOutlined />
                    </a-button>
                    <a-button type="text" @click="shareContainer(record)">
                      <ShareAltOutlined />
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
          </div>
        </a-tab-pane>
        
        <a-tab-pane key="invitations" tab="Pending Invitations">
          <div class="invitations-section">
            <a-table
              :columns="invitationColumns"
              :data-source="invitations"
              :pagination="{ pageSize: 10 }"
              row-key="id"
              class="invitations-table"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'status'">
                  <a-tag :color="getInvitationStatusColor(record.status)">
                    {{ record.status }}
                  </a-tag>
                </template>
                
                <template v-if="column.key === 'actions'">
                  <a-space>
                    <a-button 
                      v-if="record.status === 'pending'"
                      type="text" 
                      @click="resendInvitation(record)"
                    >
                      <SendOutlined />
                      Resend
                    </a-button>
                    <a-button 
                      v-if="record.status === 'pending'"
                      type="text" 
                      danger
                      @click="cancelInvitation(record.id)"
                    >
                      <CloseOutlined />
                      Cancel
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>

    <!-- Invite User Modal -->
    <a-modal
      v-model:open="inviteModalVisible"
      title="Invite User"
      width="600px"
      @ok="handleInviteSubmit"
      @cancel="handleInviteCancel"
    >
      <a-form
        ref="inviteFormRef"
        :model="inviteForm"
        :rules="inviteFormRules"
        layout="vertical"
      >
        <a-form-item label="Email Address" name="email">
          <a-input v-model:value="inviteForm.email" placeholder="user@example.com" />
        </a-form-item>
        
        <a-form-item label="Permissions" name="permissions">
          <a-checkbox-group v-model:value="inviteForm.permissions">
            <a-checkbox value="read">Read Access</a-checkbox>
            <a-checkbox value="write">Write Access</a-checkbox>
            <a-checkbox value="admin">Admin Access</a-checkbox>
          </a-checkbox-group>
        </a-form-item>
        
        <a-form-item label="Containers" name="containers">
          <a-tree-select
            v-model:value="inviteForm.containers"
            :tree-data="containerSelectOptions"
            placeholder="Select containers to share"
            tree-checkable
            multiple
            tree-default-expand-all
          />
        </a-form-item>
        
        <a-form-item label="Message" name="message">
          <a-textarea 
            v-model:value="inviteForm.message" 
            placeholder="Optional message for the invitation..."
            :rows="3"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Edit Permissions Modal -->
    <a-modal
      v-model:open="permissionModalVisible"
      title="Edit Permissions"
      width="500px"
      @ok="handlePermissionSubmit"
      @cancel="handlePermissionCancel"
    >
      <a-form
        ref="permissionFormRef"
        :model="permissionForm"
        layout="vertical"
      >
        <a-form-item label="User">
          <a-input v-model:value="permissionForm.userName" disabled />
        </a-form-item>
        
        <a-form-item label="Permissions">
          <a-checkbox-group v-model:value="permissionForm.permissions">
            <a-checkbox value="read">Read Access</a-checkbox>
            <a-checkbox value="write">Write Access</a-checkbox>
            <a-checkbox value="admin">Admin Access</a-checkbox>
          </a-checkbox-group>
        </a-form-item>
        
        <a-form-item label="Containers">
          <a-tree-select
            v-model:value="permissionForm.containers"
            :tree-data="containerSelectOptions"
            placeholder="Select containers"
            tree-checkable
            multiple
            tree-default-expand-all
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { 
  SafetyOutlined, 
  UserAddOutlined,
  UserOutlined,
  ClockCircleOutlined,
  ShareAltOutlined,
  EditOutlined,
  StopOutlined,
  DeleteOutlined,
  SendOutlined,
  CloseOutlined
} from '@ant-design/icons-vue'

// Reactive state
const activeTab = ref('users')
const inviteModalVisible = ref(false)
const permissionModalVisible = ref(false)
const editingUser = ref(null)

// Mock data
const permissionStats = ref({
  totalUsers: 8,
  activePermissions: 24,
  pendingInvites: 3,
  sharedResources: 89
})

const users = ref([
  {
    id: 1,
    name: 'John Smith',
    email: 'john@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=John',
    permissions: ['read', 'write'],
    containers: ['Documents', 'Work'],
    status: 'active',
    lastActive: '2024-01-15'
  },
  {
    id: 2,
    name: 'Sarah Johnson',
    email: 'sarah@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah',
    permissions: ['read'],
    containers: ['Photos'],
    status: 'active',
    lastActive: '2024-01-14'
  },
  {
    id: 3,
    name: 'Mike Chen',
    email: 'mike@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Mike',
    permissions: ['read', 'write', 'admin'],
    containers: ['All'],
    status: 'active',
    lastActive: '2024-01-15'
  }
])

const containerPermissions = ref([
  {
    id: 1,
    name: 'Documents',
    type: 'documents',
    permissions: ['read', 'write'],
    sharedWith: 3,
    lastModified: '2024-01-15'
  },
  {
    id: 2,
    name: 'Work Documents',
    type: 'documents',
    permissions: ['read'],
    sharedWith: 2,
    lastModified: '2024-01-14'
  },
  {
    id: 3,
    name: 'Photos',
    type: 'media',
    permissions: ['read'],
    sharedWith: 1,
    lastModified: '2024-01-13'
  }
])

const invitations = ref([
  {
    id: 1,
    email: 'newuser@example.com',
    permissions: ['read'],
    containers: ['Documents'],
    status: 'pending',
    sentAt: '2024-01-10',
    expiresAt: '2024-01-17'
  },
  {
    id: 2,
    email: 'collaborator@example.com',
    permissions: ['read', 'write'],
    containers: ['Work Documents'],
    status: 'pending',
    sentAt: '2024-01-12',
    expiresAt: '2024-01-19'
  }
])

// Table columns
const userColumns = [
  { title: '', dataIndex: 'avatar', key: 'avatar', width: 60 },
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Permissions', dataIndex: 'permissions', key: 'permissions' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Last Active', dataIndex: 'lastActive', key: 'lastActive' },
  { title: 'Actions', key: 'actions', width: 150 }
]

const containerColumns = [
  { title: 'Container', dataIndex: 'name', key: 'name' },
  { title: 'Type', dataIndex: 'type', key: 'type' },
  { title: 'Permissions', dataIndex: 'permissions', key: 'permissions' },
  { title: 'Shared With', dataIndex: 'sharedWith', key: 'sharedWith' },
  { title: 'Last Modified', dataIndex: 'lastModified', key: 'lastModified' },
  { title: 'Actions', key: 'actions', width: 120 }
]

const invitationColumns = [
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Permissions', dataIndex: 'permissions', key: 'permissions' },
  { title: 'Status', dataIndex: 'status', key: 'status' },
  { title: 'Sent', dataIndex: 'sentAt', key: 'sentAt' },
  { title: 'Expires', dataIndex: 'expiresAt', key: 'expiresAt' },
  { title: 'Actions', key: 'actions', width: 150 }
]

// Form data
const inviteForm = reactive({
  email: '',
  permissions: ['read'],
  containers: [],
  message: ''
})

const permissionForm = reactive({
  userName: '',
  permissions: [],
  containers: []
})

const inviteFormRules = {
  email: [
    { required: true, message: 'Please enter email address' },
    { type: 'email', message: 'Please enter valid email' }
  ],
  permissions: [{ required: true, message: 'Please select at least one permission' }]
}

// Mock container options
const containerSelectOptions = ref([
  { title: 'Documents', value: 'documents', key: 'documents' },
  { title: 'Work Documents', value: 'work-docs', key: 'work-docs' },
  { title: 'Photos', value: 'photos', key: 'photos' },
  { title: 'Media', value: 'media', key: 'media' }
])

// Methods
const showInviteUserModal = () => {
  resetInviteForm()
  inviteModalVisible.value = true
}

const editUserPermissions = (user: any) => {
  editingUser.value = user
  Object.assign(permissionForm, {
    userName: user.name,
    permissions: user.permissions,
    containers: user.containers
  })
  permissionModalVisible.value = true
}

const revokeAccess = (user: any) => {
  console.log('Revoking access for:', user.name)
  // Implement revoke access
}

const removeUser = (userId: number) => {
  const index = users.value.findIndex(u => u.id === userId)
  if (index > -1) {
    users.value.splice(index, 1)
  }
}

const editContainerPermissions = (container: any) => {
  console.log('Editing container permissions:', container.name)
  // Implement container permission editing
}

const shareContainer = (container: any) => {
  console.log('Sharing container:', container.name)
  // Implement container sharing
}

const resendInvitation = (invitation: any) => {
  console.log('Resending invitation to:', invitation.email)
  // Implement resend invitation
}

const cancelInvitation = (invitationId: number) => {
  const index = invitations.value.findIndex(i => i.id === invitationId)
  if (index > -1) {
    invitations.value.splice(index, 1)
  }
}

const handleInviteSubmit = () => {
  console.log('Sending invitation:', inviteForm)
  // Implement invitation sending
  inviteModalVisible.value = false
  resetInviteForm()
}

const handleInviteCancel = () => {
  inviteModalVisible.value = false
  resetInviteForm()
}

const handlePermissionSubmit = () => {
  console.log('Updating permissions:', permissionForm)
  // Implement permission update
  permissionModalVisible.value = false
}

const handlePermissionCancel = () => {
  permissionModalVisible.value = false
}

const resetInviteForm = () => {
  Object.assign(inviteForm, {
    email: '',
    permissions: ['read'],
    containers: [],
    message: ''
  })
}

const getPermissionColor = (permission: string) => {
  const colors = {
    read: 'blue',
    write: 'green',
    admin: 'red'
  }
  return colors[permission] || 'default'
}

const getInvitationStatusColor = (status: string) => {
  const colors = {
    pending: 'orange',
    accepted: 'green',
    expired: 'red'
  }
  return colors[status] || 'default'
}
</script>

<style scoped>
.permissions-page {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left h1 {
  margin: 0 0 4px 0;
  font-size: 28px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-color);
}

.header-left p {
  margin: 0;
  color: var(--text-color-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.overview-section {
  margin-bottom: 24px;
}

.overview-card {
  text-align: center;
  transition: all 0.3s ease;
}

.overview-card:hover {
  transform: translateY(-2px);
}

.tabs-section {
  margin-bottom: 24px;
}

.permission-tabs {
  background: var(--card-background);
  border-radius: var(--border-radius-lg);
  padding: 24px;
}

.users-table,
.containers-table,
.invitations-table {
  background: transparent;
}

/* Responsive design */
@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    align-items: stretch;
  }
  
  .header-actions {
    justify-content: stretch;
  }
  
  .permission-tabs {
    padding: 16px;
  }
}
</style>
