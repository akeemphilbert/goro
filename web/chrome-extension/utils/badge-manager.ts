/**
 * Badge and Icon Management Utility
 * Handles extension badge updates and icon state changes
 */

export interface BadgeState {
  count: number;
  color: string;
  text: string;
}

export interface IconState {
  isAuthenticated: boolean;
  hasErrors: boolean;
  isLoading: boolean;
}

export class BadgeManager {
  private static readonly COLORS = {
    SUCCESS: '#4CAF50',
    WARNING: '#FF9800', 
    ERROR: '#F44336',
    INACTIVE: '#757575',
    AUTHENTICATED: '#2196F3'
  } as const;

  private static readonly MAX_BADGE_COUNT = 99;

  /**
   * Update badge for microformat detection
   */
  static async updateMicroformatBadge(tabId: number, count: number): Promise<void> {
    try {
      const displayCount = count > this.MAX_BADGE_COUNT ? `${this.MAX_BADGE_COUNT}+` : count.toString();
      const badgeText = count > 0 ? displayCount : '';
      const badgeColor = count > 0 ? this.COLORS.SUCCESS : this.COLORS.INACTIVE;

      await Promise.all([
        chrome.action.setBadgeText({
          text: badgeText,
          tabId: tabId
        }),
        chrome.action.setBadgeBackgroundColor({
          color: badgeColor,
          tabId: tabId
        })
      ]);

      console.log(`Badge updated for tab ${tabId}: ${count} microformats`);
    } catch (error) {
      console.error('Error updating microformat badge:', error);
    }
  }

  /**
   * Clear badge for a specific tab
   */
  static async clearBadge(tabId: number): Promise<void> {
    try {
      await chrome.action.setBadgeText({
        text: '',
        tabId: tabId
      });
      console.log(`Badge cleared for tab ${tabId}`);
    } catch (error) {
      console.error('Error clearing badge:', error);
    }
  }

  /**
   * Update badge to show authentication status
   */
  static async updateAuthenticationBadge(tabId: number, isAuthenticated: boolean): Promise<void> {
    try {
      if (isAuthenticated) {
        await Promise.all([
          chrome.action.setBadgeText({
            text: '✓',
            tabId: tabId
          }),
          chrome.action.setBadgeBackgroundColor({
            color: this.COLORS.AUTHENTICATED,
            tabId: tabId
          })
        ]);
      } else {
        await this.clearBadge(tabId);
      }
    } catch (error) {
      console.error('Error updating authentication badge:', error);
    }
  }

  /**
   * Update badge to show error state
   */
  static async updateErrorBadge(tabId: number, hasError: boolean): Promise<void> {
    try {
      if (hasError) {
        await Promise.all([
          chrome.action.setBadgeText({
            text: '!',
            tabId: tabId
          }),
          chrome.action.setBadgeBackgroundColor({
            color: this.COLORS.ERROR,
            tabId: tabId
          })
        ]);
      }
    } catch (error) {
      console.error('Error updating error badge:', error);
    }
  }

  /**
   * Update badge to show loading state
   */
  static async updateLoadingBadge(tabId: number, isLoading: boolean): Promise<void> {
    try {
      if (isLoading) {
        await Promise.all([
          chrome.action.setBadgeText({
            text: '...',
            tabId: tabId
          }),
          chrome.action.setBadgeBackgroundColor({
            color: this.COLORS.WARNING,
            tabId: tabId
          })
        ]);
      }
    } catch (error) {
      console.error('Error updating loading badge:', error);
    }
  }

  /**
   * Get current badge state for a tab
   */
  static async getBadgeState(tabId: number): Promise<BadgeState> {
    try {
      const [text, color] = await Promise.all([
        chrome.action.getBadgeText({ tabId }),
        chrome.action.getBadgeBackgroundColor({ tabId })
      ]);

      return {
        text,
        color: Array.isArray(color) ? `rgba(${color.join(',')})` : color,
        count: text && !isNaN(parseInt(text)) ? parseInt(text) : 0
      };
    } catch (error) {
      console.error('Error getting badge state:', error);
      return { text: '', color: '', count: 0 };
    }
  }
}

export class IconManager {
  private static readonly TITLES = {
    DEFAULT: 'Microformat Extension',
    AUTHENTICATED: 'Microformat Extension - Authenticated',
    ERROR: 'Microformat Extension - Error',
    LOADING: 'Microformat Extension - Loading...'
  } as const;

  /**
   * Update extension icon and title based on authentication status
   */
  static async updateAuthenticationIcon(isAuthenticated: boolean, webId?: string): Promise<void> {
    try {
      const title = isAuthenticated 
        ? `${this.TITLES.AUTHENTICATED}${webId ? ` (${webId})` : ''}`
        : this.TITLES.DEFAULT;

      await chrome.action.setTitle({ title });

      // In a real implementation, you might want to change the actual icon
      // For now, we'll just update the title to indicate status
      console.log(`Icon updated for authentication status: ${isAuthenticated}`);
    } catch (error) {
      console.error('Error updating authentication icon:', error);
    }
  }

  /**
   * Update icon to show error state
   */
  static async updateErrorIcon(hasError: boolean, errorMessage?: string): Promise<void> {
    try {
      const title = hasError 
        ? `${this.TITLES.ERROR}${errorMessage ? ` - ${errorMessage}` : ''}`
        : this.TITLES.DEFAULT;

      await chrome.action.setTitle({ title });
      console.log(`Icon updated for error state: ${hasError}`);
    } catch (error) {
      console.error('Error updating error icon:', error);
    }
  }

  /**
   * Update icon to show loading state
   */
  static async updateLoadingIcon(isLoading: boolean): Promise<void> {
    try {
      const title = isLoading ? this.TITLES.LOADING : this.TITLES.DEFAULT;
      await chrome.action.setTitle({ title });
      console.log(`Icon updated for loading state: ${isLoading}`);
    } catch (error) {
      console.error('Error updating loading icon:', error);
    }
  }

  /**
   * Reset icon to default state
   */
  static async resetIcon(): Promise<void> {
    try {
      await chrome.action.setTitle({ title: this.TITLES.DEFAULT });
      console.log('Icon reset to default state');
    } catch (error) {
      console.error('Error resetting icon:', error);
    }
  }

  /**
   * Update icon based on overall extension state
   */
  static async updateIconState(state: IconState): Promise<void> {
    try {
      if (state.hasErrors) {
        await this.updateErrorIcon(true);
      } else if (state.isLoading) {
        await this.updateLoadingIcon(true);
      } else if (state.isAuthenticated) {
        await this.updateAuthenticationIcon(true);
      } else {
        await this.resetIcon();
      }
    } catch (error) {
      console.error('Error updating icon state:', error);
    }
  }
}

/**
 * Combined manager for coordinated badge and icon updates
 */
export class ExtensionUIManager {
  /**
   * Update UI based on microformat detection results
   */
  static async updateForMicroformats(tabId: number, count: number): Promise<void> {
    await BadgeManager.updateMicroformatBadge(tabId, count);
  }

  /**
   * Update UI based on authentication status change
   */
  static async updateForAuthentication(isAuthenticated: boolean, webId?: string): Promise<void> {
    await IconManager.updateAuthenticationIcon(isAuthenticated, webId);
  }

  /**
   * Update UI for error states
   */
  static async updateForError(tabId: number, errorMessage?: string): Promise<void> {
    await Promise.all([
      BadgeManager.updateErrorBadge(tabId, true),
      IconManager.updateErrorIcon(true, errorMessage)
    ]);
  }

  /**
   * Update UI for loading states
   */
  static async updateForLoading(tabId: number, isLoading: boolean): Promise<void> {
    await Promise.all([
      BadgeManager.updateLoadingBadge(tabId, isLoading),
      IconManager.updateLoadingIcon(isLoading)
    ]);
  }

  /**
   * Clear all UI indicators for a tab
   */
  static async clearForTab(tabId: number): Promise<void> {
    await BadgeManager.clearBadge(tabId);
  }

  /**
   * Reset all UI to default state
   */
  static async resetToDefault(): Promise<void> {
    await IconManager.resetIcon();
  }
}