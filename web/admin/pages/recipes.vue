<template>
  <div class="recipes-page">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <h1>
            <BookOutlined />
            Recipes
          </h1>
          <p>Manage your recipe collection</p>
        </div>
        <div class="header-actions">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Search recipes..."
            class="search-input"
            @search="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input-search>
          <a-button type="primary" @click="showAddRecipeModal">
            <PlusOutlined />
            Add Recipe
          </a-button>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-section">
      <a-card class="filters-card">
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
              <a-select-option value="">All Levels</a-select-option>
              <a-select-option value="easy">Easy</a-select-option>
              <a-select-option value="medium">Medium</a-select-option>
              <a-select-option value="hard">Hard</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-input-number
              v-model:value="maxPrepTime"
              placeholder="Max Prep Time (min)"
              style="width: 100%"
              :min="0"
              @change="handleFilterChange"
            />
          </a-col>
          <a-col :span="6">
            <a-button @click="clearFilters" class="clear-filters-btn">
              <ClearOutlined />
              Clear Filters
            </a-button>
          </a-col>
        </a-row>
      </a-card>
    </div>

    <!-- Recipe Grid -->
    <div class="recipes-grid">
      <a-row :gutter="[24, 24]">
        <a-col 
          v-for="recipe in filteredRecipes"
          :key="recipe.id"
          :xs="24" 
          :sm="12" 
          :md="8" 
          :lg="6"
        >
          <a-card 
            class="recipe-card"
            hoverable
            @click="viewRecipe(recipe)"
          >
            <template #cover>
              <div class="recipe-image">
                <img 
                  v-if="recipe.image" 
                  :src="recipe.image" 
                  :alt="recipe.name"
                  class="recipe-img"
                />
                <div v-else class="recipe-placeholder">
                  <BookOutlined />
                </div>
                <div class="recipe-overlay">
                  <a-button type="primary" shape="circle" @click.stop="viewRecipe(recipe)">
                    <EyeOutlined />
                  </a-button>
                </div>
              </div>
            </template>
            
            <a-card-meta>
              <template #title>
                <div class="recipe-title">{{ recipe.name }}</div>
              </template>
              <template #description>
                <div class="recipe-meta">
                  <div class="recipe-info">
                    <ClockCircleOutlined />
                    <span>{{ recipe.prepTime }} min</span>
                  </div>
                  <div class="recipe-info">
                    <UserOutlined />
                    <span>{{ recipe.servings }} servings</span>
                  </div>
                  <div class="recipe-info">
                    <StarOutlined />
                    <span>{{ recipe.rating }}/5</span>
                  </div>
                </div>
                <div class="recipe-category">
                  <a-tag :color="getCategoryColor(recipe.category)">
                    {{ recipe.category }}
                  </a-tag>
                  <a-tag :color="getDifficultyColor(recipe.difficulty)">
                    {{ recipe.difficulty }}
                  </a-tag>
                </div>
              </template>
            </a-card-meta>
            
            <template #actions>
              <a-button type="text" @click.stop="editRecipe(recipe)">
                <EditOutlined />
                Edit
              </a-button>
              <a-button type="text" @click.stop="shareRecipe(recipe)">
                <ShareAltOutlined />
                Share
              </a-button>
              <a-popconfirm
                title="Are you sure you want to delete this recipe?"
                @confirm="deleteRecipe(recipe.id)"
              >
                <a-button type="text" danger @click.stop>
                  <DeleteOutlined />
                  Delete
                </a-button>
              </a-popconfirm>
            </template>
          </a-card>
        </a-col>
      </a-row>
    </div>

    <!-- Empty State -->
    <div v-if="filteredRecipes.length === 0" class="empty-state">
      <a-empty description="No recipes found">
        <template #image>
          <BookOutlined style="font-size: 64px; color: #d9d9d9;" />
        </template>
        <a-button type="primary" @click="showAddRecipeModal">
          <PlusOutlined />
          Add Your First Recipe
        </a-button>
      </a-empty>
    </div>

    <!-- Add/Edit Recipe Modal -->
    <a-modal
      v-model:open="recipeModalVisible"
      :title="editingRecipe ? 'Edit Recipe' : 'Add New Recipe'"
      width="800px"
      @ok="handleRecipeSubmit"
      @cancel="handleRecipeCancel"
    >
      <a-form
        ref="recipeFormRef"
        :model="recipeForm"
        :rules="recipeFormRules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Recipe Name" name="name">
              <a-input v-model:value="recipeForm.name" placeholder="Enter recipe name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Category" name="category">
              <a-select v-model:value="recipeForm.category" placeholder="Select category">
                <a-select-option value="breakfast">Breakfast</a-select-option>
                <a-select-option value="lunch">Lunch</a-select-option>
                <a-select-option value="dinner">Dinner</a-select-option>
                <a-select-option value="dessert">Dessert</a-select-option>
                <a-select-option value="snack">Snack</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-row :gutter="16">
          <a-col :span="8">
            <a-form-item label="Prep Time (minutes)" name="prepTime">
              <a-input-number v-model:value="recipeForm.prepTime" style="width: 100%" :min="0" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="Servings" name="servings">
              <a-input-number v-model:value="recipeForm.servings" style="width: 100%" :min="1" />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item label="Difficulty" name="difficulty">
              <a-select v-model:value="recipeForm.difficulty" placeholder="Select difficulty">
                <a-select-option value="easy">Easy</a-select-option>
                <a-select-option value="medium">Medium</a-select-option>
                <a-select-option value="hard">Hard</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="Description" name="description">
          <a-textarea 
            v-model:value="recipeForm.description" 
            placeholder="Describe your recipe..."
            :rows="3"
          />
        </a-form-item>
        
        <a-form-item label="Ingredients" name="ingredients">
          <a-textarea 
            v-model:value="recipeForm.ingredients" 
            placeholder="List ingredients (one per line)..."
            :rows="4"
          />
        </a-form-item>
        
        <a-form-item label="Instructions" name="instructions">
          <a-textarea 
            v-model:value="recipeForm.instructions" 
            placeholder="Step-by-step instructions..."
            :rows="6"
          />
        </a-form-item>
        
        <a-form-item label="Image URL" name="image">
          <a-input v-model:value="recipeForm.image" placeholder="Enter image URL" />
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
  ClearOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  UserOutlined,
  StarOutlined,
  EditOutlined,
  ShareAltOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'

// Reactive state
const searchQuery = ref('')
const selectedCategory = ref('')
const selectedDifficulty = ref('')
const maxPrepTime = ref<number | null>(null)
const recipeModalVisible = ref(false)
const editingRecipe = ref(null)

// Mock recipe data
const recipes = ref([
  {
    id: 1,
    name: 'Chocolate Chip Cookies',
    category: 'dessert',
    difficulty: 'easy',
    prepTime: 30,
    servings: 24,
    rating: 4.5,
    description: 'Classic homemade chocolate chip cookies',
    ingredients: '2 cups flour\n1 cup butter\n1 cup sugar\n2 eggs\n1 tsp vanilla\n1 cup chocolate chips',
    instructions: '1. Mix dry ingredients\n2. Cream butter and sugar\n3. Add eggs and vanilla\n4. Combine wet and dry\n5. Add chocolate chips\n6. Bake at 375°F for 10-12 minutes',
    image: 'https://images.unsplash.com/photo-1499636136210-6f4ee915583e?w=400'
  },
  {
    id: 2,
    name: 'Spaghetti Carbonara',
    category: 'dinner',
    difficulty: 'medium',
    prepTime: 20,
    servings: 4,
    rating: 4.8,
    description: 'Creamy Italian pasta dish',
    ingredients: '400g spaghetti\n200g pancetta\n4 eggs\n100g parmesan\nBlack pepper\nSalt',
    instructions: '1. Cook pasta\n2. Fry pancetta\n3. Beat eggs with parmesan\n4. Combine hot pasta with pancetta\n5. Add egg mixture\n6. Season with pepper',
    image: 'https://images.unsplash.com/photo-1621996346565-e3dbc353d2e5?w=400'
  },
  {
    id: 3,
    name: 'Avocado Toast',
    category: 'breakfast',
    difficulty: 'easy',
    prepTime: 5,
    servings: 1,
    rating: 4.2,
    description: 'Simple and healthy breakfast',
    ingredients: '2 slices bread\n1 avocado\nLemon juice\nSalt\nRed pepper flakes',
    instructions: '1. Toast bread\n2. Mash avocado\n3. Add lemon juice and salt\n4. Spread on toast\n5. Add red pepper flakes',
    image: 'https://images.unsplash.com/photo-1541519227354-08fa5f50a44f?w=400'
  }
])

// Form data
const recipeForm = reactive({
  name: '',
  category: '',
  difficulty: '',
  prepTime: null,
  servings: null,
  description: '',
  ingredients: '',
  instructions: '',
  image: ''
})

const recipeFormRules = {
  name: [{ required: true, message: 'Please enter recipe name' }],
  category: [{ required: true, message: 'Please select category' }],
  difficulty: [{ required: true, message: 'Please select difficulty' }],
  prepTime: [{ required: true, message: 'Please enter prep time' }],
  servings: [{ required: true, message: 'Please enter servings' }]
}

// Computed properties
const filteredRecipes = computed(() => {
  return recipes.value.filter(recipe => {
    const matchesSearch = !searchQuery.value || 
      recipe.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      recipe.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesCategory = !selectedCategory.value || recipe.category === selectedCategory.value
    const matchesDifficulty = !selectedDifficulty.value || recipe.difficulty === selectedDifficulty.value
    const matchesPrepTime = !maxPrepTime.value || recipe.prepTime <= maxPrepTime.value
    
    return matchesSearch && matchesCategory && matchesDifficulty && matchesPrepTime
  })
})

// Methods
const handleSearch = (value: string) => {
  searchQuery.value = value
}

const handleFilterChange = () => {
  // Filters are reactive, no additional action needed
}

const clearFilters = () => {
  searchQuery.value = ''
  selectedCategory.value = ''
  selectedDifficulty.value = ''
  maxPrepTime.value = null
}

const showAddRecipeModal = () => {
  editingRecipe.value = null
  resetForm()
  recipeModalVisible.value = true
}

const editRecipe = (recipe: any) => {
  editingRecipe.value = recipe
  Object.assign(recipeForm, recipe)
  recipeModalVisible.value = true
}

const viewRecipe = (recipe: any) => {
  console.log('Viewing recipe:', recipe)
  // Navigate to recipe detail view
}

const shareRecipe = (recipe: any) => {
  console.log('Sharing recipe:', recipe)
  // Implement sharing functionality
}

const deleteRecipe = (id: number) => {
  const index = recipes.value.findIndex(r => r.id === id)
  if (index > -1) {
    recipes.value.splice(index, 1)
  }
}

const handleRecipeSubmit = () => {
  if (editingRecipe.value) {
    // Update existing recipe
    const index = recipes.value.findIndex(r => r.id === editingRecipe.value.id)
    if (index > -1) {
      recipes.value[index] = { ...recipeForm, id: editingRecipe.value.id, rating: editingRecipe.value.rating }
    }
  } else {
    // Add new recipe
    const newRecipe = {
      ...recipeForm,
      id: Date.now(),
      rating: 0
    }
    recipes.value.push(newRecipe)
  }
  
  recipeModalVisible.value = false
  resetForm()
}

const handleRecipeCancel = () => {
  recipeModalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(recipeForm, {
    name: '',
    category: '',
    difficulty: '',
    prepTime: null,
    servings: null,
    description: '',
    ingredients: '',
    instructions: '',
    image: ''
  })
}

const getCategoryColor = (category: string) => {
  const colors = {
    breakfast: 'green',
    lunch: 'blue',
    dinner: 'purple',
    dessert: 'pink',
    snack: 'orange'
  }
  return colors[category] || 'default'
}

const getDifficultyColor = (difficulty: string) => {
  const colors = {
    easy: 'green',
    medium: 'orange',
    hard: 'red'
  }
  return colors[difficulty] || 'default'
}
</script>

<style scoped>
.recipes-page {
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

.filters-section {
  margin-bottom: 24px;
}

.filters-card {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
}

.clear-filters-btn {
  width: 100%;
}

.recipes-grid {
  margin-bottom: 24px;
}

.recipe-card {
  height: 100%;
  transition: all 0.3s ease;
}

.recipe-card:hover {
  transform: translateY(-4px);
}

.recipe-image {
  position: relative;
  height: 200px;
  overflow: hidden;
}

.recipe-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.recipe-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  color: #d9d9d9;
  font-size: 48px;
}

.recipe-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.recipe-card:hover .recipe-overlay {
  opacity: 1;
}

.recipe-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.recipe-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 8px;
}

.recipe-info {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-color-secondary);
}

.recipe-category {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
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
  
  .filters-card .ant-row {
    flex-direction: column;
  }
  
  .filters-card .ant-col {
    width: 100%;
    margin-bottom: 12px;
  }
}
</style>
