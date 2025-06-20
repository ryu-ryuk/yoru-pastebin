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

---

## ☁️ Live Demo

Yoru Pastebin is deployed and available for use at: **[https://paste.alokranjan.me](https://paste.alokranjan.me)**

---

## 🍥 Features

✨ **Yoru Pastebin** is designed for privacy, clarity, and performance.
Each feature is tuned for modern dev workflows and secure sharing.

---

* **Web UI:** Minimal, dark-themed interface for creating and viewing pastes.
* **API Support:** RESTful API for developers and CLI tools.
* **Password Protection:** Encrypt content with a password for secure access.
* **Expiring Pastes:** Set pastes to self-destruct (10 min / 1 hour / 1 day / never).
* **Secure IDs:** Random, unguessable identifiers for accessing pastes.
* **Syntax Highlighting:** Auto and manual language detection for over 30 languages.
* **Dev Tools:** Line numbers, search navigation, word wrap, copy, raw mode.
* **Containerized:** Built with Docker, deployable in one command.
* **HTTPS by Default:** Secure access through Traefik reverse proxy.
* **Mocha Theme:** Styled for dark-mode lovers with [Catppuccin](https://github.com/catppuccin).

---

## (*^^*) Preview

| Web UI                                 | API Usage                                |
| :---------------------------------------: | :---------------------------------------: |
| ![Web UI Screenshot](/docs/assets/webui.png) | ![API Usage Screenshot](/docs/assets/api.png) |
| *The user-friendly interface* | *Example API interaction* |

---

## (＾＾；) Getting Started (For Development)

### Prerequisites

Ensure you have the following installed on your local machine:

* **Go** (version 1.22 or higher)
* **Docker** and **Docker Compose**
* **`make`** (for convenient development commands)
* **`git`**

### Local Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/ryu-ryuk/yoru-pastebin.git
    cd yoru-pastebin
    ```

2.  **Configure Database Credentials:**
    Edit `configs/config.toml` to set your PostgreSQL connection string. Ensure the `[database]` section's `connection_string` matches the credentials you'll use for your local Docker PostgreSQL.

    Example (`configs/config.toml`):
    ```toml
    [database]
    connection_string = "postgres://yoru_user:your_secure_password@localhost:5432/yoru_pastebin?sslmode=disable"
    ```
    *(Replace `your_secure_password` with a strong password.)*

3.  **Start Development Environment:**
    Use the `Makefile` to set up and run your local environment, including the Dockerized PostgreSQL database and applying migrations.

    ```bash
    make setup
    ```

4.  **Run the Application:**
    Once `make setup` completes, you can start your application.

    ```bash
    make run
    ```
    Your Yoru Pastebin will now be accessible in your browser at `http://localhost:8080`.

---

## (＾＾)ｂ API Reference

Yoru Pastebin provides a simple RESTful API for programmatic paste creation and retrieval.

**Base URL:** `https://paste.alokranjan.me/api/v1` (Remember to replace `paste.alokranjan.me` with your actual deployed domain if different).

---

#### **1. Create a Paste**

`POST /pastes`

Creates a new paste entry in the system.

**Request Body (JSON):**

```json
{
  "content": "Your paste content here.",
  "language": "plaintext",     // Optional: "go", "rust", "python", "json", "markdown", "auto", etc.
  "expires_in_minutes": 60,  // Optional: Integer, time until expiration in minutes (0 for never). Default from server config.
  "password": "my_secret_password" // Optional: If provided, paste content will be encrypted.
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

## 🗺️ Architecture Diagrams

Below is a visual overview of Yoru Pastebin’s architecture, themed with Catppuccin Mocha for clarity and consistency.

## Infrastructure Logic
```mermaid
graph LR
    subgraph UI["User Interaction"]
        A[Browser / Curl Client] -- HTTP Request --> B(Traefik Reverse Proxy)
    end

    subgraph INFRA["Infrastructure (Cloud VM)"]
        B -- Route HTTP/S --> C(Yoru Pastebin App Container)
        C -- DB Connection --> D(PostgreSQL DB Container)
    end

    %% Optional: Style the subgraph borders for emphasis
    style UI stroke:#b4befe,stroke-width:3px
    style INFRA stroke:#b4befe,stroke-width:3px

    %% Node styles below
    style A fill:#b4befe,stroke:#cba6f7,stroke-width:2px,color:#1e1e2e
    style B fill:#cba6f7,stroke:#b4befe,stroke-width:2px,color:#1e1e2e
    style C fill:#a6e3a1,stroke:#b4befe,stroke-width:2px,color:#1e1e2e
    style D fill:#f5c2e7,stroke:#b4befe,stroke-width:2px,color:#1e1e2e

```

For more detailed diagrams, see [docs/architecture.md](./docs/architecture.md).


## Contributing

Contributions are welcome! If you have suggestions for improvements, bug fixes, or new features, please feel free to contribute.

## License

This project is licensed under the [**GNU General Public License v3.0 (GPLv3)**](https://www.gnu.org/licenses/gpl-3.0.html).

> 🐾 The entire interface is styled using [Catppuccin Mocha](https://github.com/catppuccin/catppuccin), a beautiful pastel color palette for the modern dark aesthetic.


