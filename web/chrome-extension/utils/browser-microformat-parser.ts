/**
 * Browser-compatible microformat parser for content scripts
 * Uses microformat-shiv for client-side parsing
 */

import { 
  ParsedMicroformat, 
  MicroformatObject, 
  DetectedMicroformat, 
  SupportedMicroformatType,
  SUPPORTED_MICROFORMAT_TYPES,
  isSupportedMicroformat 
} from './microformat-types';

// Import microformat-shiv (browser-compatible parser)
// @ts-ignore - microformat-shiv doesn't have TypeScript definitions
import Microformats from 'microformat-shiv';

/**
 * Parse HTML content for microformats using browser-compatible parser
 */
export function parseHtmlForMicroformats(html: string, baseUrl?: string): ParsedMicroformat {
  try {
    const options = {
      baseUrl: baseUrl || window.location.href,
      textFormat: 'normalised'
    };
    
    const result = Microformats.get(options, html);
    return result as ParsedMicroformat;
  } catch (error) {
    console.error('Error parsing microformats:', error);
    return {
      items: [],
      rels: {},
      'rel-urls': {}
    };
  }
}

/**
 * Parse DOM element for microformats
 */
export function parseDomForMicroformats(element: Element, baseUrl?: string): ParsedMicroformat {
  try {
    const options = {
      baseUrl: baseUrl || window.location.href,
      textFormat: 'normalised'
    };
    
    const result = Microformats.get(options, element);
    return result as ParsedMicroformat;
  } catch (error) {
    console.error('Error parsing DOM for microformats:', error);
    return {
      items: [],
      rels: {},
      'rel-urls': {}
    };
  }
}

/**
 * Scan the entire document for microformats
 */
export function scanDocumentForMicroformats(): ParsedMicroformat {
  try {
    const options = {
      baseUrl: window.location.href,
      textFormat: 'normalised'
    };
    
    const result = Microformats.get(options, document);
    return result as ParsedMicroformat;
  } catch (error) {
    console.error('Error scanning document for microformats:', error);
    return {
      items: [],
      rels: {},
      'rel-urls': {}
    };
  }
}

/**
 * Filter microformats to only supported types
 */
export function filterSupportedMicroformats(parsed: ParsedMicroformat): MicroformatObject[] {
  return parsed.items.filter(item => 
    item.type.some(type => isSupportedMicroformat(type))
  );
}

/**
 * Get the primary type of a microformat object
 */
export function getPrimaryType(mf: MicroformatObject): string {
  // Return the first supported type, or the first type if none are supported
  const supportedType = mf.type.find(type => isSupportedMicroformat(type));
  return supportedType || mf.type[0] || 'unknown';
}

/**
 * Create a hash for a microformat object for duplicate detection
 */
export function createMicroformatHash(mf: MicroformatObject, sourceUrl: string): string {
  // Create a simple hash based on type and key properties
  const type = getPrimaryType(mf);
  const properties = JSON.stringify(mf.properties);
  const hashInput = `${type}:${sourceUrl}:${properties}`;
  
  // Simple hash function (for production, consider using crypto.subtle.digest)
  let hash = 0;
  for (let i = 0; i < hashInput.length; i++) {
    const char = hashInput.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return Math.abs(hash).toString(36);
}

/**
 * Convert parsed microformats to detected microformats with metadata
 */
export function createDetectedMicroformats(
  parsed: ParsedMicroformat, 
  sourceUrl: string
): DetectedMicroformat[] {
  const supportedMicroformats = filterSupportedMicroformats(parsed);
  
  return supportedMicroformats.map(mf => ({
    type: getPrimaryType(mf),
    data: mf,
    sourceUrl,
    detectedAt: new Date(),
    hash: createMicroformatHash(mf, sourceUrl)
  }));
}

/**
 * Get display name for a microformat
 */
export function getMicroformatDisplayName(mf: MicroformatObject): string {
  const type = getPrimaryType(mf);
  
  // Try to get a meaningful name from properties
  if (mf.properties.name && mf.properties.name.length > 0) {
    return mf.properties.name[0];
  }
  
  if (mf.properties.summary && mf.properties.summary.length > 0) {
    return mf.properties.summary[0];
  }
  
  // Fallback to type-specific logic
  switch (type) {
    case 'h-card':
      if (mf.properties.org && mf.properties.org.length > 0) {
        return mf.properties.org[0];
      }
      break;
    case 'h-event':
      if (mf.properties.start && mf.properties.start.length > 0) {
        return `Event on ${mf.properties.start[0]}`;
      }
      break;
    case 'h-product':
      if (mf.properties.brand && mf.properties.brand.length > 0) {
        return `Product by ${mf.properties.brand[0]}`;
      }
      break;
    case 'h-recipe':
      return 'Recipe';
    case 'h-entry':
      return 'Blog Entry';
  }
  
  return `${type} microformat`;
}

/**
 * Validate microformat structure
 */
export function validateMicroformat(mf: MicroformatObject): boolean {
  // Basic validation
  if (!mf.type || !Array.isArray(mf.type) || mf.type.length === 0) {
    return false;
  }
  
  if (!mf.properties || typeof mf.properties !== 'object') {
    return false;
  }
  
  // Check if it's a supported type
  const hasSupportedType = mf.type.some(type => isSupportedMicroformat(type));
  if (!hasSupportedType) {
    return false;
  }
  
  return true;
}

/**
 * Performance-optimized scanning for large pages
 */
export function scanForMicroformatsOptimized(): DetectedMicroformat[] {
  const startTime = performance.now();
  const MAX_SCAN_TIME = 100; // Maximum 100ms for scanning
  
  try {
    // First, quickly check if there are any potential microformat elements
    const potentialElements = document.querySelectorAll('[class*="h-"], [class*="p-"], [class*="u-"], [class*="dt-"], [class*="e-"]');
    
    if (potentialElements.length === 0) {
      console.log('Microformat scan completed in 0.00ms, found 0 microformats');
      return [];
    }
    
    // If we have potential elements, do a full scan
    const parsed = scanDocumentForMicroformats();
    const detected = createDetectedMicroformats(parsed, window.location.href);
    
    const endTime = performance.now();
    const scanTime = endTime - startTime;
    
    console.log(`Microformat scan completed in ${scanTime.toFixed(2)}ms, found ${detected.length} microformats`);
    
    // If scanning took too long, warn about performance
    if (scanTime > MAX_SCAN_TIME) {
      console.warn(`Microformat scanning took ${scanTime.toFixed(2)}ms, which may impact page performance`);
    }
    
    return detected;
  } catch (error) {
    console.error('Error during optimized microformat scanning:', error);
    return [];
  }
}

/**
 * Debounced scanning function to avoid excessive calls
 */
let scanTimeout: number | null = null;
export function debouncedScan(callback: (microformats: DetectedMicroformat[]) => void, delay: number = 300) {
  if (scanTimeout !== null) {
    clearTimeout(scanTimeout);
  }
  
  scanTimeout = setTimeout(() => {
    try {
      const microformats = scanForMicroformatsOptimized();
      callback(microformats);
    } catch (error) {
      console.error('Error in debounced scan:', error);
      callback([]);
    } finally {
      scanTimeout = null;
    }
  }, delay);
}