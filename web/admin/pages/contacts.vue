<template>
  <div>
    <!-- Page Header -->
    <a-page-header
      title="Contacts"
      sub-title="Manage your contact information"
    >
      <template #extra>
        <a-input-search
          v-model:value="searchQuery"
          placeholder="Search contacts..."
          style="width: 300px; margin-right: 16px"
          @search="handleSearch"
        />
        <a-button type="primary" @click="showAddContactModal">
          <PlusOutlined />
          Add Contact
        </a-button>
      </template>
    </a-page-header>

    <!-- View Toggle -->
    <a-card style="margin-bottom: 24px">
      <a-radio-group v-model:value="viewMode" button-style="solid">
        <a-radio-button value="grid">
          <AppstoreOutlined />
          Grid View
        </a-radio-button>
        <a-radio-button value="list">
          <UnorderedListOutlined />
          List View
        </a-radio-button>
      </a-radio-group>
    </a-card>

    <!-- Grid View -->
    <div v-if="viewMode === 'grid'">
      <a-row :gutter="[16, 16]">
        <a-col
          v-for="contact in filteredContacts"
          :key="contact.id"
          :xs="24"
          :sm="12"
          :md="8"
          :lg="6"
        >
          <a-card hoverable>
            <template #actions>
              <EditOutlined @click="editContact(contact)" />
              <ShareAltOutlined @click="shareContact(contact)" />
              <DeleteOutlined @click="deleteContact(contact)" />
            </template>
            <a-card-meta>
              <template #avatar>
                <a-avatar :size="64" :src="contact.avatar">
                  {{ contact.name.charAt(0) }}
                </a-avatar>
              </template>
              <template #title>{{ contact.name }}</template>
              <template #description>
                <div>
                  <div><MailOutlined /> {{ contact.email }}</div>
                  <div><PhoneOutlined /> {{ contact.phone }}</div>
                  <div v-if="contact.company"><BankOutlined /> {{ contact.company }}</div>
                </div>
              </template>
            </a-card-meta>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- List View -->
    <div v-else>
      <a-card>
        <a-table 
          :columns="tableColumns" 
          :data-source="filteredContacts"
          :pagination="{ pageSize: 10 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'avatar'">
              <a-avatar :src="record.avatar">
                {{ record.name.charAt(0) }}
              </a-avatar>
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="text" @click="editContact(record)">
                  <EditOutlined />
                </a-button>
                <a-button type="text" @click="shareContact(record)">
                  <ShareAltOutlined />
                </a-button>
                <a-button type="text" danger @click="deleteContact(record)">
                  <DeleteOutlined />
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>

    <!-- Add Contact Modal -->
    <a-modal
      v-model:open="addContactModalVisible"
      title="Add New Contact"
      width="500px"
      @ok="handleAddContact"
      @cancel="addContactModalVisible = false"
    >
      <a-form ref="addContactForm" :model="newContact" layout="vertical">
        <a-form-item label="Full Name" name="name" :rules="[{ required: true }]">
          <a-input v-model:value="newContact.name" />
        </a-form-item>
        <a-form-item label="Email" name="email" :rules="[{ required: true, type: 'email' }]">
          <a-input v-model:value="newContact.email" />
        </a-form-item>
        <a-form-item label="Phone" name="phone">
          <a-input v-model:value="newContact.phone" />
        </a-form-item>
        <a-form-item label="Company" name="company">
          <a-input v-model:value="newContact.company" />
        </a-form-item>
        <a-form-item label="Notes" name="notes">
          <a-textarea v-model:value="newContact.notes" :rows="3" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import {
  ContactsOutlined,
  SearchOutlined,
  PlusOutlined,
  EditOutlined,
  ShareAltOutlined,
  DeleteOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  MailOutlined,
  PhoneOutlined,
  BankOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const viewMode = ref('grid')
const addContactModalVisible = ref(false)

const newContact = reactive({
  name: '',
  email: '',
  phone: '',
  company: '',
  notes: ''
})

// Table columns for list view
const tableColumns = [
  {
    title: '',
    dataIndex: 'avatar',
    key: 'avatar',
    width: 60
  },
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: 'Email',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: 'Phone',
    dataIndex: 'phone',
    key: 'phone'
  },
  {
    title: 'Company',
    dataIndex: 'company',
    key: 'company'
  },
  {
    title: 'Actions',
    key: 'actions',
    width: 120
  }
]

// Mock data
const contacts = ref([
  {
    id: 1,
    name: 'John Doe',
    email: 'john@example.com',
    phone: '+1 234-567-8900',
    company: 'Tech Corp',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=John',
    notes: 'Project manager'
  },
  {
    id: 2,
    name: 'Jane Smith',
    email: 'jane@example.com',
    phone: '+1 234-567-8901',
    company: 'Design Studio',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Jane',
    notes: 'UX Designer'
  },
  {
    id: 3,
    name: 'Bob Johnson',
    email: 'bob@example.com',
    phone: '+1 234-567-8902',
    company: 'Marketing Inc',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Bob',
    notes: 'Marketing specialist'
  }
])

// Computed
const filteredContacts = computed(() => {
  return contacts.value.filter(contact => {
    const query = searchQuery.value.toLowerCase()
    return contact.name.toLowerCase().includes(query) ||
           contact.email.toLowerCase().includes(query) ||
           contact.company.toLowerCase().includes(query)
  })
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
}

const showAddContactModal = () => {
  addContactModalVisible.value = true
}

const handleAddContact = () => {
  console.log('Adding contact:', newContact)
  addContactModalVisible.value = false
  
  // Reset form
  Object.assign(newContact, {
    name: '',
    email: '',
    phone: '',
    company: '',
    notes: ''
  })
}

const editContact = (contact: any) => {
  console.log('Editing contact:', contact)
}

const shareContact = (contact: any) => {
  console.log('Sharing contact:', contact)
}

const deleteContact = (contact: any) => {
  console.log('Deleting contact:', contact)
}
</script>