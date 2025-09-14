# Microformat Chrome Extension

A Chrome browser extension that detects microformats on web pages and allows users to save them to their Solid pod.

## Development Setup

This project uses [WXT](https://wxt.dev/) framework for Chrome extension development.

### Prerequisites

- Node.js (v18 or higher)
- npm

### Installation

```bash
npm install
```

### Development

Start the development server:

```bash
npm run dev
```

This will:
- Build the extension in development mode
- Watch for file changes and rebuild automatically
- Generate the extension files in `.output/chrome-mv3/`

### Building for Production

```bash
npm run build
```

### Loading the Extension in Chrome

1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode" in the top right
3. Click "Load unpacked" and select the `.output/chrome-mv3/` directory
4. The extension should now appear in your extensions list

### Project Structure

```
entrypoints/
├── background.ts          # Service worker for extension lifecycle
├── content.ts            # Content script for microformat detection
└── popup/               # Popup interface
    ├── index.html       # Popup HTML structure
    ├── style.css        # Popup styling
    └── main.ts          # Popup functionality

components/              # Shared UI components (future)
utils/                  # Shared utilities (future)
wxt.config.ts           # WXT configuration
```

### Features (Planned)

- [x] Basic project structure and build system
- [ ] Microformat detection (h-card, h-event, h-product, etc.)
- [ ] Solid pod authentication via WebID
- [ ] RDF conversion and storage
- [ ] User interface for managing detected microformats

## License

MIT