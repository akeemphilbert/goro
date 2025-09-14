<template>
  <div>
    <!-- Page Header -->
    <a-page-header
      title="Permissions & Invitations"
      sub-title="Manage access to your data"
    />

    <!-- Tabs for different sections -->
    <a-tabs v-model:activeKey="activeTab">
      <a-tab-pane key="permissions" tab="User Permissions">
        <a-card>
          <template #extra>
            <a-button type="primary" @click="showInviteUserModal">
              <UserAddOutlined />
              Invite User
            </a-button>
          </template>
          
          <a-table 
            :columns="permissionColumns" 
            :data-source="userPermissions"
            :pagination="{ pageSize: 10 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'user'">
                <div style="display: flex; align-items: center; gap: 12px;">
                  <a-avatar :src="record.avatar">
                    {{ record.name.charAt(0) }}
                  </a-avatar>
                  <div>
                    <div style="font-weight: 500;">{{ record.name }}</div>
                    <div style="color: #666; font-size: 12px;">{{ record.email }}</div>
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'access'">
                <a-tag :color="getAccessColor(record.access)">
                  {{ record.access }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="text" @click="editPermissions(record)">
                    <EditOutlined />
                  </a-button>
                  <a-button type="text" danger @click="revokeAccess(record)">
                    <DeleteOutlined />
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-tab-pane>

      <a-tab-pane key="invitations" tab="Pending Invitations">
        <a-card>
          <a-table 
            :columns="invitationColumns" 
            :data-source="pendingInvitations"
            :pagination="{ pageSize: 10 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'status'">
                <a-tag :color="getStatusColor(record.status)">
                  {{ record.status }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'access'">
                <a-tag :color="getAccessColor(record.access)">
                  {{ record.access }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="text" @click="resendInvitation(record)">
                    <MailOutlined />
                  </a-button>
                  <a-button type="text" danger @click="cancelInvitation(record)">
                    <DeleteOutlined />
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-tab-pane>

      <a-tab-pane key="requests" tab="Access Requests">
        <a-card>
          <a-table 
            :columns="requestColumns" 
            :data-source="accessRequests"
            :pagination="{ pageSize: 10 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'user'">
                <div style="display: flex; align-items: center; gap: 12px;">
                  <a-avatar :src="record.avatar">
                    {{ record.name.charAt(0) }}
                  </a-avatar>
                  <div>
                    <div style="font-weight: 500;">{{ record.name }}</div>
                    <div style="color: #666; font-size: 12px;">{{ record.email }}</div>
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'requestedAccess'">
                <a-tag :color="getAccessColor(record.requestedAccess)">
                  {{ record.requestedAccess }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-space>
                  <a-button type="primary" size="small" @click="approveRequest(record)">
                    Approve
                  </a-button>
                  <a-button danger size="small" @click="denyRequest(record)">
                    Deny
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-tab-pane>
    </a-tabs>

    <!-- Invite User Modal -->
    <a-modal
      v-model:open="inviteUserModalVisible"
      title="Invite User"
      width="500px"
      @ok="handleInviteUser"
      @cancel="inviteUserModalVisible = false"
    >
      <a-form ref="inviteUserForm" :model="newInvitation" layout="vertical">
        <a-form-item label="Email Address" name="email" :rules="[{ required: true, type: 'email' }]">
          <a-input v-model:value="newInvitation.email" />
        </a-form-item>
        <a-form-item label="Access Level" name="access" :rules="[{ required: true }]">
          <a-select v-model:value="newInvitation.access">
            <a-select-option value="read">Read Only</a-select-option>
            <a-select-option value="write">Read & Write</a-select-option>
            <a-select-option value="admin">Admin</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Resource" name="resource">
          <a-tree-select
            v-model:value="newInvitation.resource"
            :tree-data="resourceOptions"
            placeholder="Select specific resource (optional)"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="Message" name="message">
          <a-textarea 
            v-model:value="newInvitation.message" 
            :rows="3"
            placeholder="Optional invitation message..."
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import {
  UserAddOutlined,
  EditOutlined,
  DeleteOutlined,
  MailOutlined
} from '@ant-design/icons-vue'

// Reactive state
const activeTab = ref('permissions')
const inviteUserModalVisible = ref(false)

const newInvitation = reactive({
  email: '',
  access: 'read',
  resource: '',
  message: ''
})

// Table columns
const permissionColumns = [
  { title: 'User', key: 'user', width: 200 },
  { title: 'Access Level', key: 'access', width: 120 },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Grant Date', dataIndex: 'grantDate', key: 'grantDate', width: 120 },
  { title: 'Actions', key: 'actions', width: 100 }
]

const invitationColumns = [
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Access Level', key: 'access', width: 120 },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Status', key: 'status', width: 100 },
  { title: 'Sent Date', dataIndex: 'sentDate', key: 'sentDate', width: 120 },
  { title: 'Actions', key: 'actions', width: 100 }
]

const requestColumns = [
  { title: 'User', key: 'user', width: 200 },
  { title: 'Requested Access', key: 'requestedAccess', width: 150 },
  { title: 'Resource', dataIndex: 'resource', key: 'resource' },
  { title: 'Request Date', dataIndex: 'requestDate', key: 'requestDate', width: 120 },
  { title: 'Actions', key: 'actions', width: 150 }
]

// Mock data
const userPermissions = ref([
  {
    id: 1,
    name: 'Alice Johnson',
    email: 'alice@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Alice',
    access: 'admin',
    resource: 'All Resources',
    grantDate: '2024-01-15'
  },
  {
    id: 2,
    name: 'Bob Smith',
    email: 'bob@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Bob',
    access: 'write',
    resource: '/personal/photos',
    grantDate: '2024-02-20'
  }
])

const pendingInvitations = ref([
  {
    id: 1,
    email: 'charlie@example.com',
    access: 'read',
    resource: '/personal/documents',
    status: 'pending',
    sentDate: '2024-03-01'
  },
  {
    id: 2,
    email: 'diana@example.com',
    access: 'write',
    resource: '/work/projects',
    status: 'expired',
    sentDate: '2024-02-15'
  }
])

const accessRequests = ref([
  {
    id: 1,
    name: 'Eve Wilson',
    email: 'eve@example.com',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Eve',
    requestedAccess: 'read',
    resource: '/personal/recipes',
    requestDate: '2024-03-05'
  }
])

const resourceOptions = ref([
  {
    title: 'All Resources',
    value: '/',
    children: [
      {
        title: 'Personal',
        value: '/personal',
        children: [
          { title: 'Documents', value: '/personal/documents' },
          { title: 'Photos', value: '/personal/photos' },
          { title: 'Recipes', value: '/personal/recipes' }
        ]
      },
      {
        title: 'Work',
        value: '/work',
        children: [
          { title: 'Projects', value: '/work/projects' },
          { title: 'Meetings', value: '/work/meetings' }
        ]
      }
    ]
  }
])

// Methods
const showInviteUserModal = () => {
  inviteUserModalVisible.value = true
}

const handleInviteUser = () => {
  console.log('Inviting user:', newInvitation)
  inviteUserModalVisible.value = false
  
  // Reset form
  Object.assign(newInvitation, {
    email: '',
    access: 'read',
    resource: '',
    message: ''
  })
}

const editPermissions = (user: any) => {
  console.log('Editing permissions for:', user)
}

const revokeAccess = (user: any) => {
  console.log('Revoking access for:', user)
}

const resendInvitation = (invitation: any) => {
  console.log('Resending invitation:', invitation)
}

const cancelInvitation = (invitation: any) => {
  console.log('Canceling invitation:', invitation)
}

const approveRequest = (request: any) => {
  console.log('Approving request:', request)
}

const denyRequest = (request: any) => {
  console.log('Denying request:', request)
}

const getAccessColor = (access: string): string => {
  const colors: Record<string, string> = {
    admin: 'red',
    write: 'orange',
    read: 'blue'
  }
  return colors[access] || 'default'
}

const getStatusColor = (status: string): string => {
  const colors: Record<string, string> = {
    pending: 'orange',
    expired: 'red',
    accepted: 'green'
  }
  return colors[status] || 'default'
}
</script>