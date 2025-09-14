import { DetectedMicroformat } from '../utils/microformat-types';
import { BadgeManager, IconManager, ExtensionUIManager } from '../utils/badge-manager';

// Storage keys
const STORAGE_KEYS = {
  DETECTED_MICROFORMATS: 'detected_microformats',
  AUTHENTICATION_SESSION: 'auth_session',
  SAVED_RESOURCES: 'saved_resources'
} as const;

// Message types
interface BackgroundMessage {
  type: string;
  data?: any;
  tabId?: number;
}

interface MicroformatsDetectedMessage extends BackgroundMessage {
  type: 'MICROFORMATS_DETECTED';
  data: {
    microformats: DetectedMicroformat[];
    count: number;
    url: string;
  };
}

interface AuthenticationMessage extends BackgroundMessage {
  type: 'AUTHENTICATION_STATUS_CHANGED';
  data: {
    isAuthenticated: boolean;
    webId?: string;
  };
}

interface StorageRequest extends BackgroundMessage {
  type: 'GET_MICROFORMATS' | 'SAVE_MICROFORMAT' | 'GET_AUTH_STATUS';
}

export default defineBackground(() => {
  console.log('Microformat Chrome Extension background script loaded');

  // Extension lifecycle management
  chrome.runtime.onInstalled.addListener(async (details) => {
    console.log('Extension installed/updated:', details.reason);
    
    if (details.reason === 'install') {
      // Initialize storage on first install
      await initializeStorage();
      console.log('Extension storage initialized');
    } else if (details.reason === 'update') {
      // Handle extension updates
      console.log('Extension updated from version:', details.previousVersion);
    }
  });

  // Handle extension startup
  chrome.runtime.onStartup.addListener(async () => {
    console.log('Extension startup');
    // Clear any temporary data or reset state if needed
    await clearTemporaryData();
  });

  // Message handling between content script and popup
  chrome.runtime.onMessage.addListener((message: BackgroundMessage, sender, sendResponse) => {
    console.log('Background received message:', message.type, message);
    
    // Handle async operations
    handleMessage(message, sender, sendResponse);
    
    return true; // Keep message channel open for async response
  });

  // Tab management for badge updates
  chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
    if (changeInfo.status === 'loading') {
      // Show loading state and clear stored microformats
      await ExtensionUIManager.updateForLoading(tabId, true);
      await clearTabMicroformats(tabId);
    } else if (changeInfo.status === 'complete') {
      // Clear loading state when page is complete
      await ExtensionUIManager.updateForLoading(tabId, false);
    }
  });

  // Clean up when tabs are closed
  chrome.tabs.onRemoved.addListener(async (tabId) => {
    await clearTabMicroformats(tabId);
  });

  // Handle authentication state changes for icon updates
  chrome.storage.onChanged.addListener(async (changes, namespace) => {
    if (namespace === 'local' && changes[STORAGE_KEYS.AUTHENTICATION_SESSION]) {
      await updateIconForAuthStatus();
    }
  });

  // Core message handling function
  async function handleMessage(
    message: BackgroundMessage, 
    sender: chrome.runtime.MessageSender, 
    sendResponse: (response?: any) => void
  ) {
    try {
      switch (message.type) {
        case 'MICROFORMATS_DETECTED':
          await handleMicroformatsDetected(message as MicroformatsDetectedMessage, sender);
          sendResponse({ success: true });
          break;

        case 'NO_MICROFORMATS':
          await handleNoMicroformats(sender);
          sendResponse({ success: true });
          break;

        case 'GET_MICROFORMATS':
          const microformats = await getMicroformatsForTab(sender.tab?.id);
          sendResponse({ microformats });
          break;

        case 'SAVE_MICROFORMAT':
          const saveResult = await saveMicroformatToStorage(message.data);
          sendResponse(saveResult);
          break;

        case 'GET_AUTH_STATUS':
          const authStatus = await getAuthenticationStatus();
          sendResponse(authStatus);
          break;

        case 'AUTHENTICATION_STATUS_CHANGED':
          await handleAuthenticationChange(message as AuthenticationMessage);
          sendResponse({ success: true });
          break;

        case 'CLEAR_STORAGE':
          await clearAllStorage();
          sendResponse({ success: true });
          break;

        default:
          console.warn('Unknown message type:', message.type);
          sendResponse({ error: 'Unknown message type' });
      }
    } catch (error) {
      console.error('Error handling message:', error);
      sendResponse({ error: error.message });
    }
  }

  // Handle microformats detection
  async function handleMicroformatsDetected(
    message: MicroformatsDetectedMessage, 
    sender: chrome.runtime.MessageSender
  ) {
    const tabId = sender.tab?.id;
    if (!tabId) return;

    // Store microformats for this tab
    await storeMicroformatsForTab(tabId, message.data.microformats, message.data.url);

    // Update UI using the new manager
    await ExtensionUIManager.updateForMicroformats(tabId, message.data.count);
  }

  // Handle no microformats found
  async function handleNoMicroformats(sender: chrome.runtime.MessageSender) {
    const tabId = sender.tab?.id;
    if (!tabId) return;

    await ExtensionUIManager.clearForTab(tabId);
    await clearTabMicroformats(tabId);
  }

  // Handle authentication status changes
  async function handleAuthenticationChange(message: AuthenticationMessage) {
    await ExtensionUIManager.updateForAuthentication(
      message.data.isAuthenticated, 
      message.data.webId
    );
    
    // Notify all open popups about auth change
    try {
      await chrome.runtime.sendMessage({
        type: 'AUTH_STATUS_UPDATED',
        data: message.data
      });
    } catch (error) {
      // Popup might not be open, which is fine
      console.log('No popup to notify about auth change');
    }
  }

  // Storage management functions
  async function initializeStorage() {
    const defaultData = {
      [STORAGE_KEYS.DETECTED_MICROFORMATS]: {},
      [STORAGE_KEYS.SAVED_RESOURCES]: [],
      [STORAGE_KEYS.AUTHENTICATION_SESSION]: null
    };

    // Only set defaults if keys don't exist
    const existing = await chrome.storage.local.get(Object.values(STORAGE_KEYS));
    const toSet: Record<string, any> = {};

    for (const [key, value] of Object.entries(defaultData)) {
      if (!(key in existing)) {
        toSet[key] = value;
      }
    }

    if (Object.keys(toSet).length > 0) {
      await chrome.storage.local.set(toSet);
    }
  }

  async function storeMicroformatsForTab(tabId: number, microformats: DetectedMicroformat[], url: string) {
    const storage = await chrome.storage.local.get(STORAGE_KEYS.DETECTED_MICROFORMATS);
    const tabMicroformats = storage[STORAGE_KEYS.DETECTED_MICROFORMATS] || {};
    
    tabMicroformats[tabId] = {
      microformats,
      url,
      detectedAt: new Date().toISOString()
    };

    await chrome.storage.local.set({
      [STORAGE_KEYS.DETECTED_MICROFORMATS]: tabMicroformats
    });
  }

  async function getMicroformatsForTab(tabId?: number): Promise<DetectedMicroformat[]> {
    if (!tabId) return [];

    const storage = await chrome.storage.local.get(STORAGE_KEYS.DETECTED_MICROFORMATS);
    const tabMicroformats = storage[STORAGE_KEYS.DETECTED_MICROFORMATS] || {};
    
    return tabMicroformats[tabId]?.microformats || [];
  }

  async function clearTabMicroformats(tabId: number) {
    const storage = await chrome.storage.local.get(STORAGE_KEYS.DETECTED_MICROFORMATS);
    const tabMicroformats = storage[STORAGE_KEYS.DETECTED_MICROFORMATS] || {};
    
    delete tabMicroformats[tabId];
    
    await chrome.storage.local.set({
      [STORAGE_KEYS.DETECTED_MICROFORMATS]: tabMicroformats
    });
  }

  async function saveMicroformatToStorage(microformat: DetectedMicroformat) {
    try {
      const storage = await chrome.storage.local.get(STORAGE_KEYS.SAVED_RESOURCES);
      const savedResources = storage[STORAGE_KEYS.SAVED_RESOURCES] || [];
      
      // Add timestamp and generate hash for duplicate detection
      const resourceToSave = {
        ...microformat,
        savedAt: new Date().toISOString(),
        hash: generateMicroformatHash(microformat)
      };

      // Check for duplicates
      const isDuplicate = savedResources.some((saved: any) => saved.hash === resourceToSave.hash);
      
      if (isDuplicate) {
        return { success: false, error: 'Microformat already saved', isDuplicate: true };
      }

      savedResources.push(resourceToSave);
      
      await chrome.storage.local.set({
        [STORAGE_KEYS.SAVED_RESOURCES]: savedResources
      });

      return { success: true, resourceId: resourceToSave.hash };
    } catch (error) {
      console.error('Error saving microformat:', error);
      return { success: false, error: error.message };
    }
  }

  async function getAuthenticationStatus() {
    const storage = await chrome.storage.local.get(STORAGE_KEYS.AUTHENTICATION_SESSION);
    const session = storage[STORAGE_KEYS.AUTHENTICATION_SESSION];
    
    return {
      isAuthenticated: !!session?.accessToken,
      webId: session?.webId,
      expiresAt: session?.expiresAt
    };
  }

  async function clearTemporaryData() {
    // Clear tab-specific data but keep authentication and saved resources
    await chrome.storage.local.set({
      [STORAGE_KEYS.DETECTED_MICROFORMATS]: {}
    });
    
    // Reset UI to default state
    await ExtensionUIManager.resetToDefault();
  }

  async function clearAllStorage() {
    await chrome.storage.local.clear();
    await initializeStorage();
  }

  // Icon management based on authentication status
  async function updateIconForAuthStatus() {
    try {
      const authStatus = await getAuthenticationStatus();
      await ExtensionUIManager.updateForAuthentication(
        authStatus.isAuthenticated,
        authStatus.webId
      );
    } catch (error) {
      console.error('Error updating icon for auth status:', error);
      // Show error state in UI
      await ExtensionUIManager.updateForError(0, 'Authentication status update failed');
    }
  }

  // Utility functions
  function generateMicroformatHash(microformat: DetectedMicroformat): string {
    // Simple hash generation based on type, source URL, and key properties
    const hashData = {
      type: microformat.type,
      sourceUrl: microformat.sourceUrl,
      properties: microformat.data.properties
    };
    
    // Simple string hash (in production, consider using a proper hash library)
    return btoa(JSON.stringify(hashData)).replace(/[^a-zA-Z0-9]/g, '').substring(0, 16);
  }

  // Initialize storage on startup
  initializeStorage().catch(console.error);
});