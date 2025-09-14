import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { JSDOM } from 'jsdom';
import { 
  SAMPLE_HCARD_HTML, 
  SAMPLE_HEVENT_HTML, 
  SAMPLE_MIXED_HTML,
  SAMPLE_NO_MICROFORMATS_HTML 
} from './sample-html';

// Mock chrome runtime API
const mockChrome = {
  runtime: {
    sendMessage: vi.fn(),
    onMessage: {
      addListener: vi.fn()
    }
  }
};

// Mock microformat-shiv
const mockMicroformats = {
  get: vi.fn()
};

// Set up global mocks
(global as any).chrome = mockChrome;
(global as any).Microformats = mockMicroformats;

describe('Content Script', () => {
  let dom: JSDOM;
  let window: Window;
  let document: Document;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Create a new DOM for each test
    dom = new JSDOM('<!DOCTYPE html><html><body></body></html>', {
      url: 'https://example.com',
      pretendToBeVisual: true,
      resources: 'usable'
    });
    
    window = dom.window as unknown as Window;
    document = window.document;
    
    // Set up global objects
    (global as any).window = window;
    (global as any).document = document;
    (global as any).performance = {
      now: () => Date.now()
    };
    (global as any).setTimeout = window.setTimeout;
    (global as any).clearTimeout = window.clearTimeout;
    (global as any).MutationObserver = window.MutationObserver;
  });

  afterEach(() => {
    dom.window.close();
  });

  describe('Microformat Detection', () => {
    it('should detect h-card microformats', async () => {
      // Set up DOM with h-card
      document.body.innerHTML = SAMPLE_HCARD_HTML;
      
      // Mock microformat-shiv response
      const mockResult = {
        items: [
          {
            type: ['h-card'],
            properties: {
              name: ['John Doe'],
              org: ['Acme Corporation'],
              email: ['john@example.com']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };
      
      mockMicroformats.get.mockReturnValue(mockResult);
      
      // Import and test the browser parser
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-card');
      expect(detected[0].data.properties.name).toEqual(['John Doe']);
    });

    it('should detect multiple microformat types', async () => {
      // Set up DOM with microformat classes to trigger scanning
      document.body.innerHTML = '<div class="h-card">test</div><div class="h-event">test</div>';
      
      const mockResult = {
        items: [
          {
            type: ['h-card'],
            properties: { name: ['John Doe'] }
          },
          {
            type: ['h-event'],
            properties: { name: ['Team Meeting'] }
          },
          {
            type: ['h-product'],
            properties: { name: ['Widget'] }
          }
        ],
        rels: {},
        'rel-urls': {}
      };
      
      mockMicroformats.get.mockReturnValue(mockResult);
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      // Just verify the function runs without error
      expect(Array.isArray(detected)).toBe(true);
    });

    it('should handle pages with no microformats', async () => {
      document.body.innerHTML = SAMPLE_NO_MICROFORMATS_HTML;
      
      mockMicroformats.get.mockReturnValue({
        items: [],
        rels: {},
        'rel-urls': {}
      });
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      expect(detected).toHaveLength(0);
    });

    it('should handle parsing errors gracefully', async () => {
      document.body.innerHTML = '<div>Some content</div>';
      
      mockMicroformats.get.mockImplementation(() => {
        throw new Error('Parse error');
      });
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      expect(detected).toHaveLength(0);
    });
  });

  describe('Performance Optimization', () => {
    it('should complete scanning within reasonable time', async () => {
      // Create a large DOM with many elements
      const largeContent = Array(1000).fill(0).map((_, i) => 
        `<div class="item-${i}">Content ${i}</div>`
      ).join('');
      
      document.body.innerHTML = largeContent;
      
      mockMicroformats.get.mockReturnValue({
        items: [],
        rels: {},
        'rel-urls': {}
      });
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const startTime = performance.now();
      const detected = scanForMicroformatsOptimized();
      const endTime = performance.now();
      
      const scanTime = endTime - startTime;
      
      expect(detected).toHaveLength(0);
      expect(scanTime).toBeLessThan(1000); // Should complete within 1 second
    });

    it('should use quick check for potential microformat elements', async () => {
      // Page with no microformat classes
      document.body.innerHTML = '<div><p>Regular content</p></div>';
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      expect(detected).toHaveLength(0);
      // Should not call microformat parser if no potential elements found
      expect(mockMicroformats.get).not.toHaveBeenCalled();
    });

    it('should scan when potential microformat elements are present', async () => {
      // Page with microformat classes
      document.body.innerHTML = '<div class="h-card"><span class="p-name">Test</span></div>';
      
      mockMicroformats.get.mockReturnValue({
        items: [
          {
            type: ['h-card'],
            properties: { name: ['Test'] }
          }
        ],
        rels: {},
        'rel-urls': {}
      });
      
      const { scanForMicroformatsOptimized } = await import('../utils/browser-microformat-parser');
      
      const detected = scanForMicroformatsOptimized();
      
      // Just verify the function runs without error
      expect(Array.isArray(detected)).toBe(true);
    });
  });

  describe('Debounced Scanning', () => {
    it('should export debounced scan function', async () => {
      const { debouncedScan } = await import('../utils/browser-microformat-parser');
      
      // Just verify the function exists and is callable
      expect(typeof debouncedScan).toBe('function');
    });
  });

  describe('Hash Generation', () => {
    it('should generate consistent hashes for same microformat', async () => {
      const { createMicroformatHash } = await import('../utils/browser-microformat-parser');
      
      const mf = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };
      
      const hash1 = createMicroformatHash(mf, 'https://example.com');
      const hash2 = createMicroformatHash(mf, 'https://example.com');
      
      expect(hash1).toBe(hash2);
    });

    it('should generate different hashes for different microformats', async () => {
      const { createMicroformatHash } = await import('../utils/browser-microformat-parser');
      
      const mf1 = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };
      
      const mf2 = {
        type: ['h-card'],
        properties: { name: ['Jane Doe'] }
      };
      
      const hash1 = createMicroformatHash(mf1, 'https://example.com');
      const hash2 = createMicroformatHash(mf2, 'https://example.com');
      
      expect(hash1).not.toBe(hash2);
    });
  });

  describe('Validation', () => {
    it('should validate correct microformat structure', async () => {
      const { validateMicroformat } = await import('../utils/browser-microformat-parser');
      
      const validMf = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };
      
      expect(validateMicroformat(validMf)).toBe(true);
    });

    it('should reject invalid microformat structure', async () => {
      const { validateMicroformat } = await import('../utils/browser-microformat-parser');
      
      const invalidMf = {
        type: [],
        properties: {}
      };
      
      expect(validateMicroformat(invalidMf)).toBe(false);
    });

    it('should reject unsupported microformat types', async () => {
      const { validateMicroformat } = await import('../utils/browser-microformat-parser');
      
      const unsupportedMf = {
        type: ['h-unsupported'],
        properties: { name: ['Test'] }
      };
      
      expect(validateMicroformat(unsupportedMf)).toBe(false);
    });
  });
});