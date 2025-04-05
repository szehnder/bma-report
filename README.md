# BMA Calculator

A modern web application for generating Broker Market Analysis (BMA) reports with the help of AI.

## What is a Broker Market Analysis (BMA)?

A Broker Market Analysis (BMA) is a comprehensive report that real estate professionals use to evaluate a property's market value by comparing it to similar properties in the area. A BMA typically includes:

- Detailed property comparisons
- Price analysis
- Feature comparisons
- Market trends
- Professional recommendations

This application automates the BMA process by:
1. Collecting property data through a Chrome extension
2. Using AI to analyze and compare properties
3. Generating detailed, professional reports
4. Allowing customization through LLM instructions

## Features

- **Property Management**
  - Add properties through Chrome extension
  - Set primary and comparison properties
  - View property details and summaries

- **AI-Powered Analysis**
  - Automated property comparisons
  - Detailed market analysis
  - Customizable analysis through LLM instructions

- **Professional Reports**
  - Comprehensive BMA reports
  - PDF export functionality
  - Real-time report updates

## Setup Instructions

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- MongoDB
- Chrome browser

### Environment Variables

Create a `.envrc` file in the root directory with the following variables:

```bash
# MongoDB
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DB="bma_calculator"

# Backend
export BACKEND_PORT="8080"
export API_URL="http://localhost:8080"

# Frontend
export VITE_API_URL="http://localhost:8080"

# LLM Configuration
export GEMINI_API_KEY="your_gemini_api_key"
```

### Installation

1. **Backend Setup**
   ```bash
   cd pkg/backend
   go mod download
   go run main.go
   ```

2. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. **Chrome Extension Setup**
   - Open Chrome and go to `chrome://extensions/`
   - Enable "Developer mode"
   - Click "Load unpacked"
   - Select the `extension` directory

### Usage

1. **Adding Properties**
   - Install the Chrome extension
   - Navigate to a property listing page
   - Click the extension icon to save the property

2. **Generating Reports**
   - Open the BMA Calculator web application
   - Set a primary property
   - Enable comparison properties
   - View and download the BMA report

3. **Customizing Analysis**
   - Edit LLM instructions to customize the analysis
   - Refresh the report to apply changes

## Development

### Project Structure

```
.
├── frontend/           # Svelte frontend application
├── pkg/
│   └── backend/       # Go backend server
└── extension/         # Chrome extension
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License 