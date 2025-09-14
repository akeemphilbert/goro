# Design Document

## Overview

The Microformat Chrome Extension is a browser extension that detects microformats on web pages and enables users to save them to their Solid pod. The extension uses Chrome's Extension API v3 (Manifest V3), implements Solid authentication protocols, and provides a clean user interface for microformat management.

Use https://wxt.dev/ framework to setup extension

## Architecture

### Extension Architecture (WXT Framework)
```
WXT Chrome Extension (Manifest V3)
├── entrypoints/
│   ├── background.ts (Service Worker)
│   │   ├── Microformat Detection Engine
│   │   ├── Solid Authentication Manager  
│   │   └── Pod Communication Service
│   ├── content.ts (Content Script)
│   │   └── Page Microformat Scanner
│   └── popup.html + popup.ts
│       ├── Authentication UI
│       ├── Microformat List View
│       └── Save Management UI
├── components/ (Shared UI Components)
├── utils/ (Shared Utilities)
└── Storage Layer
    ├── WXT Storage API
    └── Session Management
```

### Solid Pod Integration
```
Extension → Solid Authentication → Pod Storage
    ↓              ↓                    ↓
WebID Login → OIDC/WebID-TLS → LDP Container Creation
    ↓              ↓                    ↓
Session Store → Token Management → RDF Resource Storage
```

## Components and Interfaces

### 1. Background Service Worker (`entrypoints/background.ts`)
**Purpose:** Manages extension lifecycle, authentication, and pod communication

**Key Responsibilities:**
- Handle extension installation and updates
- Manage Solid authentication flow
- Coordinate between content scripts and popup
- Handle pod API communications
- Manage persistent storage

**APIs Used:**
- WXT Storage API for session persistence
- Chrome Tabs API for active tab detection
- Chrome Runtime API for message passing
- Fetch API for Solid pod communication

### 2. Content Script (`entrypoints/content.ts`)
**Purpose:** Scans web pages for microformats without affecting page performance

**Key Responsibilities:**
- Parse DOM for microformat patterns (hCard, hEvent, hProduct, etc.)
- Extract structured data using microformat parsing libraries
- Communicate findings to background service worker
- Minimal DOM interaction to preserve page performance

**Microformat Detection:**
- Use microformat2 parsing library (microformat-node or similar)
- Support for common microformats: h-card, h-event, h-product, h-recipe, h-entry
- Extract nested properties and relationships
- Handle multiple instances per page

### 3. Popup Interface (`entrypoints/popup.html`, `entrypoints/popup.ts`)
**Purpose:** Provides user interface for authentication and microformat management

**Components:**
- **Authentication Panel:** WebID input, login/logout controls
- **Microformat List:** Detected microformats with type indicators
- **Preview Panel:** Structured data preview for selected microformats
- **Action Buttons:** Save to pod, view in pod, settings

**UI Framework:** TypeScript with WXT's built-in UI utilities and modern CSS

### 4. Solid Authentication Manager
**Purpose:** Handles WebID authentication and session management

**Authentication Flow:**
1. User enters WebID URL
2. Discover authentication endpoints from WebID profile
3. Initiate OIDC authentication flow
4. Handle callback and token exchange
5. Store session securely in Chrome storage
6. Provide authenticated requests to pod

**Security Features:**
- Secure token storage using Chrome Storage API
- Automatic token refresh
- Session timeout handling
- Logout cleanup

### 5. Pod Communication Service
**Purpose:** Manages all interactions with the user's Solid pod

**Key Operations:**
- Create containers for different microformat types
- Convert microformats to RDF (Turtle format)
- Store resources using LDP protocol
- Check for existing resources to prevent duplicates
- Handle pod errors and network issues

**RDF Mapping:**
- hCard → vCard ontology
- hEvent → Schema.org Event
- hProduct → Schema.org Product
- hRecipe → Schema.org Recipe
- Custom predicates for microformat-specific properties

## Data Models

### Microformat Data Structure
```javascript
{
  type: 'h-card' | 'h-event' | 'h-product' | 'h-recipe' | 'h-entry',
  properties: {
    name: string[],
    url: string[],
    // type-specific properties
  },
  children: MicroformatObject[],
  value: string,
  html: string,
  sourceUrl: string,
  detectedAt: Date
}
```

### Authentication Session
```javascript
{
  webId: string,
  accessToken: string,
  refreshToken: string,
  expiresAt: Date,
  podUrl: string,
  isAuthenticated: boolean
}
```

### Saved Resource Tracking
```javascript
{
  resourceUrl: string,
  microformatType: string,
  sourceUrl: string,
  savedAt: Date,
  resourceHash: string // for duplicate detection
}
```

## Error Handling

### Authentication Errors
- Invalid WebID format → User-friendly validation message
- Authentication failure → Clear error with retry option
- Network timeout → Retry mechanism with exponential backoff
- Token expiration → Automatic refresh or re-authentication prompt

### Pod Communication Errors
- Pod unavailable → Queue for retry when connection restored
- Permission denied → Clear message about pod access requirements
- Storage quota exceeded → Warning with cleanup suggestions
- Invalid RDF → Fallback to simpler format or user notification

### Microformat Detection Errors
- Parsing failures → Log error, continue with other microformats
- Invalid microformat structure → Skip malformed data
- Large page performance → Implement parsing timeouts

## Testing Strategy

### Unit Testing
- Microformat parsing accuracy with test HTML samples
- RDF conversion correctness for each microformat type
- Authentication flow state management
- Storage operations and data persistence

### Integration Testing
- End-to-end authentication with test Solid pod
- Complete save workflow from detection to pod storage
- Cross-browser compatibility (Chrome, Edge, other Chromium-based)
- Performance testing with large pages containing many microformats

### User Acceptance Testing
- Install and setup flow
- Authentication with real WebID providers
- Microformat detection on popular websites
- Save and retrieve operations with real pods
- Error scenarios and recovery

### Security Testing
- Token storage security
- XSS prevention in content scripts
- CSRF protection in authentication flow
- Data sanitization for RDF conversion

## Performance Considerations

### Content Script Optimization
- Lazy loading of microformat parsing library
- Debounced DOM scanning to avoid excessive processing
- Minimal DOM queries using efficient selectors
- Background processing for large pages

### Memory Management
- Cleanup of event listeners on page navigation
- Efficient storage of detected microformats
- Garbage collection of old session data
- Optimized RDF serialization

### Network Efficiency
- Batch pod operations when possible
- Implement request caching for pod metadata
- Use compression for large RDF payloads
- Retry logic with exponential backoff

## Security Architecture

### Content Security Policy
- Strict CSP for popup and background scripts
- No inline scripts or eval usage
- Whitelist only necessary external domains
- Secure communication channels only

### Data Protection
- Encrypt sensitive data in Chrome storage
- Clear authentication data on logout
- No persistent storage of user browsing data
- Minimal permission requests

### Pod Security
- Validate all pod responses
- Sanitize user input before RDF conversion
- Use HTTPS only for pod communication
- Implement proper CORS handling