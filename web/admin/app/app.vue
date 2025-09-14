<template>
  <div id="app" class="solid-pod-app">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<style>
/* Global styles for the Solid Pod application */
.solid-pod-app {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

/* Custom Ant Design overrides */
:root {
  --primary-color: #667eea;
  --primary-color-hover: #5a6fd8;
  --secondary-color: #764ba2;
  --success-color: #52c41a;
  --warning-color: #faad14;
  --error-color: #ff4d4f;
  --text-color: #262626;
  --text-color-secondary: #8c8c8c;
  --border-color: #d9d9d9;
  --background-color: #fafafa;
  --card-background: #ffffff;
  --shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  --shadow-hover: 0 4px 16px rgba(0, 0, 0, 0.15);
  --border-radius: 8px;
  --border-radius-lg: 12px;
}

/* Override Ant Design default styles */
.ant-layout {
  background: transparent !important;
}

.ant-layout-header {
  background: rgba(255, 255, 255, 0.95) !important;
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: var(--shadow);
}

.ant-layout-content {
  background: transparent !important;
  padding: 24px !important;
}

.ant-card {
  border-radius: var(--border-radius-lg) !important;
  box-shadow: var(--shadow) !important;
  border: 1px solid rgba(255, 255, 255, 0.2) !important;
  background: var(--card-background) !important;
  backdrop-filter: blur(10px);
}

.ant-card:hover {
  box-shadow: var(--shadow-hover) !important;
  transform: translateY(-2px);
  transition: all 0.3s ease;
}

.ant-btn-primary {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color)) !important;
  border: none !important;
  border-radius: var(--border-radius) !important;
  box-shadow: var(--shadow) !important;
  font-weight: 500 !important;
}

.ant-btn-primary:hover {
  background: linear-gradient(135deg, var(--primary-color-hover), var(--secondary-color)) !important;
  transform: translateY(-1px);
  box-shadow: var(--shadow-hover) !important;
}

.ant-input, .ant-select-selector {
  border-radius: var(--border-radius) !important;
  border: 1px solid var(--border-color) !important;
}

.ant-input:focus, .ant-select-focused .ant-select-selector {
  border-color: var(--primary-color) !important;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2) !important;
}

.ant-menu {
  background: transparent !important;
  border: none !important;
}

.ant-menu-item {
  border-radius: var(--border-radius) !important;
  margin: 4px 0 !important;
}

.ant-menu-item-selected {
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color)) !important;
  color: white !important;
}

.ant-avatar {
  box-shadow: var(--shadow) !important;
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.7);
}

/* Responsive design */
@media (max-width: 768px) {
  .ant-layout-content {
    padding: 16px !important;
  }
}
</style>
