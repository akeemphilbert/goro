import { 
  scanForMicroformatsOptimized, 
  debouncedScan,
  getMicroformatDisplayName
} from '../utils/browser-microformat-parser';
import type { DetectedMicroformat } from '../utils/microformat-types';

export default defineContentScript({
  matches: ['<all_urls>'],
  main() {
    console.log('Microformat content script loaded');

    let lastDetectedMicroformats: DetectedMicroformat[] = [];
    let isScanning = false;

    // Detect microformats on the current page
    function detectMicroformats(): DetectedMicroformat[] {
      if (isScanning) {
        return lastDetectedMicroformats;
      }

      try {
        isScanning = true;
        const microformats = scanForMicroformatsOptimized();
        lastDetectedMicroformats = microformats;
        return microformats;
      } catch (error) {
        console.error('Error detecting microformats:', error);
        return [];
      } finally {
        isScanning = false;
      }
    }

    // Send detected microformats to background script
    function reportMicroformats(microformats?: DetectedMicroformat[]) {
      const detected = microformats || detectMicroformats();
      
      try {
        if (detected.length > 0) {
          chrome.runtime.sendMessage({
            type: 'MICROFORMATS_DETECTED',
            count: detected.length,
            microformats: detected.map(mf => ({
              type: mf.type,
              hash: mf.hash,
              sourceUrl: mf.sourceUrl,
              detectedAt: mf.detectedAt.toISOString(),
              displayName: getMicroformatDisplayName(mf.data),
              // Send minimal data to background to avoid message size limits
              data: {
                type: mf.data.type,
                properties: Object.keys(mf.data.properties).reduce((acc, key) => {
                  // Limit property values to prevent large messages
                  const values = mf.data.properties[key];
                  if (Array.isArray(values)) {
                    acc[key] = values.slice(0, 3).map(v => 
                      typeof v === 'string' ? v.substring(0, 200) : v
                    );
                  }
                  return acc;
                }, {} as any)
              }
            })),
            url: window.location.href
          });
        } else {
          chrome.runtime.sendMessage({
            type: 'NO_MICROFORMATS',
            url: window.location.href
          });
        }
      } catch (error) {
        console.error('Error sending microformats to background:', error);
      }
    }



    // Debounced scanning to avoid performance issues
    function performDebouncedScan() {
      debouncedScan((microformats) => {
        reportMicroformats(microformats);
      }, 500);
    }

    // Initial scan when page is ready
    function initialScan() {
      // Wait a bit for dynamic content to load
      setTimeout(() => {
        performDebouncedScan();
      }, 1000);
    }

    // Set up initial scanning
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', initialScan);
    } else {
      initialScan();
    }

    // Monitor for dynamic content changes (with throttling)
    let mutationTimeout: number | null = null;
    const observer = new MutationObserver((mutations) => {
      // Only scan if there are significant changes
      const hasSignificantChanges = mutations.some(mutation => 
        mutation.type === 'childList' && 
        (mutation.addedNodes.length > 0 || mutation.removedNodes.length > 0)
      );

      if (hasSignificantChanges) {
        if (mutationTimeout) {
          clearTimeout(mutationTimeout);
        }
        
        mutationTimeout = window.setTimeout(() => {
          performDebouncedScan();
          mutationTimeout = null;
        }, 2000); // Wait 2 seconds after changes stop
      }
    });

    // Start observing with throttled options
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: false, // Don't watch attribute changes to reduce noise
      characterData: false
    });

    // Listen for messages from popup and background
    chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
      try {
        switch (message.type) {
          case 'GET_MICROFORMATS':
            // Force a fresh scan when popup requests data
            const microformats = detectMicroformats();
            sendResponse({ 
              success: true,
              microformats: microformats,
              count: microformats.length,
              url: window.location.href
            });
            break;

          case 'RESCAN_PAGE':
            // Force rescan and report results
            lastDetectedMicroformats = []; // Clear cache
            const newMicroformats = detectMicroformats();
            reportMicroformats(newMicroformats);
            sendResponse({ 
              success: true,
              count: newMicroformats.length 
            });
            break;

          default:
            sendResponse({ success: false, error: 'Unknown message type' });
        }
      } catch (error) {
        console.error('Error handling message:', error);
        sendResponse({ success: false, error: error.message });
      }
      
      return true; // Keep message channel open for async response
    });

    // Clean up on page unload
    window.addEventListener('beforeunload', () => {
      observer.disconnect();
      if (mutationTimeout) {
        clearTimeout(mutationTimeout);
      }
    });

    // Report initial state
    chrome.runtime.sendMessage({
      type: 'CONTENT_SCRIPT_READY',
      url: window.location.href
    });
  }
});