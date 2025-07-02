# Forum Web Application

## Overview
A complete web forum built in Go featuring user authentication, post/comment system, image uploads, OAuth integration, and advanced features. This project implements four main components: base forum functionality, OAuth authentication, image upload capabilities, and advanced user activity tracking.

## Features

### 🔐 Authentication System
- **Local Registration/Login**: Email, username, password-based authentication
- **OAuth Integration**: Google and GitHub authentication
- **Session Management**: Cookie-based sessions with expiration
- **Password Security**: bcrypt encryption for stored passwords

### 💬 Forum Core Features
- **Posts & Comments**: Create, view, and interact with forum content
- **Categories**: Associate posts with one or more categories
- **Likes/Dislikes**: Rate posts and comments (registered users only)
- **Filtering**: Filter posts by categories, created posts, or liked posts
- **Notifications**: Real-time notification system for user interactions

### 🖼️ Image Upload
- **Post Images**: Attach images to forum posts
- **Supported Formats**: JPEG, PNG, GIF
- **Size Limit**: Maximum 20MB per image
- **Error Handling**: Proper validation and user feedback

### ⭐ Advanced Features
- **Real-time Notifications**: Users get notified when their posts are liked/disliked or commented on
- **Activity Tracking**: Personal activity page showing:
  - User's created posts
  - Posts where user left likes/dislikes
  - Comments user has made with context
- **Content Management**: Edit and delete posts and comments
- **User Activity Dashboard**: Comprehensive view of user interactions

## Quick Start

### Prerequisites
- Go 1.19+
- Docker (recommended)

### Using Docker
```bash
# Build and run with Docker
docker build -t forum .
docker run -p 8080:8080 forum
```

### Manual Setup
```bash
# Clone and run
git clone <repository-url>
cd forum
go mod tidy
go run .
```

Visit `http://localhost:8080` to access the forum.

## Project Structure

```
├── main.go              # Application entry point
├── handlers.go          # HTTP request handlers
├── routes.go           # Route definitions
├── authentication/     # OAuth and login logic
├── post/              # Post management
├── likes/             # Like/dislike system
├── notifications/     # Notification system
├── db/               # Database operations
├── static/           # CSS, JS, images
└── utils/            # Helper functions and models
```

## Database
- **SQLite**: Local database storage
- **Schema**: Users, posts, comments, likes, notifications tables
- **Security**: Parameterized queries prevent SQL injection

## Technologies Used
- **Backend**: Go (standard library + sqlite3, bcrypt, uuid)
- **Frontend**: HTML, CSS, JavaScript
- **Database**: SQLite
- **Deployment**: Docker
- **Authentication**: OAuth 2.0 (Google, GitHub)

## Key Implementation Notes
- Session-based authentication with secure cookies
- Responsive design with dark/light mode toggle
- Real-time notification updates (15-second polling)
- Image validation and secure file handling
- Comprehensive error handling and logging