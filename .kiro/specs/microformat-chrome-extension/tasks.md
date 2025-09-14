# Implementation Plan

- [x] 1. Set up WXT project structure and development environment
  - Initialize new WXT project in web/chrome-extension directory
  - Configure TypeScript, build tools, and development server
  - Set up basic manifest.json with required permissions
  - Create initial project structure with entrypoints directory
  - _Requirements: 1.1_

- [x] 2. Implement microformat detection engine
- [x] 2.1 Create microformat parsing utilities
  - Install and configure microformat2 parsing library
  - Write TypeScript interfaces for microformat data structures
  - Implement parser functions for h-card, h-event, h-product, h-recipe, h-entry
  - Create unit tests for parsing accuracy with sample HTML
  - _Requirements: 1.2, 3.1, 3.2_

- [x] 2.2 Implement content script for page scanning
  - Create entrypoints/content.ts with DOM scanning logic
  - Implement efficient microformat detection without affecting page performance
  - Add message passing to communicate detected microformats to background
  - Write tests for content script functionality
  - _Requirements: 1.2, 1.3, 5.2_

- [x] 3. Create background service worker
- [x] 3.1 Implement core background service functionality
  - Create entrypoints/background.ts with extension lifecycle management
  - Set up message handling between content script and popup
  - Implement badge updates when microformats are detected
  - Add storage management for detected microformats
  - _Requirements: 1.1, 1.3_

- [x] 3.2 Add extension icon and badge management
  - Configure extension icon and badge display logic
  - Update badge count when microformats are detected
  - Handle icon state changes based on authentication status
  - _Requirements: 1.3_

- [ ] 4. Implement Solid authentication system
- [ ] 4.1 Create WebID authentication utilities
  - Write WebID validation and discovery functions
  - Implement OIDC authentication flow for Solid pods
  - Create secure session management using WXT storage
  - Add authentication state management
  - _Requirements: 2.1, 2.2, 2.3, 5.3_

- [ ] 4.2 Implement authentication UI components
  - Create login form with WebID input validation
  - Add authentication status display and logout functionality
  - Implement error handling and user feedback for auth failures
  - Write tests for authentication flow
  - _Requirements: 2.1, 2.4, 2.5_

- [ ] 5. Build popup interface
- [ ] 5.1 Create popup HTML structure and styling
  - Design entrypoints/popup.html with authentication and microformat sections
  - Implement responsive CSS styling for popup interface
  - Create loading states and error message displays
  - _Requirements: 3.1, 3.5_

- [ ] 5.2 Implement popup TypeScript functionality
  - Create entrypoints/popup.ts with UI event handling
  - Add microformat list display and selection logic
  - Implement preview panel for selected microformats
  - Connect popup to background service worker via messaging
  - _Requirements: 3.1, 3.2, 3.3, 6.1_

- [ ] 6. Implement pod communication service
- [ ] 6.1 Create RDF conversion utilities
  - Write functions to convert microformats to RDF/Turtle format
  - Map microformat types to appropriate RDF vocabularies (vCard, Schema.org)
  - Implement proper namespace handling and URI generation
  - Create unit tests for RDF conversion accuracy
  - _Requirements: 4.2_

- [ ] 6.2 Implement pod storage operations
  - Create LDP container management for different microformat types
  - Implement resource creation and storage in user's pod
  - Add duplicate detection using resource hashing
  - Handle pod communication errors and retry logic
  - _Requirements: 4.1, 4.3, 4.4, 6.2_

- [ ] 7. Add save functionality and user feedback
- [ ] 7.1 Implement save-to-pod workflow
  - Connect popup save buttons to pod storage operations
  - Add progress indicators and success/error notifications
  - Implement save status tracking and "previously saved" indicators
  - Create confirmation dialogs for duplicate handling
  - _Requirements: 4.1, 4.3, 4.4, 6.1, 6.2_

- [ ] 7.2 Add pod resource management features
  - Implement links to view saved resources in pod
  - Add resource organization by microformat type
  - Create cleanup and management utilities
  - _Requirements: 4.5, 6.3_

- [ ] 8. Implement error handling and security
- [ ] 8.1 Add comprehensive error handling
  - Implement error boundaries for all major operations
  - Add user-friendly error messages and recovery options
  - Create retry mechanisms for network failures
  - Add logging for debugging and monitoring
  - _Requirements: 2.4, 4.4, 5.4_

- [ ] 8.2 Implement security measures
  - Add input validation and sanitization for all user inputs
  - Implement secure token storage and session management
  - Add Content Security Policy and XSS protection
  - Ensure HTTPS-only communication with pods
  - _Requirements: 5.1, 5.3, 5.4_

- [ ] 9. Create comprehensive testing suite
- [ ] 9.1 Write unit tests for core functionality
  - Test microformat parsing with various HTML samples
  - Test RDF conversion accuracy for each microformat type
  - Test authentication flow and session management
  - Test storage operations and error handling
  - _Requirements: All requirements_

- [ ] 9.2 Implement integration and end-to-end tests
  - Test complete workflow from detection to pod storage
  - Test cross-browser compatibility with WXT build system
  - Test performance with large pages containing many microformats
  - Test security measures and error scenarios
  - _Requirements: All requirements_

- [ ] 10. Build and package extension
- [ ] 10.1 Configure WXT build system
  - Set up production build configuration
  - Optimize bundle size and performance
  - Configure extension packaging for Chrome Web Store
  - Create build scripts and CI/CD pipeline
  - _Requirements: 1.1_

- [ ] 10.2 Create documentation and deployment assets
  - Write user documentation and setup instructions
  - Create extension store listing materials
  - Add developer documentation for future maintenance
  - Prepare privacy policy and permissions documentation
  - _Requirements: 5.1_