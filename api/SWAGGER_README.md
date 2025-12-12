# Swagger API Documentation

## Overview
Swagger/OpenAPI documentation has been successfully generated for the Surface API.

## Accessing the Documentation

Once the API server is running, access the interactive Swagger UI at:
- **Local**: http://localhost:8000/swagger/index.html
- **Production**: https://your-domain.com/swagger/index.html

## Documented Endpoints

### Authentication
- **POST /login** - User login
- **GET /logout** - User logout
- **GET /session** - Check current session

### User Management
- **GET /users** - List all users (passwords masked)
- **POST /users** - Create a new user
- **DELETE /users/{id}** - Delete a user and invalidate their sessions

### Sites
- **GET /sites** - List all configured sites

### Surfaces
- **GET /surfaces** - List all surfaces (with optional province filter)

### Locations
- **GET /locations** - List locations with pagination and filters

### Events
- **GET /events** - List events with pagination and filters (site, surface_id, date range)

## Security
All endpoints (except /login) require authentication via session cookie (CookieAuth).

## Regenerating Documentation

After modifying endpoint annotations, regenerate the docs:
```bash
~/go/bin/swag init -g main.go
```

## Swagger Annotations Format

Example:
```go
// @Summary Short description
// @Description Detailed description
// @Tags TagName
// @Accept json
// @Produce json
// @Param paramName query string false "Description"
// @Success 200 {object} ResponseType
// @Failure 400 {object} map[string]interface{} "error"
// @Security CookieAuth
// @Router /endpoint [method]
func handlerName(c *gin.Context) {
    // implementation
}
```

## Files Generated
- `docs/docs.go` - Generated Go code
- `docs/swagger.json` - OpenAPI specification (JSON)
- `docs/swagger.yaml` - OpenAPI specification (YAML)
