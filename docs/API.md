# Yoru Pastebin API Reference

## Base URL

```
https://paste.alokranjan.me/api/v1
```

## Authentication

No authentication required. Rate limiting applies (60 requests/minute by default).

## Endpoints

### Create Paste

**POST** `/pastes`

Creates a new paste with optional password protection and expiration.

**Request Body:**
```json
{
  "content": "Your paste content here",
  "language": "auto",
  "expires_in_minutes": 1440,
  "password": "optional_password"
}
```

**Parameters:**
- `content` (string, required): The paste content
- `language` (string, optional): Language for syntax highlighting. Use "auto" for detection
- `expires_in_minutes` (integer, optional): Expiration time (0 = never, 10, 60, 1440, 10080)
- `password` (string, optional): Password protection using AES-256-GCM encryption

**Response (201 Created):**
```json
{
  "id": "aB3kX9mP",
  "url": "https://paste.alokranjan.me/aB3kX9mP/",
  "expires_at": "2025-07-11T06:00:00Z"
}
```

**Example:**
```bash
curl -X POST https://paste.alokranjan.me/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "func main() { fmt.Println(\"Hello, World!\") }",
    "language": "go",
    "expires_in_minutes": 1440
  }'
```

**Password-protected paste:**
```bash
curl -X POST https://paste.alokranjan.me/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "SECRET_KEY=super_classified_data",
    "language": "bash",
    "password": "secure_password_123",
    "expires_in_minutes": 60
  }'
```

### Retrieve Paste

**GET** `/pastes/{id}`

Retrieves a paste by its ID.

**Parameters:**
- `id` (path): The paste ID
- `password` (query, optional): Required for password-protected pastes

**Response (200 OK):**
```json
{
  "id": "aB3kX9mP",
  "content": "func main() { fmt.Println(\"Hello, World!\") }",
  "language": "go",
  "created_at": "2025-07-10T06:00:00Z",
  "expires_at": "2025-07-11T06:00:00Z",
  "is_file": false,
  "filename": null
}
```

**Examples:**
```bash
# Public paste
curl https://paste.alokranjan.me/api/v1/pastes/aB3kX9mP

# Password-protected paste
curl "https://paste.alokranjan.me/api/v1/pastes/aB3kX9mP?password=secure_password_123"
```

### File Upload

**POST** `/` (multipart form)

Upload files up to 30MB via multipart form data.

**Form Fields:**
- `file` (file): The file to upload
- `language` (string, optional): Language override
- `expires_in_minutes` (integer, optional): Expiration time
- `password` (string, optional): Password protection

**Example:**
```bash
curl -X POST https://paste.alokranjan.me/ \
  -F "file=@config.yaml" \
  -F "language=yaml" \
  -F "expires_in_minutes=1440"
```

## Error Responses

**400 Bad Request:**
```json
{
  "error": "Paste content cannot be empty"
}
```

**401 Unauthorized:**
```json
{
  "error": "Invalid password"
}
```

**404 Not Found:**
```json
{
  "error": "Paste not found or has expired"
}
```

**413 Payload Too Large:**
```json
{
  "error": "Content exceeds maximum size limit"
}
```

**429 Too Many Requests:**
```json
{
  "error": "Rate limit exceeded"
}
```

## Supported Languages

Auto-detection and syntax highlighting support for:

`bash`, `c`, `cpp`, `csharp`, `css`, `dart`, `diff`, `dockerfile`, `elixir`, `go`, `graphql`, `haskell`, `html`, `java`, `javascript`, `json`, `kotlin`, `lua`, `makefile`, `markdown`, `nginx`, `objectivec`, `perl`, `php`, `python`, `r`, `ruby`, `rust`, `scss`, `shell`, `sql`, `swift`, `typescript`, `vim`, `xml`, `yaml`

## Rate Limiting

- **Default limit:** 60 requests per minute per IP
- **Headers included in response:**
  - `X-RateLimit-Limit`: Request limit per window
  - `X-RateLimit-Remaining`: Requests remaining in current window
  - `X-RateLimit-Reset`: Time when rate limit resets

## Security Features

- **AES-256-GCM encryption** for password-protected pastes
- **PBKDF2** key derivation with 100,000 iterations
- **Cryptographically secure ID generation** using Base62 encoding
- **Content Security Policy** headers on all responses
- **Input validation** and sanitization
- **XSS protection** for all user content
