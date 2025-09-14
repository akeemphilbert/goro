import { defineConfig } from 'wxt';

// See https://wxt.dev/api/config.html
export default defineConfig({
  manifest: {
    name: 'Microformat to Solid Pod Extension',
    description: 'Detect microformats on web pages and save them to your Solid pod',
    version: '1.0.0',
    icons: {
      16: 'icon-16.png',
      48: 'icon-48.png',
      128: 'icon-128.png'
    },
    action: {
      default_popup: 'popup/index.html',
      default_title: 'Microformat Extension',
      default_icon: {
        16: 'icon-16.png',
        48: 'icon-48.png',
        128: 'icon-128.png'
      }
    },
    permissions: [
      'activeTab',
      'storage',
      'identity',
      'tabs'
    ],
    host_permissions: [
      'https://*/*'
    ],
    web_accessible_resources: [
      {
        resources: ['*.png', '*.jpg', '*.svg'],
        matches: ['<all_urls>']
      }
    ]
  }
});