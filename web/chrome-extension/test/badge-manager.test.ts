import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BadgeManager, IconManager, ExtensionUIManager } from '../utils/badge-manager';

// Mock Chrome APIs
const mockChrome = {
  action: {
    setBadgeText: vi.fn(),
    setBadgeBackgroundColor: vi.fn(),
    getBadgeText: vi.fn(),
    getBadgeBackgroundColor: vi.fn(),
    setTitle: vi.fn(),
  }
};

// @ts-ignore
global.chrome = mockChrome;

describe('BadgeManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('updateMicroformatBadge', () => {
    it('should set badge text and color for positive count', async () => {
      await BadgeManager.updateMicroformatBadge(1, 5);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '5',
        tabId: 1
      });
      expect(mockChrome.action.setBadgeBackgroundColor).toHaveBeenCalledWith({
        color: '#4CAF50',
        tabId: 1
      });
    });

    it('should clear badge for zero count', async () => {
      await BadgeManager.updateMicroformatBadge(1, 0);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '',
        tabId: 1
      });
      expect(mockChrome.action.setBadgeBackgroundColor).toHaveBeenCalledWith({
        color: '#757575',
        tabId: 1
      });
    });

    it('should show 99+ for counts over 99', async () => {
      await BadgeManager.updateMicroformatBadge(1, 150);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '99+',
        tabId: 1
      });
    });

    it('should handle errors gracefully', async () => {
      mockChrome.action.setBadgeText.mockRejectedValue(new Error('API Error'));
      
      // Should not throw
      await expect(BadgeManager.updateMicroformatBadge(1, 5)).resolves.toBeUndefined();
    });
  });

  describe('clearBadge', () => {
    it('should clear badge text', async () => {
      await BadgeManager.clearBadge(1);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '',
        tabId: 1
      });
    });
  });

  describe('updateAuthenticationBadge', () => {
    it('should show checkmark for authenticated state', async () => {
      await BadgeManager.updateAuthenticationBadge(1, true);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '✓',
        tabId: 1
      });
      expect(mockChrome.action.setBadgeBackgroundColor).toHaveBeenCalledWith({
        color: '#2196F3',
        tabId: 1
      });
    });

    it('should clear badge for unauthenticated state', async () => {
      await BadgeManager.updateAuthenticationBadge(1, false);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '',
        tabId: 1
      });
    });
  });

  describe('updateErrorBadge', () => {
    it('should show error indicator', async () => {
      await BadgeManager.updateErrorBadge(1, true);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '!',
        tabId: 1
      });
      expect(mockChrome.action.setBadgeBackgroundColor).toHaveBeenCalledWith({
        color: '#F44336',
        tabId: 1
      });
    });
  });

  describe('updateLoadingBadge', () => {
    it('should show loading indicator', async () => {
      await BadgeManager.updateLoadingBadge(1, true);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '...',
        tabId: 1
      });
      expect(mockChrome.action.setBadgeBackgroundColor).toHaveBeenCalledWith({
        color: '#FF9800',
        tabId: 1
      });
    });
  });

  describe('getBadgeState', () => {
    it('should return current badge state', async () => {
      mockChrome.action.getBadgeText.mockResolvedValue('5');
      mockChrome.action.getBadgeBackgroundColor.mockResolvedValue([76, 175, 80, 255]);

      const state = await BadgeManager.getBadgeState(1);

      expect(state).toEqual({
        text: '5',
        color: 'rgba(76,175,80,255)',
        count: 5
      });
    });

    it('should handle non-numeric badge text', async () => {
      mockChrome.action.getBadgeText.mockResolvedValue('✓');
      mockChrome.action.getBadgeBackgroundColor.mockResolvedValue('#2196F3');

      const state = await BadgeManager.getBadgeState(1);

      expect(state.count).toBe(0);
      expect(state.text).toBe('✓');
    });
  });
});

describe('IconManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('updateAuthenticationIcon', () => {
    it('should update title for authenticated state', async () => {
      await IconManager.updateAuthenticationIcon(true, 'https://example.com/profile#me');

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Authenticated (https://example.com/profile#me)'
      });
    });

    it('should update title for unauthenticated state', async () => {
      await IconManager.updateAuthenticationIcon(false);

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension'
      });
    });
  });

  describe('updateErrorIcon', () => {
    it('should update title for error state', async () => {
      await IconManager.updateErrorIcon(true, 'Network error');

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Error - Network error'
      });
    });
  });

  describe('updateLoadingIcon', () => {
    it('should update title for loading state', async () => {
      await IconManager.updateLoadingIcon(true);

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Loading...'
      });
    });
  });

  describe('resetIcon', () => {
    it('should reset title to default', async () => {
      await IconManager.resetIcon();

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension'
      });
    });
  });
});

describe('ExtensionUIManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('updateForMicroformats', () => {
    it('should update badge for microformat count', async () => {
      await ExtensionUIManager.updateForMicroformats(1, 3);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '3',
        tabId: 1
      });
    });
  });

  describe('updateForAuthentication', () => {
    it('should update icon for authentication status', async () => {
      await ExtensionUIManager.updateForAuthentication(true, 'https://example.com/profile#me');

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Authenticated (https://example.com/profile#me)'
      });
    });
  });

  describe('updateForError', () => {
    it('should update both badge and icon for error state', async () => {
      await ExtensionUIManager.updateForError(1, 'Test error');

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '!',
        tabId: 1
      });
      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Error - Test error'
      });
    });
  });

  describe('updateForLoading', () => {
    it('should update both badge and icon for loading state', async () => {
      await ExtensionUIManager.updateForLoading(1, true);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '...',
        tabId: 1
      });
      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension - Loading...'
      });
    });
  });

  describe('clearForTab', () => {
    it('should clear badge for tab', async () => {
      await ExtensionUIManager.clearForTab(1);

      expect(mockChrome.action.setBadgeText).toHaveBeenCalledWith({
        text: '',
        tabId: 1
      });
    });
  });

  describe('resetToDefault', () => {
    it('should reset icon to default state', async () => {
      await ExtensionUIManager.resetToDefault();

      expect(mockChrome.action.setTitle).toHaveBeenCalledWith({
        title: 'Microformat Extension'
      });
    });
  });
});