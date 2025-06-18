# Yoru Pastebin

A fast, secure, and ephemeral pastebin service built with Go, PostgreSQL, and Docker.

## Features

* **Web UI:** Simple interface for creating and viewing pastes.
* **API:** Programmatic access for developers to integrate Yoru into their tools.
* **Unique, Unguessable IDs:** Securely generated paste identifiers.
* **Configurable Expiration:** Set pastes to auto-delete after a specified time (e.g., 10m, 1h, 1d, never).
* **Password Protection:** Encrypt paste content with a password, viewable only with correct password.
* **Language Syntax Highlighting:** Automatic or user-specified highlighting for popular languages (Go, Rust, Python, Markdown, JSON, etc.).
* **Developer-Friendly Viewer:** Line numbers, search with navigation, word wrap toggle, copy raw content, and copy share link.
* **Containerized:** Easy deployment using Docker and Docker Compose.
* **Secure Communications:** Designed for HTTPS (via Traefik in recommended deployments).

## Getting Started (For Users/Developers)

### Try it Live!

Yoru Pastebin is deployed at: ``

### **Local API Testing Example (`curl`)**

Yes, you can absolutely test your API locally using `curl` while your Yoru Pastebin is running on `localhost:8080`.

**First, ensure your Yoru Pastebin app is running locally:**
```bash
make run
```

**Then, open a *new* terminal window and use these `curl` commands:**

#### **1. Create a Paste (POST /api/v1/pastes)**

**Example 2: Plain text paste with password**

```bash
curl -X POST \
  http://localhost:8080/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This is a very secret message.",
    "language": "plaintext",
    "password": "bohooo"
  }'
```

**Expected Local Output (for successful creation):**

```json
{"id":"<generated_id>","url":"http://localhost:8080/<generated_id>/"}
```
You can then open `http://localhost:8080/<generated_id>/` in your browser to verify it.

#### **2. Retrieve a Paste (GET /api/v1/pastes/{id})**

**Example 1: Retrieve a public paste**
(Replace `<PUBLIC_PASTE_ID>` with an ID you just created without a password)

```bash
curl http://localhost:8080/api/v1/pastes/<PUBLIC_PASTE_ID>
```

**Expected Local Output (for successful retrieval):**

```json
{"id":"<PUBLIC_PASTE_ID>","content":"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello from API test!\")\n}","language":"go","created_at":"2025-06-18T14:45:00Z","expires_at":"2025-06-18T14:55:00Z"}
```

**Example 2: Retrieve a password-protected paste (with correct password)**
(Replace `<PROTECTED_PASTE_ID>` and `myapipassword` accordingly)

```bash
curl "http://localhost:8080/api/v1/pastes/<PROTECTED_PASTE_ID>?password=myapipassword"
```
**Important:** Double quotes are needed around the URL if it contains `?` or `&` in Bash.

**Expected Local Output:** The decrypted content as JSON.

**Example 3: Retrieve a password-protected paste (with incorrect password)**

```bash
curl http://localhost:8080/api/v1/pastes/<PROTECTED_PASTE_ID>?password=wrongpass
```

**Expected Local Output:**

```json
{"error":"Incorrect password."}
```

---

### API Reference

Yoru Pastebin provides a simple RESTful API for programmatic paste creation and retrieval.

**Base URL:** `https://localhost:8080/api/v1`

---

#### **1. Create a Paste**

`POST /pastes`

**Request Body (JSON):**

```json
{
  "content": "Your paste content here.",
  "language": "plaintext",     // Optional: "go", "rust", "python", "json", "markdown", "auto", etc.
  "expires_in_minutes": 60,  // Optional: Integer, time until expiration in minutes (0 for never). Default from server config.
  "password": "scary_pass" // Optional: If provided, paste content will be encrypted.
}
```

**`expires_in_minutes` Options:**

* `0`: Never expires.
* `10`: 10 minutes.
* `60`: 1 hour.
* `1440`: 1 day.
* `10080`: 1 week.
* `43200`: 1 month.

**Example Request (using `curl`):**

```bash
curl -X POST \
  https://localhost:8080/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "func main() {\n  fmt.Println(\"Hello, API!\")\n}",
    "language": "go",
    "expires_in_minutes": 10
  }'
```

**Example Request (with password):**

```bash
curl -X POST \
  https://localhost:8080/api/v1/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This is very sensitive data.",
    "language": "plaintext",
    "password": "supersecurepassword123"
  }'
```

**Successful Response (HTTP 201 Created):**

```json
{
  "id": "aBcD1eFg",
  "url": "[http://paste.alokranjan.me/aBcD1eFg/](http://paste.alokranjan.me/aBcD1eFg/)"
}
```

**Error Response (Example - HTTP 400 Bad Request):**

```json
{
  "error": "Paste content cannot be empty."
}
```
Other error codes include `413 Request Entity Too Large`, `500 Internal Server Error`.

---

#### **2. Retrieve a Paste**

`GET /pastes/{id}`

**Parameters:**

* `id`: The unique ID of the paste.
* `password` (Query Parameter, Optional): Required if the paste is password-protected.

**Example Request (public paste):**

```bash
curl "http://localhost:8080/api/v1/pastes/aBcD1eFg"
```

**Example Request (password-protected paste):**

```bash
curl "http://localhost:8080/api/v1/pastes/xYz1w2uV?password=supersecurepassword123"
```

**Successful Response (HTTP 200 OK):**

```json
{
  "id": "aBcD1eFg",
  "content": "func main() {\n  fmt.Println(\"Hello, API!\")\n}",
  "language": "go",
  "created_at": "2025-06-18T10:00:00Z",
  "expires_at": "2025-06-18T10:10:00Z"
}
```

**Error Response (Example - HTTP 404 Not Found):**

```json
{
  "error": "Paste not found or has expired."
}
```
Other error codes include `401 Unauthorized` (for incorrect password), `500 Internal Server Error`.

## preview

![image](https://github.com/user-attachments/assets/999773b4-c5da-4b1f-b889-b526b2eee52a)
| ![Screenshot 1](/docs/assets/webui.png) | ![Screenshot 2](/docs/assets/api.png) |
|:----------------------------------------:|:----------------------------------------:|
|        *The WebUI for the tool*         |        *API usage*         |


# docker setup | to be updated
```sh
docker run --name yoru-postgres \
  -e POSTGRES_USER=ryu \
  -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_DB=yoru_pastebin \
  -p 5432:5432 \
  -d postgres:16-alpine
```


## Contributing

Contributions are welcome! If you have suggestions for improvements, bug fixes, or new features, please feel free to contribute.

## License

This project is licensed under the [**GNU General Public License v3.0 (GPLv3)**](https://www.gnu.org/licenses/gpl-3.0.html).
