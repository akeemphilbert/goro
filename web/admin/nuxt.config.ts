// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  css: [
    'ant-design-vue/dist/reset.css'
  ],
  modules: [
    '@pinia/nuxt'
  ],
  build: {
    transpile: ['ant-design-vue']
  },
  typescript: {
    typeCheck: false
  }
})
