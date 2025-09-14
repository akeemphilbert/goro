/**
 * Microformat parsing utilities using microformat-node library
 */

import { 
  ParsedMicroformat, 
  MicroformatObject, 
  DetectedMicroformat, 
  SupportedMicroformatType,
  SUPPORTED_MICROFORMAT_TYPES,
  isSupportedMicroformat 
} from './microformat-types';

// Import microformat-node (will be available in browser context)
declare const mfParser: any;

/**
 * Parse HTML content for microformats
 */
export function parseHtmlForMicroformats(html: string, baseUrl?: string): ParsedMicroformat {
  try {
    // Use microformat-node parser
    const result = mfParser(html, { baseUrl });
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
    const html = element.outerHTML;
    return parseHtmlForMicroformats(html, baseUrl);
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
 * Filter microformats to only supported types
 */
export function filterSupportedMicroformats(parsed: ParsedMicroformat): MicroformatObject[] {
  return parsed.items.filter(item => 
    item.type.some(type => isSupportedMicroformat(type))
  );
}

/**
 * Extract specific microformat type from parsed results
 */
export function extractMicroformatsByType(
  parsed: ParsedMicroformat, 
  type: SupportedMicroformatType
): MicroformatObject[] {
  return parsed.items.filter(item => item.type.includes(type));
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
 * Get a short description for a microformat
 */
export function getMicroformatDescription(mf: MicroformatObject): string {
  const type = getPrimaryType(mf);
  
  // Type-specific descriptions first (more useful for UI)
  switch (type) {
    case 'h-card':
      const parts = [];
      if (mf.properties.org && mf.properties.org.length > 0) {
        parts.push(mf.properties.org[0]);
      }
      if (mf.properties['job-title'] && mf.properties['job-title'].length > 0) {
        parts.push(mf.properties['job-title'][0]);
      }
      if (parts.length > 0) {
        return parts.join(' - ');
      }
      break;
      
    case 'h-event':
      const eventParts = [];
      if (mf.properties.start && mf.properties.start.length > 0) {
        eventParts.push(`Starts: ${mf.properties.start[0]}`);
      }
      if (mf.properties.location && mf.properties.location.length > 0) {
        eventParts.push(`Location: ${mf.properties.location[0]}`);
      }
      if (eventParts.length > 0) {
        return eventParts.join(' | ');
      }
      break;
      
    case 'h-product':
      const productParts = [];
      if (mf.properties.price && mf.properties.price.length > 0) {
        productParts.push(`Price: ${mf.properties.price[0]}`);
      }
      if (mf.properties.category && mf.properties.category.length > 0) {
        productParts.push(`Category: ${mf.properties.category[0]}`);
      }
      if (productParts.length > 0) {
        return productParts.join(' | ');
      }
      break;
      
    case 'h-recipe':
      const recipeParts = [];
      if (mf.properties.yield && mf.properties.yield.length > 0) {
        recipeParts.push(`Serves: ${mf.properties.yield[0]}`);
      }
      if (mf.properties.duration && mf.properties.duration.length > 0) {
        recipeParts.push(`Duration: ${mf.properties.duration[0]}`);
      }
      if (recipeParts.length > 0) {
        return recipeParts.join(' | ');
      }
      break;
      
    case 'h-entry':
      const entryParts = [];
      if (mf.properties.published && mf.properties.published.length > 0) {
        entryParts.push(`Published: ${mf.properties.published[0]}`);
      }
      if (mf.properties.author && mf.properties.author.length > 0) {
        entryParts.push(`By: ${mf.properties.author[0]}`);
      }
      if (entryParts.length > 0) {
        return entryParts.join(' | ');
      }
      break;
  }
  
  // Fallback to generic descriptions
  if (mf.properties.description && mf.properties.description.length > 0) {
    return mf.properties.description[0].substring(0, 100) + '...';
  }
  
  if (mf.properties.summary && mf.properties.summary.length > 0) {
    return mf.properties.summary[0].substring(0, 100) + '...';
  }
  
  if (mf.properties.note && mf.properties.note.length > 0) {
    return mf.properties.note[0].substring(0, 100) + '...';
  }
  
  // Final fallbacks
  switch (type) {
    case 'h-card':
      return 'Contact information';
    case 'h-event':
      return 'Event information';
    case 'h-product':
      return 'Product information';
    case 'h-recipe':
      return 'Recipe information';
    case 'h-entry':
      return 'Blog entry';
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
 * Clean and normalize microformat data
 */
export function normalizeMicroformat(mf: MicroformatObject): MicroformatObject {
  const normalized = { ...mf };
  
  // Ensure all property values are arrays
  Object.keys(normalized.properties).forEach(key => {
    const value = normalized.properties[key];
    if (!Array.isArray(value)) {
      normalized.properties[key] = [value as string];
    }
  });
  
  // Remove empty properties
  Object.keys(normalized.properties).forEach(key => {
    const value = normalized.properties[key];
    if (Array.isArray(value) && value.length === 0) {
      delete normalized.properties[key];
    }
  });
  
  return normalized;
}