# Adrenochain Explorer Web Interface

This package provides a complete web interface for the Adrenochain blockchain explorer, built with Go HTML templates, modern CSS, and vanilla JavaScript.

## Features

- **Dashboard**: Overview of blockchain statistics and recent activity
- **Block Explorer**: Browse and view detailed block information
- **Transaction Explorer**: View transaction details and history
- **Address Explorer**: Check address balances and transaction history
- **Search**: Search for blocks, transactions, and addresses
- **Responsive Design**: Works on all device sizes
- **Modern UI**: Clean, professional interface with smooth animations

## Architecture

### Components

- **WebHandler**: HTTP handlers for web pages
- **Templates**: HTML template management with Go templates
- **WebServer**: HTTP server configuration and management
- **Routes**: URL routing for web pages

### Template System

The web interface uses Go's built-in HTML template system with:

- Base template with common layout
- Page-specific content templates
- Template helper functions for formatting
- Responsive CSS framework

### Static Assets

- **CSS**: Modern styling with CSS Grid and Flexbox
- **JavaScript**: Vanilla ES6+ with modern features
- **Icons**: Simple favicon and UI elements

## Usage

### Basic Setup

```go
package main

import (
    "log"
    "github.com/palaseus/adrenochain/pkg/explorer/service"
"github.com/palaseus/adrenochain/pkg/explorer/web"
)

func main() {
    // Create explorer service
    explorerService := service.NewExplorerService(...)
    
    // Create web server
    webServer := web.NewWebServer(explorerService)
    
    // Start web interface on port 8080
    log.Fatal(webServer.Start(8080))
}
```

### Integration with Existing API

The web interface can run alongside the existing API:

```go
// Start API server on port 8081
go func() {
    apiServer := api.NewAPIServer(explorerService)
    log.Fatal(apiServer.Start(8081))
}()

// Start web interface on port 8080
webServer := web.NewWebServer(explorerService)
log.Fatal(webServer.Start(8080))
```

### Custom Configuration

```go
webServer := web.NewWebServer(explorerService)

// Access handler for custom routing
handler := webServer.GetHandler()

// Access templates for custom rendering
templates := webServer.GetTemplates()

// Custom health check
err := webServer.HealthCheck()
if err != nil {
    log.Printf("Web server health check failed: %v", err)
}
```

## API Endpoints

The web interface provides these HTML pages:

- `/` - Dashboard
- `/blocks` - Block list
- `/blocks/{hash}` - Block details
- `/transactions` - Transaction list
- `/transactions/{hash}` - Transaction details
- `/addresses/{address}` - Address details
- `/search` - Search interface

## Template Functions

The templates include helper functions:

- `formatHash(hash)` - Format hash for display
- `formatAddress(address)` - Format address for display
- `formatAmount(amount)` - Format amount in satoshis
- `formatTime(timestamp)` - Format timestamp
- `formatDifficulty(difficulty)` - Format difficulty

## Styling

The CSS provides:

- **Responsive Grid**: CSS Grid for layout
- **Modern Colors**: Professional color scheme
- **Smooth Animations**: Hover effects and transitions
- **Mobile First**: Responsive design for all devices
- **Accessibility**: High contrast and readable fonts

## JavaScript Features

- **Search Suggestions**: Real-time search with autocomplete
- **Copy to Clipboard**: One-click copying of hashes and addresses
- **Mobile Navigation**: Responsive mobile menu
- **Auto-refresh**: Dashboard updates every 30 seconds
- **Performance Monitoring**: Built-in performance tracking

## Testing

Run the test suite:

```bash
cd pkg/explorer/web
go test -v
```

The tests cover:

- Handler functionality
- Template rendering
- Error handling
- Pagination helpers
- Server management

## Performance

The web interface is optimized for:

- **Fast Loading**: Sub-second page loads
- **Efficient Rendering**: Server-side rendering with Go templates
- **Minimal JavaScript**: Lightweight client-side code
- **Caching**: Built-in caching support
- **Compression**: Gzip compression for static assets

## Security

Built-in security features:

- **Input Validation**: All user inputs validated
- **XSS Protection**: Template escaping
- **CSRF Protection**: Form token validation
- **Security Headers**: HSTS, CSP, and other headers
- **Rate Limiting**: Built-in rate limiting support

## Deployment

### Production Considerations

1. **HTTPS**: Always use HTTPS in production
2. **Reverse Proxy**: Use nginx or similar for static file serving
3. **Caching**: Implement Redis or similar for caching
4. **Monitoring**: Add Prometheus metrics and logging
5. **Load Balancing**: Use multiple web server instances

### Docker Support

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o explorer ./cmd/explorer

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/explorer .
EXPOSE 8080
CMD ["./explorer"]
```

## Contributing

When contributing to the web interface:

1. **Follow Go conventions**: Use standard Go formatting
2. **Test coverage**: Maintain 95%+ test coverage
3. **Responsive design**: Ensure mobile compatibility
4. **Accessibility**: Follow WCAG 2.1 AA guidelines
5. **Performance**: Keep page load times under 1 second

## Roadmap

Future enhancements planned:

- **Charts**: Interactive charts for blockchain metrics
- **Real-time Updates**: WebSocket support for live data
- **Advanced Search**: Full-text search capabilities
- **API Documentation**: Interactive API docs
- **Multi-language**: Internationalization support
- **Dark Mode**: Theme switching capability
- **Export Features**: CSV/JSON data export
- **Mobile App**: Progressive Web App support

## License

This package is part of the Adrenochain project and follows the same license terms.
