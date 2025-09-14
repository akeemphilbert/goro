import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mockMfParser } from './setup';
import {
  parseHtmlForMicroformats,
  createDetectedMicroformats,
  getMicroformatDisplayName,
  getMicroformatDescription
} from '../utils/microformat-parser';
import {
  SAMPLE_HCARD_HTML,
  SAMPLE_HEVENT_HTML,
  SAMPLE_HPRODUCT_HTML,
  SAMPLE_HRECIPE_HTML,
  SAMPLE_HENTRY_HTML,
  SAMPLE_MIXED_HTML,
  SAMPLE_NO_MICROFORMATS_HTML
} from './sample-html';
import { ParsedMicroformat } from '../utils/microformat-types';

describe('Microformat Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('H-Card Integration', () => {
    it('should parse h-card HTML correctly', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-card'],
            properties: {
              name: ['John Doe'],
              org: ['Acme Corporation'],
              'job-title': ['Software Engineer'],
              email: ['john@example.com'],
              url: ['https://johndoe.com'],
              tel: ['+1-555-123-4567'],
              note: ['Passionate software engineer with 10 years of experience.'],
              photo: ['https://example.com/photo.jpg']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_HCARD_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-card');
      expect(getMicroformatDisplayName(detected[0].data)).toBe('John Doe');
      expect(getMicroformatDescription(detected[0].data)).toBe('Acme Corporation - Software Engineer');
    });
  });

  describe('H-Event Integration', () => {
    it('should parse h-event HTML correctly', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-event'],
            properties: {
              name: ['Team Meeting'],
              summary: ['Weekly team synchronization meeting'],
              start: ['2023-12-01T10:00:00'],
              end: ['2023-12-01T11:00:00'],
              location: ['Conference Room A'],
              description: ['Weekly meeting to discuss project progress, blockers, and upcoming tasks. All team members are expected to attend.'],
              url: ['https://example.com/meetings/team-weekly'],
              category: ['work', 'meeting']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_HEVENT_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-event');
      expect(getMicroformatDisplayName(detected[0].data)).toBe('Team Meeting');
      expect(getMicroformatDescription(detected[0].data)).toContain('Starts: 2023-12-01T10:00:00');
      expect(getMicroformatDescription(detected[0].data)).toContain('Location: Conference Room A');
    });
  });

  describe('H-Product Integration', () => {
    it('should parse h-product HTML correctly', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-product'],
            properties: {
              name: ['Awesome Widget Pro'],
              brand: ['Acme Corporation'],
              category: ['Electronics', 'Gadgets'],
              description: ['The most advanced widget on the market. Features include wireless connectivity, AI-powered automation, and a sleek modern design.'],
              price: ['$299.99'],
              identifier: ['WIDGET-PRO-2023'],
              url: ['https://example.com/products/widget-pro'],
              photo: ['https://example.com/widget-pro.jpg']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_HPRODUCT_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-product');
      expect(getMicroformatDisplayName(detected[0].data)).toBe('Awesome Widget Pro');
      expect(getMicroformatDescription(detected[0].data)).toContain('Price: $299.99');
      expect(getMicroformatDescription(detected[0].data)).toContain('Category: Electronics');
    });
  });

  describe('H-Recipe Integration', () => {
    it('should parse h-recipe HTML correctly', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-recipe'],
            properties: {
              name: ['Chocolate Chip Cookies'],
              summary: ['Classic homemade chocolate chip cookies'],
              ingredient: [
                '2 cups all-purpose flour',
                '1 cup granulated sugar',
                '1/2 cup brown sugar',
                '1/2 cup butter, softened',
                '2 large eggs',
                '1 tsp vanilla extract',
                '1 cup chocolate chips'
              ],
              instructions: [
                'Preheat oven to 350°F (175°C)',
                'Mix dry ingredients in a large bowl',
                'Cream butter and sugars, add eggs and vanilla',
                'Combine wet and dry ingredients, fold in chocolate chips',
                'Drop spoonfuls on baking sheet',
                'Bake for 10-12 minutes until golden brown'
              ],
              yield: ['Makes 24 cookies'],
              duration: ['PT45M'],
              category: ['dessert', 'baking'],
              photo: ['https://example.com/cookies.jpg']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_HRECIPE_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-recipe');
      expect(getMicroformatDisplayName(detected[0].data)).toBe('Chocolate Chip Cookies');
      expect(getMicroformatDescription(detected[0].data)).toContain('Serves: Makes 24 cookies');
      expect(getMicroformatDescription(detected[0].data)).toContain('Duration: PT45M');
    });
  });

  describe('H-Entry Integration', () => {
    it('should parse h-entry HTML correctly', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-entry'],
            properties: {
              name: ['Getting Started with Microformats'],
              summary: ['An introduction to microformats and how they can improve your website\'s semantic markup.'],
              published: ['2023-12-01T09:00:00'],
              updated: ['2023-12-01T10:30:00'],
              url: ['https://blog.example.com/microformats-intro'],
              category: ['web development', 'semantic markup', 'microformats'],
              content: ['Microformats are a simple way to add semantic meaning to your HTML markup...']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_HENTRY_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(1);
      expect(detected[0].type).toBe('h-entry');
      expect(getMicroformatDisplayName(detected[0].data)).toBe('Getting Started with Microformats');
      expect(getMicroformatDescription(detected[0].data)).toContain('Published: 2023-12-01T09:00:00');
    });
  });

  describe('Mixed Content Integration', () => {
    it('should parse multiple microformats from mixed HTML', () => {
      const expectedResult: ParsedMicroformat = {
        items: [
          {
            type: ['h-card'],
            properties: {
              name: ['John Doe'],
              org: ['Acme Corporation']
            }
          },
          {
            type: ['h-entry'],
            properties: {
              name: ['Getting Started with Microformats'],
              published: ['2023-12-01T09:00:00']
            }
          },
          {
            type: ['h-event'],
            properties: {
              name: ['Team Meeting'],
              start: ['2023-12-01T10:00:00']
            }
          },
          {
            type: ['h-product'],
            properties: {
              name: ['Awesome Widget Pro'],
              price: ['$299.99']
            }
          },
          {
            type: ['h-recipe'],
            properties: {
              name: ['Chocolate Chip Cookies'],
              yield: ['Makes 24 cookies']
            }
          }
        ],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_MIXED_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(5);
      
      const types = detected.map(d => d.type);
      expect(types).toContain('h-card');
      expect(types).toContain('h-entry');
      expect(types).toContain('h-event');
      expect(types).toContain('h-product');
      expect(types).toContain('h-recipe');
    });
  });

  describe('No Microformats Integration', () => {
    it('should handle HTML with no microformats', () => {
      const expectedResult: ParsedMicroformat = {
        items: [],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(SAMPLE_NO_MICROFORMATS_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(0);
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle parser errors gracefully', () => {
      mockMfParser.mockImplementation(() => {
        throw new Error('Parser failed');
      });

      const result = parseHtmlForMicroformats(SAMPLE_HCARD_HTML);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(0);
      expect(result.items).toEqual([]);
    });

    it('should handle malformed HTML gracefully', () => {
      const malformedHtml = '<div class="h-card"><span class="p-name">John</span'; // Missing closing tag

      const expectedResult: ParsedMicroformat = {
        items: [],
        rels: {},
        'rel-urls': {}
      };

      mockMfParser.mockReturnValue(expectedResult);

      const result = parseHtmlForMicroformats(malformedHtml);
      const detected = createDetectedMicroformats(result, 'https://example.com');

      expect(detected).toHaveLength(0);
    });
  });
});