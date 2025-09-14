<template>
  <div class="contacts-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1>
            <ContactsOutlined />
            Contacts
          </h1>
          <p>Manage your contact information</p>
        </div>
        <div class="header-actions">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search contacts..."
            class="search-input"
            @search="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input-search>
          <a-button type="primary" @click="showAddContactModal">
            <PlusOutlined />
            Add Contact
          </a-button>
        </div>
      </div>
    </div>

    <!-- View Toggle -->
    <div class="view-toggle">
      <a-radio-group v-model:value="viewMode" button-style="solid">
        <a-radio-button value="grid">
          <AppstoreOutlined />
          Grid
        </a-radio-button>
        <a-radio-button value="list">
          <UnorderedListOutlined />
          List
        </a-radio-button>
      </a-radio-group>
    </div>

    <!-- Contacts Grid/List -->
    <div v-if="viewMode === 'grid'" class="contacts-grid">
      <a-row :gutter="[24, 24]">
        <a-col 
          v-for="contact in filteredContacts"
          :key="contact.id"
          :xs="24" 
          :sm="12" 
          :md="8" 
          :lg="6"
        >
          <a-card 
            class="contact-card"
            hoverable
            @click="viewContact(contact)"
          >
            <div class="contact-avatar">
              <a-avatar 
                :size="64" 
                :src="contact.avatar"
                :style="{ backgroundColor: getAvatarColor(contact.name) }"
              >
                {{ contact.name.charAt(0).toUpperCase() }}
              </a-avatar>
            </div>
            
            <div class="contact-info">
              <h3 class="contact-name">{{ contact.name }}</h3>
              <p class="contact-title">{{ contact.title }}</p>
              <p class="contact-company">{{ contact.company }}</p>
              
              <div class="contact-details">
                <div v-if="contact.email" class="contact-detail">
                  <MailOutlined />
                  <span>{{ contact.email }}</span>
                </div>
                <div v-if="contact.phone" class="contact-detail">
                  <PhoneOutlined />
                  <span>{{ contact.phone }}</span>
                </div>
                <div v-if="contact.location" class="contact-detail">
                  <EnvironmentOutlined />
                  <span>{{ contact.location }}</span>
                </div>
              </div>
            </div>
            
            <template #actions>
              <a-button type="text" @click.stop="editContact(contact)">
                <EditOutlined />
              </a-button>
              <a-button type="text" @click.stop="callContact(contact)">
                <PhoneOutlined />
              </a-button>
              <a-button type="text" @click.stop="emailContact(contact)">
                <MailOutlined />
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this contact?"
                @confirm="deleteContact(contact.id)"
              >
                <a-button type="text" danger @click.stop>
                  <DeleteOutlined />
                </a-button>
              </a-popconfirm>
            </template>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- Contacts List View -->
    <div v-else class="contacts-list">
      <a-table
        :columns="contactColumns"
        :data-source="filteredContacts"
        :pagination="{ pageSize: 10 }"
        row-key="id"
        class="contacts-table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'avatar'">
            <a-avatar 
              :src="record.avatar"
              :style="{ backgroundColor: getAvatarColor(record.name) }"
            >
              {{ record.name.charAt(0).toUpperCase() }}
            </a-avatar>
          </template>
          
          <template v-if="column.key === 'actions'">
            <a-space>
              <a-button type="text" @click="viewContact(record)">
                <EyeOutlined />
              </a-button>
              <a-button type="text" @click="editContact(record)">
                <EditOutlined />
              </a-button>
              <a-button type="text" @click="callContact(record)">
                <PhoneOutlined />
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this contact?"
                @confirm="deleteContact(record.id)"
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

    <!-- Empty State -->
    <div v-if="filteredContacts.length === 0" class="empty-state">
      <a-empty description="No contacts found">
        <template #image>
          <ContactsOutlined style="font-size: 64px; color: #d9d9d9;" />
        </template>
        <a-button type="primary" @click="showAddContactModal">
          <PlusOutlined />
          Add Your First Contact
        </a-button>
      </a-empty>
    </div>

    <!-- Add/Edit Contact Modal -->
    <a-modal
      v-model:open="contactModalVisible"
      :title="editingContact ? 'Edit Contact' : 'Add New Contact'"
      width="600px"
      @ok="handleContactSubmit"
      @cancel="handleContactCancel"
    >
      <a-form
        ref="contactFormRef"
        :model="contactForm"
        :rules="contactFormRules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Full Name" name="name">
              <a-input v-model:value="contactForm.name" placeholder="Enter full name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Title" name="title">
              <a-input v-model:value="contactForm.title" placeholder="Job title" />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Company" name="company">
              <a-input v-model:value="contactForm.company" placeholder="Company name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Email" name="email">
              <a-input v-model:value="contactForm.email" placeholder="email@example.com" />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Phone" name="phone">
              <a-input v-model:value="contactForm.phone" placeholder="Phone number" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Location" name="location">
              <a-input v-model:value="contactForm.location" placeholder="City, Country" />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="Notes" name="notes">
          <a-textarea 
            v-model:value="contactForm.notes" 
            placeholder="Additional notes..."
            :rows="3"
          />
        </a-form-item>
        
        <a-form-item label="Avatar URL" name="avatar">
          <a-input v-model:value="contactForm.avatar" placeholder="Avatar image URL" />
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
  AppstoreOutlined,
  UnorderedListOutlined,
  EditOutlined,
  PhoneOutlined,
  MailOutlined,
  DeleteOutlined,
  EyeOutlined,
  EnvironmentOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const viewMode = ref('grid')
const contactModalVisible = ref(false)
const editingContact = ref(null)

// Mock contact data
const contacts = ref([
  {
    id: 1,
    name: 'John Smith',
    title: 'Software Engineer',
    company: 'Tech Corp',
    email: 'john@techcorp.com',
    phone: '+1 (555) 123-4567',
    location: 'San Francisco, CA',
    notes: 'Met at conference, interested in collaboration',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=John'
  },
  {
    id: 2,
    name: 'Sarah Johnson',
    title: 'Product Manager',
    company: 'Design Studio',
    email: 'sarah@designstudio.com',
    phone: '+1 (555) 987-6543',
    location: 'New York, NY',
    notes: 'Former colleague, great product sense',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Sarah'
  },
  {
    id: 3,
    name: 'Mike Chen',
    title: 'UX Designer',
    company: 'Creative Agency',
    email: 'mike@creative.com',
    phone: '+1 (555) 456-7890',
    location: 'Seattle, WA',
    notes: 'Freelance collaborator, excellent design skills',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Mike'
  },
  {
    id: 4,
    name: 'Emily Davis',
    title: 'Marketing Director',
    company: 'Startup Inc',
    email: 'emily@startup.com',
    phone: '+1 (555) 321-0987',
    location: 'Austin, TX',
    notes: 'Industry contact, potential partnership opportunity',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Emily'
  }
])

// Table columns for list view
const contactColumns = [
  {
    title: '',
    dataIndex: 'avatar',
    key: 'avatar',
    width: 60
  },
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name',
    sorter: (a: any, b: any) => a.name.localeCompare(b.name)
  },
  {
    title: 'Title',
    dataIndex: 'title',
    key: 'title'
  },
  {
    title: 'Company',
    dataIndex: 'company',
    key: 'company'
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
    title: 'Actions',
    key: 'actions',
    width: 150
  }
]

// Form data
const contactForm = reactive({
  name: '',
  title: '',
  company: '',
  email: '',
  phone: '',
  location: '',
  notes: '',
  avatar: ''
})

const contactFormRules = {
  name: [{ required: true, message: 'Please enter contact name' }],
  email: [
    { required: true, message: 'Please enter email' },
    { type: 'email', message: 'Please enter valid email' }
  ]
}

// Computed properties
const filteredContacts = computed(() => {
  return contacts.value.filter(contact => {
    const searchLower = searchQuery.value.toLowerCase()
    return !searchQuery.value || 
      contact.name.toLowerCase().includes(searchLower) ||
      contact.title.toLowerCase().includes(searchLower) ||
      contact.company.toLowerCase().includes(searchLower) ||
      contact.email.toLowerCase().includes(searchLower)
  })
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
}

const showAddContactModal = () => {
  editingContact.value = null
  resetForm()
  contactModalVisible.value = true
}

const editContact = (contact: any) => {
  editingContact.value = contact
  Object.assign(contactForm, contact)
  contactModalVisible.value = true
}

const viewContact = (contact: any) => {
  console.log('Viewing contact:', contact)
  // Navigate to contact detail view
}

const callContact = (contact: any) => {
  if (contact.phone) {
    window.open(`tel:${contact.phone}`)
  }
}

const emailContact = (contact: any) => {
  if (contact.email) {
    window.open(`mailto:${contact.email}`)
  }
}

const deleteContact = (id: number) => {
  const index = contacts.value.findIndex(c => c.id === id)
  if (index > -1) {
    contacts.value.splice(index, 1)
  }
}

const handleContactSubmit = () => {
  if (editingContact.value) {
    // Update existing contact
    const index = contacts.value.findIndex(c => c.id === editingContact.value.id)
    if (index > -1) {
      contacts.value[index] = { ...contactForm, id: editingContact.value.id }
    }
  } else {
    // Add new contact
    const newContact = {
      ...contactForm,
      id: Date.now()
    }
    contacts.value.push(newContact)
  }
  
  contactModalVisible.value = false
  resetForm()
}

const handleContactCancel = () => {
  contactModalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(contactForm, {
    name: '',
    title: '',
    company: '',
    email: '',
    phone: '',
    location: '',
    notes: '',
    avatar: ''
  })
}

const getAvatarColor = (name: string) => {
  const colors = ['#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe']
  const index = name.charCodeAt(0) % colors.length
  return colors[index]
}
</script>

<style scoped>
.contacts-page {
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
  flex-wrap: wrap;
}

.search-input {
  width: 300px;
}

.view-toggle {
  margin-bottom: 24px;
  text-align: right;
}

.contacts-grid {
  margin-bottom: 24px;
}

.contact-card {
  height: 100%;
  text-align: center;
  transition: all 0.3s ease;
}

.contact-card:hover {
  transform: translateY(-4px);
}

.contact-avatar {
  margin-bottom: 16px;
}

.contact-info {
  margin-bottom: 16px;
}

.contact-name {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
}

.contact-title {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: var(--text-color-secondary);
  font-weight: 500;
}

.contact-company {
  margin: 0 0 12px 0;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.contact-details {
  text-align: left;
}

.contact-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.contact-detail .anticon {
  color: var(--primary-color);
}

.contacts-list {
  margin-bottom: 24px;
}

.contacts-table {
  background: var(--card-background);
  border-radius: var(--border-radius-lg);
  overflow: hidden;
}

.empty-state {
  text-align: center;
  padding: 48px 0;
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
  
  .search-input {
    width: 100%;
  }
  
  .view-toggle {
    text-align: center;
  }
  
  .contacts-table {
    overflow-x: auto;
  }
}
</style>
