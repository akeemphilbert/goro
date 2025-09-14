/**
 * TypeScript interfaces for microformat data structures
 * Based on microformat2 specification
 */

export interface MicroformatProperty {
  [key: string]: string[] | MicroformatObject[];
}

export interface MicroformatObject {
  type: string[];
  properties: MicroformatProperty;
  children?: MicroformatObject[];
  value?: string;
  html?: string;
}

export interface ParsedMicroformat {
  items: MicroformatObject[];
  rels: { [key: string]: string[] };
  'rel-urls': { [key: string]: { text?: string; type?: string; hreflang?: string } };
}

// Specific microformat types
export interface HCard extends MicroformatObject {
  type: ['h-card'];
  properties: {
    name?: string[];
    url?: string[];
    email?: string[];
    tel?: string[];
    photo?: string[];
    org?: string[];
    'job-title'?: string[];
    note?: string[];
    adr?: MicroformatObject[];
    bday?: string[];
    nickname?: string[];
    uid?: string[];
    category?: string[];
    geo?: MicroformatObject[];
    key?: string[];
    logo?: string[];
    sound?: string[];
    [key: string]: string[] | MicroformatObject[] | undefined;
  };
}

export interface HEvent extends MicroformatObject {
  type: ['h-event'];
  properties: {
    name?: string[];
    summary?: string[];
    start?: string[];
    end?: string[];
    duration?: string[];
    description?: string[];
    url?: string[];
    category?: string[];
    location?: string[] | MicroformatObject[];
    geo?: MicroformatObject[];
    status?: string[];
    uid?: string[];
    dtstart?: string[];
    dtend?: string[];
    [key: string]: string[] | MicroformatObject[] | undefined;
  };
}

export interface HProduct extends MicroformatObject {
  type: ['h-product'];
  properties: {
    name?: string[];
    photo?: string[];
    brand?: string[] | MicroformatObject[];
    category?: string[];
    description?: string[];
    identifier?: string[];
    url?: string[];
    price?: string[];
    review?: MicroformatObject[];
    [key: string]: string[] | MicroformatObject[] | undefined;
  };
}

export interface HRecipe extends MicroformatObject {
  type: ['h-recipe'];
  properties: {
    name?: string[];
    ingredient?: string[];
    yield?: string[];
    instructions?: string[];
    duration?: string[];
    photo?: string[];
    summary?: string[];
    author?: string[] | MicroformatObject[];
    published?: string[];
    nutrition?: string[];
    recipe?: string[];
    category?: string[];
    [key: string]: string[] | MicroformatObject[] | undefined;
  };
}

export interface HEntry extends MicroformatObject {
  type: ['h-entry'];
  properties: {
    name?: string[];
    summary?: string[];
    content?: string[] | MicroformatObject[];
    published?: string[];
    updated?: string[];
    author?: string[] | MicroformatObject[];
    category?: string[];
    url?: string[];
    uid?: string[];
    location?: string[] | MicroformatObject[];
    syndication?: string[];
    'in-reply-to'?: string[] | MicroformatObject[];
    'like-of'?: string[] | MicroformatObject[];
    'repost-of'?: string[] | MicroformatObject[];
    [key: string]: string[] | MicroformatObject[] | undefined;
  };
}

// Detected microformat with metadata
export interface DetectedMicroformat {
  type: string;
  data: MicroformatObject;
  sourceUrl: string;
  detectedAt: Date;
  element?: Element;
  hash?: string;
}

// Supported microformat types
export type SupportedMicroformatType = 'h-card' | 'h-event' | 'h-product' | 'h-recipe' | 'h-entry';

export const SUPPORTED_MICROFORMAT_TYPES: SupportedMicroformatType[] = [
  'h-card',
  'h-event', 
  'h-product',
  'h-recipe',
  'h-entry'
];

// Type guards
export function isHCard(mf: MicroformatObject): mf is HCard {
  return mf.type.includes('h-card');
}

export function isHEvent(mf: MicroformatObject): mf is HEvent {
  return mf.type.includes('h-event');
}

export function isHProduct(mf: MicroformatObject): mf is HProduct {
  return mf.type.includes('h-product');
}

export function isHRecipe(mf: MicroformatObject): mf is HRecipe {
  return mf.type.includes('h-recipe');
}

export function isHEntry(mf: MicroformatObject): mf is HEntry {
  return mf.type.includes('h-entry');
}

export function isSupportedMicroformat(type: string): type is SupportedMicroformatType {
  return SUPPORTED_MICROFORMAT_TYPES.includes(type as SupportedMicroformatType);
}