console.log('Popup script loaded');

// DOM elements
const authForm = document.getElementById('auth-form')!;
const authStatus = document.getElementById('auth-status')!;
const webidInput = document.getElementById('webid-input') as HTMLInputElement;
const loginBtn = document.getElementById('login-btn')!;
const logoutBtn = document.getElementById('logout-btn')!;
const webidDisplay = document.getElementById('webid-display')!;
const statusMessages = document.getElementById('status-messages')!;
const microformatsList = document.getElementById('microformats-list')!;

// State management
let isAuthenticated = false;
let currentWebId = '';
let detectedMicroformats: any[] = [];

// Initialize popup
async function init() {
  await loadAuthState();
  await loadMicroformats();
  setupEventListeners();
}

// Load authentication state from storage
async function loadAuthState() {
  try {
    const result = await chrome.storage.local.get(['webId', 'isAuthenticated']);
    if (result.isAuthenticated && result.webId) {
      isAuthenticated = true;
      currentWebId = result.webId;
      showAuthenticatedState();
    } else {
      showUnauthenticatedState();
    }
  } catch (error) {
    console.error('Failed to load auth state:', error);
    showUnauthenticatedState();
  }
}

// Load microformats from current tab
async function loadMicroformats() {
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (!tab.id) {
      showError('Unable to access current tab');
      return;
    }

    // Send message to content script to get microformats
    const response = await chrome.tabs.sendMessage(tab.id, { type: 'GET_MICROFORMATS' });
    
    if (response && response.microformats) {
      detectedMicroformats = response.microformats;
      displayMicroformats(detectedMicroformats);
    } else {
      displayNoMicroformats();
    }
  } catch (error) {
    console.error('Failed to load microformats:', error);
    showError('Failed to scan page for microformats');
  }
}

// Setup event listeners
function setupEventListeners() {
  loginBtn.addEventListener('click', handleLogin);
  logoutBtn.addEventListener('click', handleLogout);
  webidInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      handleLogin();
    }
  });
}

// Handle login
async function handleLogin() {
  const webId = webidInput.value.trim();
  
  if (!webId) {
    showError('Please enter your WebID');
    return;
  }

  if (!isValidWebId(webId)) {
    showError('Please enter a valid WebID URL');
    return;
  }

  try {
    showInfo('Authenticating...');
    
    // TODO: Implement actual Solid authentication
    // For now, just simulate successful authentication
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Store authentication state
    await chrome.storage.local.set({
      webId: webId,
      isAuthenticated: true
    });

    isAuthenticated = true;
    currentWebId = webId;
    showAuthenticatedState();
    showSuccess('Successfully authenticated!');
    
  } catch (error) {
    console.error('Authentication failed:', error);
    showError('Authentication failed. Please check your WebID and try again.');
  }
}

// Handle logout
async function handleLogout() {
  try {
    await chrome.storage.local.remove(['webId', 'isAuthenticated', 'accessToken']);
    isAuthenticated = false;
    currentWebId = '';
    showUnauthenticatedState();
    showInfo('Logged out successfully');
  } catch (error) {
    console.error('Logout failed:', error);
    showError('Failed to logout');
  }
}

// Validate WebID format
function isValidWebId(webId: string): boolean {
  try {
    const url = new URL(webId);
    return url.protocol === 'https:' && url.hash.length > 0;
  } catch {
    return false;
  }
}

// Show authenticated state
function showAuthenticatedState() {
  authForm.style.display = 'none';
  authStatus.style.display = 'block';
  webidDisplay.textContent = `Logged in as: ${currentWebId}`;
}

// Show unauthenticated state
function showUnauthenticatedState() {
  authForm.style.display = 'block';
  authStatus.style.display = 'none';
  webidInput.value = '';
}

// Display microformats
function displayMicroformats(microformats: any[]) {
  if (microformats.length === 0) {
    displayNoMicroformats();
    return;
  }

  microformatsList.innerHTML = microformats.map((mf, index) => `
    <div class="microformat-item">
      <div class="microformat-type">${mf.type}</div>
      <div class="microformat-preview">${mf.element}</div>
      <div class="microformat-actions">
        <button class="btn btn-primary save-btn" data-index="${index}" ${!isAuthenticated ? 'disabled' : ''}>
          Save to Pod
        </button>
        <button class="btn btn-secondary preview-btn" data-index="${index}">
          Preview
        </button>
      </div>
    </div>
  `).join('');

  // Add event listeners to dynamically created buttons
  document.querySelectorAll('.save-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const index = parseInt((e.target as HTMLElement).getAttribute('data-index')!);
      saveMicroformat(index);
    });
  });

  document.querySelectorAll('.preview-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const index = parseInt((e.target as HTMLElement).getAttribute('data-index')!);
      previewMicroformat(index);
    });
  });
}

// Display no microformats message
function displayNoMicroformats() {
  microformatsList.innerHTML = `
    <div class="status-message status-info">
      No microformats found on this page.
    </div>
  `;
}

// Show status messages
function showMessage(message: string, type: 'info' | 'success' | 'error') {
  const messageDiv = document.createElement('div');
  messageDiv.className = `status-message status-${type}`;
  messageDiv.textContent = message;
  
  statusMessages.innerHTML = '';
  statusMessages.appendChild(messageDiv);
  
  // Auto-hide after 3 seconds
  setTimeout(() => {
    if (statusMessages.contains(messageDiv)) {
      statusMessages.removeChild(messageDiv);
    }
  }, 3000);
}

function showInfo(message: string) { showMessage(message, 'info'); }
function showSuccess(message: string) { showMessage(message, 'success'); }
function showError(message: string) { showMessage(message, 'error'); }

// Save microformat to pod
async function saveMicroformat(index: number) {
  if (!isAuthenticated) {
    showError('Please login first');
    return;
  }

  const microformat = detectedMicroformats[index];
  if (!microformat) {
    showError('Microformat not found');
    return;
  }

  try {
    showInfo('Saving to pod...');
    
    // TODO: Implement actual pod saving
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    showSuccess('Microformat saved to pod!');
  } catch (error) {
    console.error('Failed to save microformat:', error);
    showError('Failed to save microformat to pod');
  }
}

// Preview microformat
function previewMicroformat(index: number) {
  const microformat = detectedMicroformats[index];
  if (!microformat) {
    showError('Microformat not found');
    return;
  }

  // TODO: Implement preview modal
  alert(`Microformat Preview:\n\nType: ${microformat.type}\nURL: ${microformat.url}\nDetected: ${microformat.detectedAt}`);
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}