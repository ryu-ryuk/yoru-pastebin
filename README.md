

<h1 align="center">
  <img src="https://raw.githubusercontent.com/ryu-ryuk/yoru-pastebin/main/docs/assets/yoru_logo.png" width="800" alt="Yoru Pastebin Banner"/>
  <img src="https://raw.githubusercontent.com/catppuccin/catppuccin/main/assets/misc/transparent.png" height="16" width="0px"/>
  
  <span style="color:#cdd6f4;">Yoru Pastebin</span>
</h1>

<h6 align="center" style="color:#bac2de;">
  A fast, secure, and ephemeral pastebin service.
</h6>

<p align="center">
  <a href="https://github.com/ryu-ryuk/yoru-pastebin/stargazers"><img src="https://img.shields.io/github/stars/ryu-ryuk/yoru-pastebin?colorA=1e1e2e&colorB=cba6f7&style=for-the-badge&logo=github&logoColor=cdd6f4"></a><a href="https://github.com/ryu-ryuk/yoru-pastebin/issues"><img src="https://img.shields.io/github/issues/ryu-ryuk/yoru-pastebin?colorA=1e1e2e&colorB=f38ba8&style=for-the-badge&logo=github&logoColor=cdd6f4"></a><a href="https://github.com/ryu-ryuk/yoru-pastebin/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-GPLv3-89b4fa?style=for-the-badge&logo=gnu&logoColor=1e1e2e&colorA=1e1e2e"></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-89b4fa?style=for-the-badge&logo=go&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/PostgreSQL-DB-b4befe?style=for-the-badge&logo=postgresql&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Built_with-Docker-94e2d5?style=for-the-badge&logo=docker&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Proxy-Traefik-fab387?style=for-the-badge&logo=traefikmesh&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Hosted_on-AWS-f9e2af?style=for-the-badge&logo=amazonaws&logoColor=white&colorA=1e1e2e" />
  <img src="https://img.shields.io/badge/Maintained-Yes-89b4fa?style=for-the-badge&logo=github&logoColor=white&colorA=1e1e2e" />
</p>


<p align="center" style="color:#a6adc8; font-size: 14.5px; line-height: 1.6; max-width: 700px; margin: auto;">
  <strong style="color:#cdd6f4;">Yoru Pastebin</strong> is a robust, privacy-focused pastebin for developers to securely share code, logs, and confidential info.<br/>
  Built with <span style="color:#89b4fa;">Go</span>, backed by <span style="color:#b4befe;">PostgreSQL</span>, and deployed using <span style="color:#94e2d5;">Docker</span> + <span style="color:#fab387;">Traefik</span> on <span style="color:#f9e2af;">AWS</span>.<br/><br/>
  <em style="color:#f38ba8;">"Yoru" (夜)</em> means <em>"night"</em> in Japanese — symbolizing secure, ephemeral, and transient pastes.
</p>


## preview

| ![Screenshot 1](/docs/assets/webui.png) | ![Screenshot 2](/docs/assets/api.png) |
|:----------------------------------------:|:----------------------------------------:|
|        *The WebUI for the tool*         |        *API usage*         |


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

Yoru Pastebin is deployed at: https://paste.alokranjan.me

### **Local API Testing Example (`curl`)**

Yes, you can absolutely test your API locally using `curl` while your Yoru Pastebin is running on `localhost:8080`.

**First, ensure your Yoru Pastebin app is running locally:**
```bash
make start
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

**Base URL:** `https://paste.alokranjan.me/api/v1`

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
  https://paste.alokranjan.me/api/v1/pastes \
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
  https://paste.alokranjan.me/api/v1/pastes \
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
  "url": "http://paste.alokranjan.me/aBcD1eFg/"
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
curl "http://paste.alokranjan.me/api/v1/pastes/aBcD1eFg"
```

**Example Request (password-protected paste):**

```bash
curl "http://paste.alokranjan.me/api/v1/pastes/xYz1w2uV?password=supersecurepassword123"
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

---

## Contributing

Contributions are welcome! If you have suggestions for improvements, bug fixes, or new features, please feel free to contribute.

## License

This project is licensed under the [**GNU General Public License v3.0 (GPLv3)**](https://www.gnu.org/licenses/gpl-3.0.html).
