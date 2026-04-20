# Forum

A web forum built with Go and SQLite featuring authentication (local + OAuth), posts, comments, likes/dislikes, image uploads, notifications, and user activity tracking.

## Features

- **Authentication** — Local register/login with bcrypt, OAuth via Google and GitHub, cookie-based sessions
- **Posts & Comments** — Create, edit, delete; associate posts with categories; filter by category
- **Likes/Dislikes** — Vote on posts and comments (registered users only)
- **Image Uploads** — Attach JPEG, PNG, or GIF images to posts (max 20MB)
- **Notifications** — Get notified when your content receives likes, dislikes, or comments
- **Activity Dashboard** — View your authored posts, reacted posts, and comments

## Getting Started

### Prerequisites

- [Go 1.19+](https://go.dev/dl/) for building from source
- [Docker](https://docs.docker.com/get-docker/) for containerized deployment

### Run with Docker

```bash
docker build -t forum .
docker run -it --rm -p 8080:8080 forum
```

### Run from Source

```bash
git clone <repository-url>
cd forum
go mod tidy
go run .
```

The server starts at **http://localhost:8080**.

## Project Structure

```
main.go                 Entry point
handlers.go             HTTP handlers
routes.go               Route definitions
authentication/         Login, register, OAuth
post/                   Post and comment CRUD, filtering
likes/                  Like/dislike logic
notifications/          Notification CRUD
db/                     SQLite initialization and schema
utils/                  Shared models and helpers
static/                 CSS, JS, templates, uploaded images
```

## Tech Stack

| Layer          | Technology                              |
|----------------|-----------------------------------------|
| Backend        | Go (standard library)                   |
| Database       | SQLite (`mattn/go-sqlite3`)             |
| Auth           | bcrypt, UUID sessions, OAuth 2.0        |
| Frontend       | HTML, CSS, vanilla JavaScript           |
| Deployment     | Docker (multi-stage Alpine build)       |