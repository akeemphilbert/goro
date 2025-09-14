<template>
  <div class="all-resources-page">
    <a-row :gutter="[24, 24]">
      <a-col :span="24">
        <a-card>
          <template #title>
            <a-space>
              <FolderOutlined />
              <span>All Resources Tree View</span>
            </a-space>
          </template>
          <template #extra>
            <a-space>
              <a-tooltip title="Create new resource">
                <a-button 
                  type="primary" 
                  :icon="h(PlusOutlined)"
                  @click="showCreateModal = true"
                >
                  Create Resource
                </a-button>
              </a-tooltip>
              <a-tooltip title="Refresh tree">
                <a-button 
                  :icon="h(ReloadOutlined)" 
                  :loading="loading"
                  @click="loadTree"
                >
                  Refresh
                </a-button>
              </a-tooltip>
              <a-tooltip title="Expand all nodes">
                <a-button 
                  :icon="h(ExpandOutlined)"
                  @click="expandAll"
                >
                  Expand All
                </a-button>
              </a-tooltip>
              <a-tooltip title="Collapse all nodes">
                <a-button 
                  :icon="h(CompressOutlined)"
                  @click="collapseAll"
                >
                  Collapse All
                </a-button>
              </a-tooltip>
            </a-space>
          </template>

          <!-- Search and filter controls -->
          <div class="controls-section">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-input
                  v-model:value="searchTerm"
                  placeholder="Search containers and resources..."
                  :prefix="h(SearchOutlined)"
                  allow-clear
                  @input="handleSearch"
                />
              </a-col>
              <a-col :span="6">
                <a-select
                  v-model:value="filterType"
                  placeholder="Filter by type"
                  style="width: 100%"
                  allow-clear
                  @change="handleFilter"
                >
                  <a-select-option value="all">All Types</a-select-option>
                  <a-select-option value="container">Containers Only</a-select-option>
                  <a-select-option value="resource">Resources Only</a-select-option>
                </a-select>
              </a-col>
              <a-col :span="6">
                <a-select
                  v-model:value="sortBy"
                  placeholder="Sort by"
                  style="width: 100%"
                  @change="handleSort"
                >
                  <a-select-option value="name">Name</a-select-option>
                  <a-select-option value="type">Type</a-select-option>
                  <a-select-option value="size">Size</a-select-option>
                </a-select>
              </a-col>
            </a-row>
          </div>

          <!-- Tree view -->
          <div class="tree-container">
            <a-spin :spinning="loading" tip="Loading resources...">
              <div v-if="error" class="error-state">
                <a-result
                  status="error"
                  title="Failed to load resources"
                  :sub-title="error"
                >
                  <template #extra>
                    <a-button type="primary" @click="loadTree">
                      Try Again
                    </a-button>
                  </template>
                </a-result>
              </div>
              
              <div v-else-if="filteredTreeData.length === 0 && !loading" class="empty-state">
                <a-empty
                  description="No resources found"
                  :image="Empty.PRESENTED_IMAGE_SIMPLE"
                >
                  <template #description>
                    <span v-if="searchTerm">
                      No resources match your search: "{{ searchTerm }}"
                    </span>
                    <span v-else>
                      No containers or resources are available
                    </span>
                  </template>
                  <a-button type="primary" @click="loadTree">
                    Refresh
                  </a-button>
                </a-empty>
              </div>

              <a-tree
                v-else
                v-model:expandedKeys="expandedKeys"
                v-model:selectedKeys="selectedKeys"
                :tree-data="filteredTreeData"
                :show-icon="true"
                :show-line="true"
                block-node
                @select="handleNodeSelect"
                @expand="handleNodeExpand"
              >
                <template #icon="{ type, contentType }">
                  <FolderOutlined v-if="type === 'container'" style="color: #1890ff;" />
                  <FileTextOutlined v-else-if="isTextFile(contentType)" style="color: #52c41a;" />
                  <FileImageOutlined v-else-if="isImageFile(contentType)" style="color: #722ed1;" />
                  <FilePdfOutlined v-else-if="isPdfFile(contentType)" style="color: #f5222d;" />
                  <FileOutlined v-else style="color: #8c8c8c;" />
                </template>

                <template #title="{ title, type, size, contentType }">
                  <div class="tree-node-content">
                    <span class="node-title">{{ title }}</span>
                    <div class="node-meta">
                      <a-tag 
                        :color="type === 'container' ? 'blue' : 'green'" 
                        size="small"
                      >
                        {{ type }}
                      </a-tag>
                      <span v-if="size !== undefined" class="file-size">
                        {{ formatFileSize(size) }}
                      </span>
                      <span v-if="contentType" class="content-type">
                        {{ contentType }}
                      </span>
                    </div>
                  </div>
                </template>
              </a-tree>
            </a-spin>
          </div>

          <!-- Statistics -->
          <div v-if="!loading && !error" class="statistics-section">
            <a-row :gutter="16">
              <a-col :span="8">
                <a-statistic
                  title="Total Containers"
                  :value="statistics.containers"
                  :prefix="h(FolderOutlined)"
                />
              </a-col>
              <a-col :span="8">
                <a-statistic
                  title="Total Resources"
                  :value="statistics.resources"
                  :prefix="h(FileOutlined)"
                />
              </a-col>
              <a-col :span="8">
                <a-statistic
                  title="Total Size"
                  :value="formatFileSize(statistics.totalSize)"
                  :prefix="h(DatabaseOutlined)"
                />
              </a-col>
            </a-row>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- Resource details modal -->
    <a-modal
      v-model:open="showDetailsModal"
      title="Resource Details"
      :footer="null"
      width="800px"
    >
      <div v-if="selectedResource">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="ID">
            {{ selectedResource.key }}
          </a-descriptions-item>
          <a-descriptions-item label="Type">
            <a-tag :color="selectedResource.type === 'container' ? 'blue' : 'green'">
              {{ selectedResource.type }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Title">
            {{ selectedResource.title }}
          </a-descriptions-item>
          <a-descriptions-item label="Content Type" v-if="selectedResource.contentType">
            {{ selectedResource.contentType }}
          </a-descriptions-item>
          <a-descriptions-item label="Size" v-if="selectedResource.size !== undefined">
            {{ formatFileSize(selectedResource.size) }}
          </a-descriptions-item>
          <a-descriptions-item label="Children" v-if="selectedResource.children">
            {{ selectedResource.children.length }} items
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>

    <!-- Create resource modal -->
    <a-modal
      v-model:open="showCreateModal"
      title="Create New Resource"
      :confirm-loading="createLoading"
      @ok="handleCreateResource"
      @cancel="resetCreateForm"
      width="800px"
    >
      <a-form
        :model="createForm"
        :label-col="{ span: 6 }"
        :wrapper-col="{ span: 18 }"
        layout="horizontal"
      >
        <a-form-item 
          label="Resource ID" 
          required
          :validate-status="createForm.id ? 'success' : 'error'"
          help="Unique identifier for the resource"
        >
          <a-input
            v-model:value="createForm.id"
            placeholder="Enter resource ID (e.g., my-document)"
            :disabled="createLoading"
          />
        </a-form-item>

        <a-form-item 
          label="Content Type" 
          required
          :validate-status="createForm.contentType ? 'success' : 'error'"
          help="MIME type of the resource"
        >
          <a-select
            v-model:value="createForm.contentType"
            placeholder="Select content type"
            show-search
            :disabled="createLoading"
            :filter-option="filterContentTypeOption"
          >
            <a-select-option 
              v-for="type in contentTypes" 
              :key="type.value" 
              :value="type.value"
            >
              <div class="content-type-option">
                <span class="content-type-label">{{ type.label }}</span>
                <span class="content-type-value">{{ type.value }}</span>
              </div>
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item 
          label="Content" 
          required
          :validate-status="createForm.content ? 'success' : 'error'"
          help="The content/data for the resource"
        >
          <a-textarea
            v-model:value="createForm.content"
            placeholder="Enter resource content..."
            :rows="12"
            :disabled="createLoading"
            show-count
            :maxlength="50000"
          />
        </a-form-item>

        <a-form-item label="Preview">
          <div class="content-preview">
            <a-typography-text type="secondary">
              Size: {{ formatFileSize(getContentSize(createForm.content || '')) }}
            </a-typography-text>
            <br>
            <a-typography-text type="secondary" v-if="createForm.contentType">
              Type: {{ createForm.contentType }}
            </a-typography-text>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { message, Empty } from 'ant-design-vue'
import type { Key } from 'ant-design-vue/es/_util/type'
import {
  FolderOutlined,
  FileOutlined,
  FileTextOutlined,
  FileImageOutlined,
  FilePdfOutlined,
  ReloadOutlined,
  ExpandOutlined,
  CompressOutlined,
  SearchOutlined,
  DatabaseOutlined,
  PlusOutlined
} from '@ant-design/icons-vue'
import { StorageClient, type TreeNode } from '~/services/resource'

// Reactive state
const loading = ref(true)
const error = ref<string | null>(null)
const treeData = ref<TreeNode[]>([])
const expandedKeys = ref<Key[]>([])
const selectedKeys = ref<Key[]>([])
const searchTerm = ref('')
const filterType = ref<string>('all')
const sortBy = ref<string>('name')
const showDetailsModal = ref(false)
const selectedResource = ref<TreeNode | null>(null)

// Create resource state
const showCreateModal = ref(false)
const createLoading = ref(false)
const createForm = ref({
  id: '',
  contentType: '',
  content: ''
})

// Storage client instance
const storageClient = new StorageClient({
  baseURL: 'http://localhost:8081', // Backend runs on port 8081
  timeout: 30000
})

// Content types for the dropdown
const contentTypes = ref([
  // Text formats
  { label: 'Plain Text', value: 'text/plain' },
  { label: 'HTML', value: 'text/html' },
  { label: 'CSS', value: 'text/css' },
  { label: 'JavaScript', value: 'application/javascript' },
  { label: 'JSON', value: 'application/json' },
  { label: 'XML', value: 'application/xml' },
  { label: 'CSV', value: 'text/csv' },
  { label: 'Markdown', value: 'text/markdown' },
  
  // RDF formats (Solid Pod specific)
  { label: 'JSON-LD', value: 'application/ld+json' },
  { label: 'Turtle', value: 'text/turtle' },
  { label: 'RDF/XML', value: 'application/rdf+xml' },
  { label: 'N-Triples', value: 'application/n-triples' },
  { label: 'N-Quads', value: 'application/n-quads' },
  
  // Images
  { label: 'JPEG Image', value: 'image/jpeg' },
  { label: 'PNG Image', value: 'image/png' },
  { label: 'GIF Image', value: 'image/gif' },
  { label: 'SVG Image', value: 'image/svg+xml' },
  { label: 'WebP Image', value: 'image/webp' },
  
  // Documents
  { label: 'PDF', value: 'application/pdf' },
  { label: 'Word Document', value: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' },
  { label: 'Excel Spreadsheet', value: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' },
  
  // Archives
  { label: 'ZIP Archive', value: 'application/zip' },
  { label: 'TAR Archive', value: 'application/x-tar' },
  
  // Others
  { label: 'Binary Data', value: 'application/octet-stream' },
  { label: 'Form Data', value: 'application/x-www-form-urlencoded' },
  { label: 'Multipart Form', value: 'multipart/form-data' }
])

// Computed properties
const filteredTreeData = computed(() => {
  let data = [...treeData.value]
  
  // Apply search filter
  if (searchTerm.value) {
    data = filterTreeBySearch(data, searchTerm.value.toLowerCase())
  }
  
  // Apply type filter
  if (filterType.value !== 'all') {
    data = filterTreeByType(data, filterType.value)
  }
  
  // Apply sorting
  if (sortBy.value) {
    data = sortTree(data, sortBy.value)
  }
  
  return data
})

const statistics = computed(() => {
  const stats = { containers: 0, resources: 0, totalSize: 0 }
  
  const countNodes = (nodes: TreeNode[]) => {
    for (const node of nodes) {
      if (node.type === 'container') {
        stats.containers++
      } else {
        stats.resources++
        if (node.size) {
          stats.totalSize += node.size
        }
      }
      
      if (node.children) {
        countNodes(node.children)
      }
    }
  }
  
  countNodes(treeData.value)
  return stats
})

// Methods
const loadTree = async () => {
  loading.value = true
  error.value = null
  
  try {
    const tree = await storageClient.getFullTree()
    treeData.value = tree
    
    // Auto-expand root level nodes
    expandedKeys.value = tree.map(node => node.key)
    
    message.success('Resources loaded successfully')
  } catch (err) {
    console.error('Failed to load tree:', err)
    error.value = err instanceof Error ? err.message : 'An unknown error occurred'
    message.error('Failed to load resources')
  } finally {
    loading.value = false
  }
}

const expandAll = () => {
  const getAllKeys = (nodes: TreeNode[]): Key[] => {
    const keys: Key[] = []
    for (const node of nodes) {
      keys.push(node.key)
      if (node.children) {
        keys.push(...getAllKeys(node.children))
      }
    }
    return keys
  }
  
  expandedKeys.value = getAllKeys(treeData.value)
}

const collapseAll = () => {
  expandedKeys.value = []
}

const handleSearch = () => {
  // Search is handled reactively through computed property
}

const handleFilter = () => {
  // Filter is handled reactively through computed property
}

const handleSort = () => {
  // Sort is handled reactively through computed property
}

const handleNodeSelect = (keys: Key[], { node }: any) => {
  if (keys.length > 0) {
    selectedResource.value = node
    showDetailsModal.value = true
  }
}

const handleNodeExpand = (keys: Key[]) => {
  expandedKeys.value = keys
}

// Create resource methods
const handleCreateResource = async () => {
  if (!createForm.value.id || !createForm.value.contentType || !createForm.value.content) {
    message.error('Please fill in all required fields')
    return
  }

  createLoading.value = true
  
  try {
    const response = await storageClient.createResource(
      createForm.value.content,
      {
        contentType: createForm.value.contentType
      },
      createForm.value.id
    )
    
    message.success(`Resource "${createForm.value.id}" created successfully`)
    
    // Reset form and close modal
    resetCreateForm()
    showCreateModal.value = false
    
    // Refresh the tree to show the new resource
    await loadTree()
    
  } catch (error) {
    console.error('Failed to create resource:', error)
    message.error(error instanceof Error ? error.message : 'Failed to create resource')
  } finally {
    createLoading.value = false
  }
}

const resetCreateForm = () => {
  createForm.value = {
    id: '',
    contentType: '',
    content: ''
  }
}

const filterContentTypeOption = (input: string, option: any) => {
  const label = option.children?.[0]?.children?.[0]?.children || ''
  const value = option.children?.[0]?.children?.[1]?.children || ''
  return label.toLowerCase().includes(input.toLowerCase()) ||
         value.toLowerCase().includes(input.toLowerCase())
}

// Helper functions
const filterTreeBySearch = (nodes: TreeNode[], search: string): TreeNode[] => {
  const filtered: TreeNode[] = []
  
  for (const node of nodes) {
    const titleMatch = node.title.toLowerCase().includes(search)
    const childrenMatch = node.children ? filterTreeBySearch(node.children, search) : []
    
    if (titleMatch || childrenMatch.length > 0) {
      filtered.push({
        ...node,
        children: childrenMatch.length > 0 ? childrenMatch : node.children
      })
    }
  }
  
  return filtered
}

const filterTreeByType = (nodes: TreeNode[], type: string): TreeNode[] => {
  const filtered: TreeNode[] = []
  
  for (const node of nodes) {
    if (node.type === type) {
      filtered.push(node)
    } else if (node.children) {
      const filteredChildren = filterTreeByType(node.children, type)
      if (filteredChildren.length > 0) {
        filtered.push({
          ...node,
          children: filteredChildren
        })
      }
    }
  }
  
  return filtered
}

const sortTree = (nodes: TreeNode[], sortBy: string): TreeNode[] => {
  const sorted = [...nodes].sort((a, b) => {
    switch (sortBy) {
      case 'name':
        return a.title.localeCompare(b.title)
      case 'type':
        return a.type.localeCompare(b.type)
      case 'size':
        return (a.size || 0) - (b.size || 0)
      default:
        return 0
    }
  })
  
  return sorted.map(node => ({
    ...node,
    children: node.children ? sortTree(node.children, sortBy) : undefined
  }))
}

const isTextFile = (contentType?: string): boolean => {
  if (!contentType) return false
  return contentType.startsWith('text/') || 
         contentType.includes('json') || 
         contentType.includes('xml')
}

const isImageFile = (contentType?: string): boolean => {
  if (!contentType) return false
  return contentType.startsWith('image/')
}

const isPdfFile = (contentType?: string): boolean => {
  if (!contentType) return false
  return contentType.includes('pdf')
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getContentSize = (content: string): number => {
  // Use Blob to calculate the byte size in a browser-compatible way
  return new Blob([content], { type: 'text/plain' }).size
}

// Lifecycle
onMounted(() => {
  loadTree()
})
</script>

<style scoped>
.all-resources-page {
  padding: 24px;
}

.controls-section {
  margin-bottom: 24px;
}

.tree-container {
  min-height: 400px;
  margin-bottom: 24px;
}

.tree-node-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.node-title {
  font-weight: 500;
  flex: 1;
}

.node-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #8c8c8c;
}

.file-size {
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 4px;
}

.content-type {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 6px;
  border-radius: 4px;
}

.statistics-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #f0f0f0;
}

.error-state,
.empty-state {
  text-align: center;
  padding: 40px 20px;
}

:deep(.ant-tree-node-content-wrapper) {
  width: 100%;
}

:deep(.ant-tree-title) {
  width: 100%;
}

/* Create resource modal styles */
.content-type-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.content-type-label {
  font-weight: 500;
}

.content-type-value {
  font-size: 12px;
  color: #8c8c8c;
  background: #f0f0f0;
  padding: 2px 6px;
  border-radius: 4px;
}

.content-preview {
  padding: 12px;
  background: #fafafa;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
}

:deep(.ant-form-item-explain) {
  font-size: 12px;
}

:deep(.ant-textarea) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.4;
}
</style>
