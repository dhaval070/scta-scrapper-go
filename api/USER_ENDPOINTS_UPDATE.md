# User Endpoints Update

## Changes Made

The user management endpoints have been updated to match the actual database schema.

## Database Schema
The `users` table has the following structure:
- **username** (varchar(16), PRIMARY KEY) - User's unique username
- **password** (varchar(64)) - SHA256 + base64 encoded password
- **created_at** (timestamp) - Auto-generated creation timestamp
- **updated_at** (timestamp) - Auto-generated update timestamp

**Note:** There is NO `id` column. Username is the primary key.

## Updated Endpoints

### DELETE /users/{username}
**Changed from:** `/users/{id}` (with integer ID)
**Changed to:** `/users/{username}` (with string username)

**Example:**
```bash
# Delete user by username
DELETE /users/johndoe
```

### GET /users
Returns all users with the following fields:
```json
{
  "username": "johndoe",
  "password": "",  // Always masked/empty in response
  "created_at": "2025-12-12 15:00:00",
  "updated_at": "2025-12-12 15:00:00"
}
```

### POST /users
Create a new user with username and password:
```json
{
  "username": "newuser",
  "password": "plaintext_password"
}
```

## Updated Models

### Login Model
```go
type Login struct {
    Username  string `json:"username" gorm:"primaryKey"`
    Password  string `json:"password" gorm:"not null"`
    CreatedAt string `json:"created_at,omitempty"`
    UpdatedAt string `json:"updated_at,omitempty"`
}
```

All endpoints have been tested and the build is successful.
