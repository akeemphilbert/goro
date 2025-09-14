import { describe, it, expect } from 'vitest';
import {
  isHCard,
  isHEvent,
  isHProduct,
  isHRecipe,
  isHEntry,
  isSupportedMicroformat,
  SUPPORTED_MICROFORMAT_TYPES,
  MicroformatObject,
  HCard,
  HEvent,
  HProduct,
  HRecipe,
  HEntry
} from '../utils/microformat-types';

describe('Microformat Types', () => {
  describe('Type Guards', () => {
    describe('isHCard', () => {
      it('should return true for h-card microformat', () => {
        const hcard: MicroformatObject = {
          type: ['h-card'],
          properties: {
            name: ['John Doe'],
            email: ['john@example.com']
          }
        };

        expect(isHCard(hcard)).toBe(true);
      });

      it('should return false for non h-card microformat', () => {
        const hevent: MicroformatObject = {
          type: ['h-event'],
          properties: {
            name: ['Meeting']
          }
        };

        expect(isHCard(hevent)).toBe(false);
      });

      it('should return true for microformat with multiple types including h-card', () => {
        const mixed: MicroformatObject = {
          type: ['h-card', 'vcard'],
          properties: {
            name: ['John Doe']
          }
        };

        expect(isHCard(mixed)).toBe(true);
      });
    });

    describe('isHEvent', () => {
      it('should return true for h-event microformat', () => {
        const hevent: MicroformatObject = {
          type: ['h-event'],
          properties: {
            name: ['Team Meeting'],
            start: ['2023-12-01T10:00:00']
          }
        };

        expect(isHEvent(hevent)).toBe(true);
      });

      it('should return false for non h-event microformat', () => {
        const hcard: MicroformatObject = {
          type: ['h-card'],
          properties: {
            name: ['John Doe']
          }
        };

        expect(isHEvent(hcard)).toBe(false);
      });
    });

    describe('isHProduct', () => {
      it('should return true for h-product microformat', () => {
        const hproduct: MicroformatObject = {
          type: ['h-product'],
          properties: {
            name: ['Widget'],
            price: ['$19.99']
          }
        };

        expect(isHProduct(hproduct)).toBe(true);
      });

      it('should return false for non h-product microformat', () => {
        const hcard: MicroformatObject = {
          type: ['h-card'],
          properties: {
            name: ['John Doe']
          }
        };

        expect(isHProduct(hcard)).toBe(false);
      });
    });

    describe('isHRecipe', () => {
      it('should return true for h-recipe microformat', () => {
        const hrecipe: MicroformatObject = {
          type: ['h-recipe'],
          properties: {
            name: ['Chocolate Cake'],
            ingredient: ['flour', 'sugar', 'cocoa']
          }
        };

        expect(isHRecipe(hrecipe)).toBe(true);
      });

      it('should return false for non h-recipe microformat', () => {
        const hcard: MicroformatObject = {
          type: ['h-card'],
          properties: {
            name: ['John Doe']
          }
        };

        expect(isHRecipe(hcard)).toBe(false);
      });
    });

    describe('isHEntry', () => {
      it('should return true for h-entry microformat', () => {
        const hentry: MicroformatObject = {
          type: ['h-entry'],
          properties: {
            name: ['Blog Post Title'],
            content: ['This is the content of the blog post']
          }
        };

        expect(isHEntry(hentry)).toBe(true);
      });

      it('should return false for non h-entry microformat', () => {
        const hcard: MicroformatObject = {
          type: ['h-card'],
          properties: {
            name: ['John Doe']
          }
        };

        expect(isHEntry(hcard)).toBe(false);
      });
    });

    describe('isSupportedMicroformat', () => {
      it('should return true for supported microformat types', () => {
        expect(isSupportedMicroformat('h-card')).toBe(true);
        expect(isSupportedMicroformat('h-event')).toBe(true);
        expect(isSupportedMicroformat('h-product')).toBe(true);
        expect(isSupportedMicroformat('h-recipe')).toBe(true);
        expect(isSupportedMicroformat('h-entry')).toBe(true);
      });

      it('should return false for unsupported microformat types', () => {
        expect(isSupportedMicroformat('h-unsupported')).toBe(false);
        expect(isSupportedMicroformat('vcard')).toBe(false);
        expect(isSupportedMicroformat('h-unknown')).toBe(false);
        expect(isSupportedMicroformat('')).toBe(false);
      });
    });
  });

  describe('Constants', () => {
    describe('SUPPORTED_MICROFORMAT_TYPES', () => {
      it('should contain all expected supported types', () => {
        expect(SUPPORTED_MICROFORMAT_TYPES).toEqual([
          'h-card',
          'h-event',
          'h-product',
          'h-recipe',
          'h-entry'
        ]);
      });

      it('should have length of 5', () => {
        expect(SUPPORTED_MICROFORMAT_TYPES).toHaveLength(5);
      });
    });
  });

  describe('Interface Compliance', () => {
    describe('HCard', () => {
      it('should accept valid h-card structure', () => {
        const hcard: HCard = {
          type: ['h-card'],
          properties: {
            name: ['John Doe'],
            url: ['https://johndoe.com'],
            email: ['john@example.com'],
            tel: ['+1-555-123-4567'],
            photo: ['https://example.com/photo.jpg'],
            org: ['Acme Corporation'],
            'job-title': ['Software Engineer'],
            note: ['A passionate developer'],
            bday: ['1990-01-01'],
            nickname: ['Johnny'],
            uid: ['john-doe-123'],
            category: ['developer', 'engineer']
          }
        };

        expect(hcard.type).toEqual(['h-card']);
        expect(hcard.properties.name).toEqual(['John Doe']);
      });
    });

    describe('HEvent', () => {
      it('should accept valid h-event structure', () => {
        const hevent: HEvent = {
          type: ['h-event'],
          properties: {
            name: ['Team Meeting'],
            summary: ['Weekly team sync'],
            start: ['2023-12-01T10:00:00'],
            end: ['2023-12-01T11:00:00'],
            duration: ['PT1H'],
            description: ['Weekly team synchronization meeting'],
            url: ['https://example.com/meeting'],
            category: ['work', 'meeting'],
            location: ['Conference Room A'],
            status: ['confirmed'],
            uid: ['meeting-123']
          }
        };

        expect(hevent.type).toEqual(['h-event']);
        expect(hevent.properties.name).toEqual(['Team Meeting']);
      });
    });

    describe('HProduct', () => {
      it('should accept valid h-product structure', () => {
        const hproduct: HProduct = {
          type: ['h-product'],
          properties: {
            name: ['Awesome Widget'],
            photo: ['https://example.com/widget.jpg'],
            brand: ['Acme Corp'],
            category: ['electronics', 'gadgets'],
            description: ['An amazing widget that does everything'],
            identifier: ['WIDGET-123'],
            url: ['https://example.com/products/widget'],
            price: ['$29.99']
          }
        };

        expect(hproduct.type).toEqual(['h-product']);
        expect(hproduct.properties.name).toEqual(['Awesome Widget']);
      });
    });

    describe('HRecipe', () => {
      it('should accept valid h-recipe structure', () => {
        const hrecipe: HRecipe = {
          type: ['h-recipe'],
          properties: {
            name: ['Chocolate Chip Cookies'],
            ingredient: ['2 cups flour', '1 cup sugar', '1/2 cup butter', 'chocolate chips'],
            yield: ['24 cookies'],
            instructions: ['Mix ingredients', 'Bake at 350°F for 12 minutes'],
            duration: ['PT30M'],
            photo: ['https://example.com/cookies.jpg'],
            summary: ['Delicious homemade chocolate chip cookies'],
            author: ['Chef Jane'],
            published: ['2023-12-01'],
            nutrition: ['250 calories per cookie'],
            category: ['dessert', 'baking']
          }
        };

        expect(hrecipe.type).toEqual(['h-recipe']);
        expect(hrecipe.properties.name).toEqual(['Chocolate Chip Cookies']);
      });
    });

    describe('HEntry', () => {
      it('should accept valid h-entry structure', () => {
        const hentry: HEntry = {
          type: ['h-entry'],
          properties: {
            name: ['My First Blog Post'],
            summary: ['An introduction to my blog'],
            content: ['This is the full content of my first blog post...'],
            published: ['2023-12-01T09:00:00'],
            updated: ['2023-12-01T10:00:00'],
            author: ['John Blogger'],
            category: ['personal', 'introduction'],
            url: ['https://blog.example.com/first-post'],
            uid: ['post-1'],
            syndication: ['https://twitter.com/user/status/123']
          }
        };

        expect(hentry.type).toEqual(['h-entry']);
        expect(hentry.properties.name).toEqual(['My First Blog Post']);
      });
    });
  });
});