import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mockMfParser } from './setup';
import {
  parseHtmlForMicroformats,
  parseDomForMicroformats,
  filterSupportedMicroformats,
  extractMicroformatsByType,
  getPrimaryType,
  createMicroformatHash,
  createDetectedMicroformats,
  getMicroformatDisplayName,
  getMicroformatDescription,
  validateMicroformat,
  normalizeMicroformat
} from '../utils/microformat-parser';
import { 
  MicroformatObject, 
  ParsedMicroformat,
  HCard,
  HEvent,
  HProduct,
  HRecipe,
  HEntry
} from '../utils/microformat-types';

describe('Microformat Parser', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('parseHtmlForMicroformats', () => {
    it('should parse HTML and return microformats', () => {
      const mockResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-card'],
            properties: {
              name: ['John Doe'],
              email: ['john@example.com']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(mockResult);

      const html = '<div class="h-card"><span class="p-name">John Doe</span></div>';
      const result = parseHtmlForMicroformats(html, 'https://example.com');

      expect(mockMfParser).toHaveBeenCalledWith(html, { baseUrl: 'https://example.com' });
      expect(result).toEqual(mockResult);
    });

    it('should handle parsing errors gracefully', () => {
      mockMfParser.mockImplementation(() => {
        throw new Error('Parse error');
      });

      const html = '<div>invalid</div>';
      const result = parseHtmlForMicroformats(html);

      expect(result).toEqual({
        items: [],
        rels: {},
        'rel-urls': {}
      });
    });
  });

  describe('parseDomForMicroformats', () => {
    it('should parse DOM element for microformats', () => {
      const mockResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-card'],
            properties: {
              name: ['Jane Doe']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(mockResult);

      // Create a mock DOM element
      const element = document.createElement('div');
      element.className = 'h-card';
      element.innerHTML = '<span class="p-name">Jane Doe</span>';

      const result = parseDomForMicroformats(element, 'https://example.com');

      expect(result).toEqual(mockResult);
    });
  });

  describe('filterSupportedMicroformats', () => {
    it('should filter only supported microformat types', () => {
      const parsed: ParsedMicroformat = {
        items: [
          { type: ['h-card'], properties: { name: ['John'] } },
          { type: ['h-event'], properties: { name: ['Meeting'] } },
          { type: ['h-unsupported'], properties: { name: ['Test'] } },
          { type: ['h-product'], properties: { name: ['Widget'] } }
        ],
        rels: {},
        'rel-urls': {}
      };

      const result = filterSupportedMicroformats(parsed);

      expect(result).toHaveLength(3);
      expect(result.map(item => item.type[0])).toEqual(['h-card', 'h-event', 'h-product']);
    });
  });

  describe('extractMicroformatsByType', () => {
    it('should extract microformats of specific type', () => {
      const parsed: ParsedMicroformat = {
        items: [
          { type: ['h-card'], properties: { name: ['John'] } },
          { type: ['h-event'], properties: { name: ['Meeting'] } },
          { type: ['h-card'], properties: { name: ['Jane'] } }
        ],
        rels: {},
        'rel-urls': {}
      };

      const result = extractMicroformatsByType(parsed, 'h-card');

      expect(result).toHaveLength(2);
      expect(result.every(item => item.type.includes('h-card'))).toBe(true);
    });
  });

  describe('getPrimaryType', () => {
    it('should return first supported type', () => {
      const mf: MicroformatObject = {
        type: ['h-unsupported', 'h-card', 'h-event'],
        properties: {}
      };

      const result = getPrimaryType(mf);
      expect(result).toBe('h-card');
    });

    it('should return first type if none are supported', () => {
      const mf: MicroformatObject = {
        type: ['h-unsupported', 'h-unknown'],
        properties: {}
      };

      const result = getPrimaryType(mf);
      expect(result).toBe('h-unsupported');
    });
  });

  describe('createMicroformatHash', () => {
    it('should create consistent hash for same microformat', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };

      const hash1 = createMicroformatHash(mf, 'https://example.com');
      const hash2 = createMicroformatHash(mf, 'https://example.com');

      expect(hash1).toBe(hash2);
      expect(typeof hash1).toBe('string');
      expect(hash1.length).toBeGreaterThan(0);
    });

    it('should create different hashes for different microformats', () => {
      const mf1: MicroformatObject = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };

      const mf2: MicroformatObject = {
        type: ['h-card'],
        properties: { name: ['Jane Doe'] }
      };

      const hash1 = createMicroformatHash(mf1, 'https://example.com');
      const hash2 = createMicroformatHash(mf2, 'https://example.com');

      expect(hash1).not.toBe(hash2);
    });
  });

  describe('createDetectedMicroformats', () => {
    it('should convert parsed microformats to detected microformats', () => {
      const parsed: ParsedMicroformat = {
        items: [
          { type: ['h-card'], properties: { name: ['John Doe'] } },
          { type: ['h-event'], properties: { name: ['Meeting'] } }
        ],
        rels: {},
        'rel-urls': {}
      };

      const result = createDetectedMicroformats(parsed, 'https://example.com');

      expect(result).toHaveLength(2);
      expect(result[0]).toMatchObject({
        type: 'h-card',
        sourceUrl: 'https://example.com',
        data: { type: ['h-card'], properties: { name: ['John Doe'] } }
      });
      expect(result[0].detectedAt).toBeInstanceOf(Date);
      expect(result[0].hash).toBeDefined();
    });
  });

  describe('getMicroformatDisplayName', () => {
    it('should return name property if available', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { name: ['John Doe'], org: ['Acme Corp'] }
      };

      expect(getMicroformatDisplayName(mf)).toBe('John Doe');
    });

    it('should return summary if name not available', () => {
      const mf: MicroformatObject = {
        type: ['h-event'],
        properties: { summary: ['Team Meeting'] }
      };

      expect(getMicroformatDisplayName(mf)).toBe('Team Meeting');
    });

    it('should return type-specific fallback for h-card', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { org: ['Acme Corp'] }
      };

      expect(getMicroformatDisplayName(mf)).toBe('Acme Corp');
    });

    it('should return generic fallback', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: {}
      };

      expect(getMicroformatDisplayName(mf)).toBe('h-card microformat');
    });
  });

  describe('getMicroformatDescription', () => {
    it('should return description if available', () => {
      const mf: MicroformatObject = {
        type: ['h-product'],
        properties: { 
          description: ['A great product that does amazing things and more text to test truncation'] 
        }
      };

      const result = getMicroformatDescription(mf);
      expect(result).toContain('A great product');
      expect(result).toContain('...');
    });

    it('should return type-specific description for h-card', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { 
          org: ['Acme Corp'],
          'job-title': ['Software Engineer']
        }
      };

      expect(getMicroformatDescription(mf)).toBe('Acme Corp - Software Engineer');
    });

    it('should return type-specific description for h-event', () => {
      const mf: MicroformatObject = {
        type: ['h-event'],
        properties: { 
          start: ['2023-12-01'],
          location: ['Conference Room A']
        }
      };

      expect(getMicroformatDescription(mf)).toBe('Starts: 2023-12-01 | Location: Conference Room A');
    });
  });

  describe('validateMicroformat', () => {
    it('should validate correct microformat', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { name: ['John Doe'] }
      };

      expect(validateMicroformat(mf)).toBe(true);
    });

    it('should reject microformat without type', () => {
      const mf = {
        properties: { name: ['John Doe'] }
      } as any;

      expect(validateMicroformat(mf)).toBe(false);
    });

    it('should reject microformat with empty type array', () => {
      const mf: MicroformatObject = {
        type: [],
        properties: { name: ['John Doe'] }
      };

      expect(validateMicroformat(mf)).toBe(false);
    });

    it('should reject microformat without properties', () => {
      const mf = {
        type: ['h-card']
      } as any;

      expect(validateMicroformat(mf)).toBe(false);
    });

    it('should reject unsupported microformat type', () => {
      const mf: MicroformatObject = {
        type: ['h-unsupported'],
        properties: { name: ['Test'] }
      };

      expect(validateMicroformat(mf)).toBe(false);
    });
  });

  describe('normalizeMicroformat', () => {
    it('should ensure all properties are arrays', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { 
          name: 'John Doe' as any,
          email: ['john@example.com']
        }
      };

      const result = normalizeMicroformat(mf);

      expect(Array.isArray(result.properties.name)).toBe(true);
      expect(result.properties.name).toEqual(['John Doe']);
      expect(result.properties.email).toEqual(['john@example.com']);
    });

    it('should remove empty properties', () => {
      const mf: MicroformatObject = {
        type: ['h-card'],
        properties: { 
          name: ['John Doe'],
          empty: []
        }
      };

      const result = normalizeMicroformat(mf);

      expect(result.properties.name).toEqual(['John Doe']);
      expect(result.properties.empty).toBeUndefined();
    });
  });
});