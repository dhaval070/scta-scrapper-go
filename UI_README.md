# Hockey Calendar Scraper UI

A React-based UI for managing site configurations for the Hockey Calendar Scraper.

## Features

- ✅ View all site configurations
- ✅ Add new site configurations
- ✅ Edit existing site configurations
- ✅ Delete site configurations
- ✅ Enable/disable sites
- ✅ JSON editor for parser configurations

## Architecture

### Backend API (Go + Echo)
- Location: `cmd/api/main.go`
- Framework: Echo v4
- Database: MySQL
- Port: 8080 (default)

### Frontend UI (React + Vite)
- Location: `ui/`
- Framework: React 18 + Vite
- Port: 5173 (default)

## Prerequisites

1. **Go 1.23+** installed
2. **Node.js 18+** and npm installed
3. **MySQL database** running with `sites_config` table
4. Database migration completed: `20251102130500_sites-config.up.sql`

## Setup & Installation

### 1. Build the API Server

```bash
# From project root
go build -o bin/api ./cmd/api/
```

### 2. Install UI Dependencies

```bash
cd ui
npm install
```

## Running the Application

### Start the API Server

```bash
# From project root
# Uses default database connection: root:root@tcp(127.0.0.1:3306)/hockey_calendar

./bin/api

# Or with custom database connection:
DB_DSN="user:pass@tcp(host:port)/dbname?parseTime=true" ./bin/api

# Or with custom port:
PORT=8080 ./bin/api
```

The API will start on `http://localhost:8080`

### Start the UI Development Server

```bash
cd ui
npm run dev
```

The UI will start on `http://localhost:5173`

Open your browser and navigate to: **http://localhost:5173**

## API Endpoints

### Base URL: `http://localhost:8080/api`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/sites` | List all site configurations |
| GET | `/sites/:id` | Get a specific site configuration |
| POST | `/sites` | Create a new site configuration |
| PUT | `/sites/:id` | Update a site configuration |
| DELETE | `/sites/:id` | Delete a site configuration |

### Example API Request

```bash
# List all sites
curl http://localhost:8080/api/sites

# Get a specific site
curl http://localhost:8080/api/sites/1

# Create a new site
curl -X POST http://localhost:8080/api/sites \
  -H "Content-Type: application/json" \
  -d '{
    "site_name": "testsite",
    "display_name": "Test Site",
    "base_url": "https://testsite.com/",
    "home_team": "test",
    "parser_type": "day_details",
    "parser_config": {
      "url_template": "Calendar/?Month=%d&Year=%d",
      "tournament_check_exact": true
    },
    "enabled": true,
    "scrape_frequency_hours": 24,
    "notes": "Test site configuration"
  }'
```

## UI Features

### Site Configs List
- View all sites in a sortable table
- See status (enabled/disabled)
- See parser type
- See last scraped timestamp
- Quick edit/delete actions

### Add/Edit Site Config
- Modal form with validation
- Fields:
  - Site Name (required)
  - Display Name (required)
  - Base URL (required)
  - Home Team
  - Parser Type (dropdown)
  - Parser Config (JSON editor)
  - Scrape Frequency (hours)
  - Notes
  - Enabled checkbox

### Parser Config JSON Editor
- Syntax highlighting (monospace font)
- JSON validation on save
- Examples per parser type

## Parser Types

The UI supports the following parser types:

1. **day_details** - Standard day details parser
2. **day_details_parser1** - Alternative day details parser
3. **day_details_parser2** - Second alternative day details parser
4. **month_based** - Month-based schedule parser
5. **group_based** - Group-based parser with seasons
6. **external** - External binary/script parser
7. **custom** - Custom implementation

## Parser Config Examples

### Day Details Parser
```json
{
  "url_template": "Calendar/?Month=%d&Year=%d",
  "tournament_check_exact": true,
  "log_errors": true,
  "content_filter": "REGULAR SEASON"
}
```

### Group Based Parser
```json
{
  "group_xpath": "//div[@class='site-list']/div/a",
  "group_url_template": "Groups/%s/Calendar/?Month=%d&Year=%d",
  "seasons_url": "Seasons/Current/"
}
```

### External Parser
```json
{
  "binary_path": "./bin/custom-scraper"
}
```

### Month Based Parser
```json
{
  "team_parse_strategy": "opponent",
  "url_prefix": "https://stats.example.com/"
}
```

## Environment Variables

### API Server

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DSN` | MySQL connection string | `root:root@tcp(127.0.0.1:3306)/hockey_calendar?parseTime=true` |
| `PORT` | API server port | `8080` |

### UI (Vite)

The UI is configured to connect to the API at `http://localhost:8080/api`.

To change this, edit `ui/src/components/SiteConfigs.jsx`:
```javascript
const API_BASE = 'http://localhost:8080/api'
```

## Building for Production

### Build API Binary

```bash
go build -o bin/api ./cmd/api/
```

### Build UI for Production

```bash
cd ui
npm run build
```

The production build will be in `ui/dist/`.

### Serve Production Build

You can serve the production build with any static file server:

```bash
cd ui/dist
python3 -m http.server 5173
```

Or use a production web server like Nginx or Apache.

## Development

### Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # API server
├── ui/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SiteConfigs.jsx   # Main site list component
│   │   │   └── SiteEditModal.jsx # Edit/add modal
│   │   ├── App.jsx               # Main app component
│   │   ├── App.css               # Styles
│   │   └── main.jsx              # Entry point
│   ├── package.json
│   └── vite.config.js
└── database/
    └── migrations/
        └── 20251102130500_sites-config.up.sql
```

## Troubleshooting

### API Connection Error
- Ensure MySQL is running
- Check database credentials in `DB_DSN`
- Verify the `sites_config` table exists
- Check API server is running on port 8080

### CORS Error
- API server has CORS enabled for `localhost:5173` and `localhost:3000`
- If using different ports, update CORS settings in `cmd/api/main.go`

### JSON Parse Error
- Ensure parser_config is valid JSON
- Use JSON validator before saving
- Check for trailing commas or syntax errors

## License

Same as main project.
