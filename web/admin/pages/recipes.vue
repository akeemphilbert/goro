<template>
  <div>
    <!-- Page Header -->
    <a-page-header
      title="Recipes"
      sub-title="Manage your recipe collection"
    >
      <template #extra>
        <a-input-search
          v-model:value="searchQuery"
          placeholder="Search recipes..."
          style="width: 300px; margin-right: 16px"
          @search="handleSearch"
        />
        <a-button type="primary" @click="showAddRecipeModal">
          <PlusOutlined />
          Add Recipe
        </a-button>
      </template>
    </a-page-header>

    <!-- Filters -->
    <a-card style="margin-bottom: 24px">
      <a-row :gutter="16" align="middle">
        <a-col :span="6">
          <a-select
            v-model:value="selectedCategory"
            placeholder="Category"
            style="width: 100%"
            @change="handleFilterChange"
          >
            <a-select-option value="">All Categories</a-select-option>
            <a-select-option value="breakfast">Breakfast</a-select-option>
            <a-select-option value="lunch">Lunch</a-select-option>
            <a-select-option value="dinner">Dinner</a-select-option>
            <a-select-option value="dessert">Dessert</a-select-option>
            <a-select-option value="snack">Snack</a-select-option>
          </a-select>
        </a-col>
        <a-col :span="6">
          <a-select
            v-model:value="selectedDifficulty"
            placeholder="Difficulty"
            style="width: 100%"
            @change="handleFilterChange"
          >
            <a-select-option value="">All Difficulties</a-select-option>
            <a-select-option value="easy">Easy</a-select-option>
            <a-select-option value="medium">Medium</a-select-option>
            <a-select-option value="hard">Hard</a-select-option>
          </a-select>
        </a-col>
        <a-col :span="6">
          <a-select
            v-model:value="selectedCookTime"
            placeholder="Cook Time"
            style="width: 100%"
            @change="handleFilterChange"
          >
            <a-select-option value="">Any Time</a-select-option>
            <a-select-option value="quick">Under 30 min</a-select-option>
            <a-select-option value="medium">30-60 min</a-select-option>
            <a-select-option value="long">Over 1 hour</a-select-option>
          </a-select>
        </a-col>
        <a-col :span="6">
          <a-button @click="clearFilters">Clear Filters</a-button>
        </a-col>
      </a-row>
    </a-card>

    <!-- Recipe Grid -->
    <a-row :gutter="[16, 16]">
      <a-col
        v-for="recipe in filteredRecipes"
        :key="recipe.id"
        :xs="24"
        :sm="12"
        :md="8"
        :lg="6"
      >
        <a-card
          hoverable
          :cover="recipe.image ? `<img alt='recipe' src='${recipe.image}' />` : undefined"
        >
          <template #actions>
            <EditOutlined @click="editRecipe(recipe)" />
            <ShareAltOutlined @click="shareRecipe(recipe)" />
            <DeleteOutlined @click="deleteRecipe(recipe)" />
          </template>
          <a-card-meta
            :title="recipe.title"
            :description="recipe.description"
          />
          <div style="margin-top: 16px">
            <a-tag :color="getCategoryColor(recipe.category)">{{ recipe.category }}</a-tag>
            <a-rate :value="recipe.rating" disabled style="font-size: 14px" />
            <div style="margin-top: 8px; color: #666;">
              <ClockCircleOutlined /> {{ recipe.cookTime }} min
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <!-- Add Recipe Modal -->
    <a-modal
      v-model:open="addRecipeModalVisible"
      title="Add New Recipe"
      width="600px"
      @ok="handleAddRecipe"
      @cancel="addRecipeModalVisible = false"
    >
      <a-form ref="addRecipeForm" :model="newRecipe" layout="vertical">
        <a-form-item label="Recipe Title" name="title" :rules="[{ required: true }]">
          <a-input v-model:value="newRecipe.title" />
        </a-form-item>
        <a-form-item label="Description" name="description">
          <a-textarea v-model:value="newRecipe.description" :rows="3" />
        </a-form-item>
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Category" name="category">
              <a-select v-model:value="newRecipe.category">
                <a-select-option value="breakfast">Breakfast</a-select-option>
                <a-select-option value="lunch">Lunch</a-select-option>
                <a-select-option value="dinner">Dinner</a-select-option>
                <a-select-option value="dessert">Dessert</a-select-option>
                <a-select-option value="snack">Snack</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Cook Time (minutes)" name="cookTime">
              <a-input-number v-model:value="newRecipe.cookTime" style="width: 100%" />
            </a-form-item>
          </a-col>
        </a-row>
        <a-form-item label="Ingredients" name="ingredients">
          <a-textarea v-model:value="newRecipe.ingredients" :rows="4" placeholder="List ingredients..." />
        </a-form-item>
        <a-form-item label="Instructions" name="instructions">
          <a-textarea v-model:value="newRecipe.instructions" :rows="4" placeholder="Cooking instructions..." />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import {
  BookOutlined,
  SearchOutlined,
  PlusOutlined,
  EditOutlined,
  ShareAltOutlined,
  DeleteOutlined,
  ClockCircleOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const selectedCategory = ref('')
const selectedDifficulty = ref('')
const selectedCookTime = ref('')
const addRecipeModalVisible = ref(false)

const newRecipe = reactive({
  title: '',
  description: '',
  category: '',
  cookTime: 30,
  ingredients: '',
  instructions: ''
})

// Mock data
const recipes = ref([
  {
    id: 1,
    title: 'Chocolate Chip Cookies',
    description: 'Classic homemade cookies',
    category: 'dessert',
    cookTime: 25,
    difficulty: 'easy',
    rating: 5,
    image: 'https://via.placeholder.com/300x200'
  },
  {
    id: 2,
    title: 'Chicken Teriyaki',
    description: 'Japanese-style grilled chicken',
    category: 'dinner',
    cookTime: 45,
    difficulty: 'medium',
    rating: 4,
    image: 'https://via.placeholder.com/300x200'
  },
  {
    id: 3,
    title: 'Caesar Salad',
    description: 'Fresh and crispy salad',
    category: 'lunch',
    cookTime: 15,
    difficulty: 'easy',
    rating: 4,
    image: 'https://via.placeholder.com/300x200'
  },
  {
    id: 4,
    title: 'Pancakes',
    description: 'Fluffy breakfast pancakes',
    category: 'breakfast',
    cookTime: 20,
    difficulty: 'easy',
    rating: 5,
    image: 'https://via.placeholder.com/300x200'
  }
])

// Computed
const filteredRecipes = computed(() => {
  return recipes.value.filter(recipe => {
    const matchesSearch = recipe.title.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesCategory = !selectedCategory.value || recipe.category === selectedCategory.value
    const matchesDifficulty = !selectedDifficulty.value || recipe.difficulty === selectedDifficulty.value
    const matchesCookTime = !selectedCookTime.value || (
      (selectedCookTime.value === 'quick' && recipe.cookTime < 30) ||
      (selectedCookTime.value === 'medium' && recipe.cookTime >= 30 && recipe.cookTime <= 60) ||
      (selectedCookTime.value === 'long' && recipe.cookTime > 60)
    )
    
    return matchesSearch && matchesCategory && matchesDifficulty && matchesCookTime
  })
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
}

const handleFilterChange = () => {
  // Filters are reactive, so this triggers recomputation
}

const clearFilters = () => {
  selectedCategory.value = ''
  selectedDifficulty.value = ''
  selectedCookTime.value = ''
  searchQuery.value = ''
}

const showAddRecipeModal = () => {
  addRecipeModalVisible.value = true
}

const handleAddRecipe = () => {
  // Add new recipe logic
  console.log('Adding recipe:', newRecipe)
  addRecipeModalVisible.value = false
  
  // Reset form
  Object.assign(newRecipe, {
    title: '',
    description: '',
    category: '',
    cookTime: 30,
    ingredients: '',
    instructions: ''
  })
}

const editRecipe = (recipe: any) => {
  console.log('Editing recipe:', recipe)
}

const shareRecipe = (recipe: any) => {
  console.log('Sharing recipe:', recipe)
}

const deleteRecipe = (recipe: any) => {
  console.log('Deleting recipe:', recipe)
}

const getCategoryColor = (category: string): string => {
  const colors: Record<string, string> = {
    breakfast: 'orange',
    lunch: 'blue',
    dinner: 'green',
    dessert: 'purple',
    snack: 'gold'
  }
  return colors[category] || 'default'
}
</script>