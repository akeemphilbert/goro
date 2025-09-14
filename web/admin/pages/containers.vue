<template>
  <div>
    <!-- Page Header -->
    <a-page-header
      title="Containers"
      sub-title="Organize your data with containers"
    >
      <template #extra>
        <a-input-search
          v-model:value="searchQuery"
          placeholder="Search containers..."
          style="width: 300px; margin-right: 16px"
          @search="handleSearch"
        />
        <a-button type="primary" @click="showCreateContainerModal">
          <PlusOutlined />
          Create Container
        </a-button>
      </template>
    </a-page-header>

    <!-- Container Tree -->
    <a-card>
      <a-tree
        :tree-data="containerTree"
        :expanded-keys="expandedKeys"
        :selected-keys="selectedKeys"
        show-icon
        @expand="onExpand"
        @select="onSelect"
      >
        <template #icon="{ key, dataRef }">
          <FolderOutlined v-if="dataRef.type === 'container'" />
          <FileOutlined v-else />
        </template>
        <template #title="{ key, title, dataRef }">
          <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
            <span>{{ title }}</span>
            <a-dropdown :trigger="['click']" v-if="dataRef.type === 'container'">
              <a-button type="text" size="small" @click.stop>
                <MoreOutlined />
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item @click="editContainer(dataRef)">
                    <EditOutlined />
                    Edit
                  </a-menu-item>
                  <a-menu-item @click="shareContainer(dataRef)">
                    <ShareAltOutlined />
                    Share
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item @click="deleteContainer(dataRef)" danger>
                    <DeleteOutlined />
                    Delete
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </template>
      </a-tree>
    </a-card>

    <!-- Create Container Modal -->
    <a-modal
      v-model:open="createContainerModalVisible"
      title="Create New Container"
      width="500px"
      @ok="handleCreateContainer"
      @cancel="createContainerModalVisible = false"
    >
      <a-form ref="createContainerForm" :model="newContainer" layout="vertical">
        <a-form-item label="Container Name" name="name" :rules="[{ required: true }]">
          <a-input v-model:value="newContainer.name" />
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="newContainer.description" :rows="3" />
        </a-form-item>
        <a-form-item label="Parent Container" name="parent">
          <a-tree-select
            v-model:value="newContainer.parent"
            :tree-data="containerSelectOptions"
            placeholder="Select parent container (optional)"
            allow-clear
          />
        </a-form-item>
        <a-form-item label="Access Level" name="accessLevel">
          <a-select v-model:value="newContainer.accessLevel">
            <a-select-option value="private">Private</a-select-option>
            <a-select-option value="shared">Shared</a-select-option>
            <a-select-option value="public">Public</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import {
  FolderOutlined,
  FileOutlined,
  SearchOutlined,
  PlusOutlined,
  EditOutlined,
  ShareAltOutlined,
  DeleteOutlined,
  MoreOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const createContainerModalVisible = ref(false)
const expandedKeys = ref(['0-0', '0-0-0'])
const selectedKeys = ref(['0-0-0'])

const newContainer = reactive({
  name: '',
  description: '',
  parent: '',
  accessLevel: 'private'
})

// Mock container tree data
const containerTree = ref([
  {
    title: 'Personal',
    key: '0-0',
    type: 'container',
    children: [
      {
        title: 'Documents',
        key: '0-0-0',
        type: 'container',
        children: [
          {
            title: 'Resume.pdf',
            key: '0-0-0-0',
            type: 'file',
            isLeaf: true
          },
          {
            title: 'Tax_2023.pdf',
            key: '0-0-0-1',
            type: 'file',
            isLeaf: true
          }
        ]
      },
      {
        title: 'Photos',
        key: '0-0-1',
        type: 'container',
        children: [
          {
            title: 'Vacation_2023',
            key: '0-0-1-0',
            type: 'container',
            children: [
              {
                title: 'beach.jpg',
                key: '0-0-1-0-0',
                type: 'file',
                isLeaf: true
              },
              {
                title: 'sunset.jpg',
                key: '0-0-1-0-1',
                type: 'file',
                isLeaf: true
              }
            ]
          }
        ]
      }
    ]
  },
  {
    title: 'Work',
    key: '0-1',
    type: 'container',
    children: [
      {
        title: 'Projects',
        key: '0-1-0',
        type: 'container',
        children: [
          {
            title: 'project_spec.doc',
            key: '0-1-0-0',
            type: 'file',
            isLeaf: true
          }
        ]
      },
      {
        title: 'Meetings',
        key: '0-1-1',
        type: 'container',
        children: [
          {
            title: 'notes_jan.txt',
            key: '0-1-1-0',
            type: 'file',
            isLeaf: true
          }
        ]
      }
    ]
  }
])

// Container options for tree select
const containerSelectOptions = computed(() => {
  const buildOptions = (nodes: any[]): any[] => {
    return nodes.filter(node => node.type === 'container').map(node => ({
      title: node.title,
      value: node.key,
      children: node.children ? buildOptions(node.children) : undefined
    }))
  }
  return buildOptions(containerTree.value)
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
  // In a real app, this would filter the tree
}

const onExpand = (keys: string[]) => {
  expandedKeys.value = keys
}

const onSelect = (keys: string[]) => {
  selectedKeys.value = keys
}

const showCreateContainerModal = () => {
  createContainerModalVisible.value = true
}

const handleCreateContainer = () => {
  console.log('Creating container:', newContainer)
  createContainerModalVisible.value = false
  
  // Reset form
  Object.assign(newContainer, {
    name: '',
    description: '',
    parent: '',
    accessLevel: 'private'
  })
}

const editContainer = (container: any) => {
  console.log('Editing container:', container)
}

const shareContainer = (container: any) => {
  console.log('Sharing container:', container)
}

const deleteContainer = (container: any) => {
  console.log('Deleting container:', container)
}
</script>