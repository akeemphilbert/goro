<template>
  <div class="containers-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1>
            <FolderOutlined />
            Containers
          </h1>
          <p>Manage your data containers and organization</p>
        </div>
        <div class="header-actions">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search containers..."
            class="search-input"
            @search="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input-search>
          <a-button type="primary" @click="showCreateContainerModal">
            <PlusOutlined />
            Create Container
          </a-button>
        </div>
      </div>
    </div>

    <!-- Container Stats -->
    <div class="stats-section">
      <a-row :gutter="24">
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Total Containers"
              :value="stats.totalContainers"
              :value-style="{ color: '#667eea' }"
            >
              <template #prefix>
                <FolderOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Resources"
              :value="stats.totalResources"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <FileTextOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
          <a-card class="stat-card">
            <a-statistic
              title="Shared"
              :value="stats.sharedContainers"
              :value-style="{ color: '#faad14' }"
            >
              <template #prefix>
                <ShareAltOutlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="8" :md="6">
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

    <!-- Container Tree View -->
    <div class="containers-section">
      <a-card title="Container Structure" class="containers-card">
        <template #extra>
          <a-space>
            <a-button @click="expandAll">
              <ExpandOutlined />
              Expand All
            </a-button>
            <a-button @click="collapseAll">
              <ShrinkOutlined />
              Collapse All
            </a-button>
          </a-space>
        </template>
        
        <a-tree
          v-model:expandedKeys="expandedKeys"
          v-model:selectedKeys="selectedKeys"
          :tree-data="containerTree"
          :show-line="true"
          :show-icon="true"
          class="container-tree"
          @select="handleContainerSelect"
        >
          <template #title="{ title, key, type, resourceCount, isShared }">
            <div class="tree-node">
              <div class="node-info">
                <span class="node-title">{{ title }}</span>
                <div class="node-meta">
                  <a-tag v-if="type" :color="getTypeColor(type)">{{ type }}</a-tag>
                  <span v-if="resourceCount !== undefined" class="resource-count">
                    {{ resourceCount }} resources
                  </span>
                  <a-tag v-if="isShared" color="blue">Shared</a-tag>
                </div>
              </div>
              <div class="node-actions">
                <a-button type="text" size="small" @click.stop="editContainer(key)">
                  <EditOutlined />
                </a-button>
                <a-button type="text" size="small" @click.stop="shareContainer(key)">
                  <ShareAltOutlined />
                </a-button>
                <a-popconfirm
                  title="Are you sure you want to delete this container?"
                  @confirm="deleteContainer(key)"
                >
                  <a-button type="text" size="small" danger @click.stop>
                    <DeleteOutlined />
                  </a-button>
                </a-popconfirm>
              </div>
            </div>
          </template>
        </a-tree>
      </a-card>
    </div>

    <!-- Container Details -->
    <div v-if="selectedContainer" class="container-details">
      <a-card :title="`Container: ${selectedContainer.title}`" class="details-card">
        <template #extra>
          <a-space>
            <a-button @click="editContainer(selectedContainer.key)">
              <EditOutlined />
              Edit
            </a-button>
            <a-button @click="shareContainer(selectedContainer.key)">
              <ShareAltOutlined />
              Share
            </a-button>
            <a-button @click="addResource(selectedContainer.key)">
              <PlusOutlined />
              Add Resource
            </a-button>
          </a-space>
        </template>
        
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="Name">
            {{ selectedContainer.title }}
          </a-descriptions-item>
          <a-descriptions-item label="Type">
            <a-tag :color="getTypeColor(selectedContainer.type)">
              {{ selectedContainer.type }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Resources">
            {{ selectedContainer.resourceCount || 0 }}
          </a-descriptions-item>
          <a-descriptions-item label="Created">
            {{ selectedContainer.createdAt }}
          </a-descriptions-item>
          <a-descriptions-item label="Last Modified">
            {{ selectedContainer.updatedAt }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="selectedContainer.isShared ? 'blue' : 'green'">
              {{ selectedContainer.isShared ? 'Shared' : 'Private' }}
            </a-tag>
          </a-descriptions-item>
        </a-descriptions>
        
        <div v-if="selectedContainer.description" class="container-description">
          <h4>Description</h4>
          <p>{{ selectedContainer.description }}</p>
        </div>
      </a-card>
    </div>

    <!-- Create/Edit Container Modal -->
    <a-modal
      v-model:open="containerModalVisible"
      :title="editingContainer ? 'Edit Container' : 'Create New Container'"
      width="600px"
      @ok="handleContainerSubmit"
      @cancel="handleContainerCancel"
    >
      <a-form
        ref="containerFormRef"
        :model="containerForm"
        :rules="containerFormRules"
        layout="vertical"
      >
        <a-form-item label="Container Name" name="name">
          <a-input v-model:value="containerForm.name" placeholder="Enter container name" />
        </a-form-item>
        
        <a-form-item label="Parent Container" name="parentId">
          <a-tree-select
            v-model:value="containerForm.parentId"
            :tree-data="containerSelectOptions"
            placeholder="Select parent container (optional)"
            tree-default-expand-all
            allow-clear
          />
        </a-form-item>
        
        <a-form-item label="Type" name="type">
          <a-select v-model:value="containerForm.type" placeholder="Select container type">
            <a-select-option value="general">General</a-select-option>
            <a-select-option value="documents">Documents</a-select-option>
            <a-select-option value="media">Media</a-select-option>
            <a-select-option value="data">Data</a-select-option>
            <a-select-option value="shared">Shared</a-select-option>
          </a-select>
        </a-form-item>
        
        <a-form-item label="Description" name="description">
          <a-textarea 
            v-model:value="containerForm.description" 
            placeholder="Describe this container..."
            :rows="3"
          />
        </a-form-item>
        
        <a-form-item label="Visibility" name="isShared">
          <a-radio-group v-model:value="containerForm.isShared">
            <a-radio :value="false">Private</a-radio>
            <a-radio :value="true">Shared</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { 
  FolderOutlined, 
  SearchOutlined, 
  PlusOutlined,
  FileTextOutlined,
  ShareAltOutlined,
  DatabaseOutlined,
  EditOutlined,
  DeleteOutlined,
  ExpandOutlined,
  ShrinkOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const expandedKeys = ref(['root'])
const selectedKeys = ref([])
const containerModalVisible = ref(false)
const editingContainer = ref(null)

// Mock data
const stats = ref({
  totalContainers: 12,
  totalResources: 1247,
  sharedContainers: 3,
  storageUsed: 245
})

const containers = ref([
  {
    key: 'root',
    title: 'My Pod',
    type: 'root',
    resourceCount: 1247,
    isShared: false,
    createdAt: '2024-01-01',
    updatedAt: '2024-01-15',
    description: 'Root container for all personal data'
  },
  {
    key: 'documents',
    title: 'Documents',
    type: 'documents',
    resourceCount: 89,
    isShared: false,
    createdAt: '2024-01-02',
    updatedAt: '2024-01-14',
    description: 'Personal and work documents'
  },
  {
    key: 'work-docs',
    title: 'Work Documents',
    type: 'documents',
    resourceCount: 45,
    isShared: true,
    createdAt: '2024-01-03',
    updatedAt: '2024-01-13',
    description: 'Shared work documents with team'
  },
  {
    key: 'media',
    title: 'Media',
    type: 'media',
    resourceCount: 234,
    isShared: false,
    createdAt: '2024-01-04',
    updatedAt: '2024-01-12',
    description: 'Photos, videos, and audio files'
  },
  {
    key: 'photos',
    title: 'Photos',
    type: 'media',
    resourceCount: 189,
    isShared: false,
    createdAt: '2024-01-05',
    updatedAt: '2024-01-11',
    description: 'Personal photo collection'
  },
  {
    key: 'data',
    title: 'Data',
    type: 'data',
    resourceCount: 156,
    isShared: false,
    createdAt: '2024-01-06',
    updatedAt: '2024-01-10',
    description: 'Structured data and datasets'
  }
])

// Form data
const containerForm = reactive({
  name: '',
  parentId: null,
  type: 'general',
  description: '',
  isShared: false
})

const containerFormRules = {
  name: [{ required: true, message: 'Please enter container name' }],
  type: [{ required: true, message: 'Please select container type' }]
}

// Computed properties
const containerTree = computed(() => {
  const buildTree = (items: any[], parentId = null) => {
    return items
      .filter(item => item.parentId === parentId)
      .map(item => ({
        key: item.key,
        title: item.title,
        type: item.type,
        resourceCount: item.resourceCount,
        isShared: item.isShared,
        children: buildTree(items, item.key)
      }))
  }
  
  // Mock tree structure
  return [
    {
      key: 'root',
      title: 'My Pod',
      type: 'root',
      resourceCount: 1247,
      isShared: false,
      children: [
        {
          key: 'documents',
          title: 'Documents',
          type: 'documents',
          resourceCount: 89,
          isShared: false,
          children: [
            {
              key: 'work-docs',
              title: 'Work Documents',
              type: 'documents',
              resourceCount: 45,
              isShared: true
            }
          ]
        },
        {
          key: 'media',
          title: 'Media',
          type: 'media',
          resourceCount: 234,
          isShared: false,
          children: [
            {
              key: 'photos',
              title: 'Photos',
              type: 'media',
              resourceCount: 189,
              isShared: false
            }
          ]
        },
        {
          key: 'data',
          title: 'Data',
          type: 'data',
          resourceCount: 156,
          isShared: false
        }
      ]
    }
  ]
})

const containerSelectOptions = computed(() => {
  const flattenTree = (nodes: any[]): any[] => {
    return nodes.reduce((acc, node) => {
      acc.push({
        title: node.title,
        value: node.key,
        key: node.key,
        children: node.children ? flattenTree(node.children) : []
      })
      return acc
    }, [])
  }
  
  return flattenTree(containerTree.value)
})

const selectedContainer = computed(() => {
  if (selectedKeys.value.length === 0) return null
  
  const findContainer = (nodes: any[], key: string): any => {
    for (const node of nodes) {
      if (node.key === key) return node
      if (node.children) {
        const found = findContainer(node.children, key)
        if (found) return found
      }
    }
    return null
  }
  
  return findContainer(containerTree.value, selectedKeys.value[0])
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
  // Implement search functionality
}

const expandAll = () => {
  const getAllKeys = (nodes: any[]): string[] => {
    return nodes.reduce((keys, node) => {
      keys.push(node.key)
      if (node.children) {
        keys.push(...getAllKeys(node.children))
      }
      return keys
    }, [])
  }
  
  expandedKeys.value = getAllKeys(containerTree.value)
}

const collapseAll = () => {
  expandedKeys.value = ['root']
}

const handleContainerSelect = (selectedKeys: string[]) => {
  // Handle container selection
}

const showCreateContainerModal = () => {
  editingContainer.value = null
  resetForm()
  containerModalVisible.value = true
}

const editContainer = (key: string) => {
  const container = containers.value.find(c => c.key === key)
  if (container) {
    editingContainer.value = container
    Object.assign(containerForm, {
      name: container.title,
      parentId: null, // Would need to determine parent
      type: container.type,
      description: container.description,
      isShared: container.isShared
    })
    containerModalVisible.value = true
  }
}

const shareContainer = (key: string) => {
  console.log('Sharing container:', key)
  // Implement sharing functionality
}

const deleteContainer = (key: string) => {
  const index = containers.value.findIndex(c => c.key === key)
  if (index > -1) {
    containers.value.splice(index, 1)
  }
}

const addResource = (key: string) => {
  console.log('Adding resource to container:', key)
  // Navigate to resource creation
}

const handleContainerSubmit = () => {
  if (editingContainer.value) {
    // Update existing container
    const index = containers.value.findIndex(c => c.key === editingContainer.value.key)
    if (index > -1) {
      containers.value[index] = {
        ...containers.value[index],
        title: containerForm.name,
        type: containerForm.type,
        description: containerForm.description,
        isShared: containerForm.isShared,
        updatedAt: new Date().toISOString().split('T')[0]
      }
    }
  } else {
    // Add new container
    const newContainer = {
      key: `container-${Date.now()}`,
      title: containerForm.name,
      type: containerForm.type,
      resourceCount: 0,
      isShared: containerForm.isShared,
      createdAt: new Date().toISOString().split('T')[0],
      updatedAt: new Date().toISOString().split('T')[0],
      description: containerForm.description
    }
    containers.value.push(newContainer)
  }
  
  containerModalVisible.value = false
  resetForm()
}

const handleContainerCancel = () => {
  containerModalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(containerForm, {
    name: '',
    parentId: null,
    type: 'general',
    description: '',
    isShared: false
  })
}

const getTypeColor = (type: string) => {
  const colors = {
    root: 'purple',
    documents: 'blue',
    media: 'green',
    data: 'orange',
    shared: 'cyan',
    general: 'default'
  }
  return colors[type] || 'default'
}
</script>

<style scoped>
.containers-page {
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

.stats-section {
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.containers-section {
  margin-bottom: 24px;
}

.containers-card {
  min-height: 400px;
}

.container-tree {
  background: transparent;
}

.tree-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 4px 0;
}

.node-info {
  flex: 1;
}

.node-title {
  font-weight: 500;
  color: var(--text-color);
}

.node-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.resource-count {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.tree-node:hover .node-actions {
  opacity: 1;
}

.container-details {
  margin-bottom: 24px;
}

.details-card {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
}

.container-description {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
}

.container-description h4 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.container-description p {
  margin: 0;
  color: var(--text-color-secondary);
  line-height: 1.6;
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
  
  .tree-node {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .node-actions {
    opacity: 1;
    align-self: flex-end;
  }
}
</style>
