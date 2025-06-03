# Forum Notification System

## Overview
The forum application includes a comprehensive notification system that alerts users when someone interacts with their content. The system is designed to be efficient, responsive, and provide detailed tracking of user interactions.

## Core Components

### Backend (notifications.go)
The notification backend provides four primary functions:

#### 1. Create()
- Creates new notification records when users interact with content
- Parameters:
  - `userID`: User receiving the notification
  - `initiatorID`: User who triggered the action (liker, commenter)
  - `postID`: Related post ID
  - `commentID`: Optional comment reference (nil for post notifications)
  - `kind`: Notification type (like, dislike, comment)
- Handles nullable comment IDs for post-level notifications
- Includes detailed logging for debugging

#### 2. UnreadCount()
- Returns the count of unread notifications for a specific user
- Used to display the notification badge in the UI
- Simple and efficient query

#### 3. List()
- Retrieves all notifications for a user, sorted by newest first
- Joins with users table to include initiator information
- Properly handles nullable comment IDs
- Returns complete notification objects with all metadata
- Includes comprehensive logging

#### 4. MarkAllRead()
- Marks all notifications as read for a specific user
- Used when a user views their notifications page
- Reports count of affected notifications for logging

## Implementation Details

- **Security**: Uses parameterized SQL queries to prevent SQL injection
- **Data Integrity**: Transaction-safe database operations
- **Error Handling**: Thorough error handling with contextual logging
- **Flexibility**: Support for multiple notification types (likes, dislikes, comments)
- **Robustness**: Proper handling of nullable fields with `sql.NullInt64`
- **Debugging**: Extensive logging throughout for debugging and audit trails

## Notification Types

The system supports several notification types:
- 👍 Post likes/dislikes
- 💬 Comment likes/dislikes
- 📝 New comments on posts

## Data Flow

1. User action triggers notification creation (like, comment, etc.)
2. Backend `Create()` function inserts notification record
3. Frontend polls `UnreadCount()` to display notification badge
4. User views notifications via `List()` function
5. Viewing notifications page triggers `MarkAllRead()`

## Integration Points

| Component | Integration |
|-----------|-------------|
| Frontend | Polling mechanism updates notification count every 15 seconds |
| User Interface | Badge in header shows unread count |
| Database | Uses transactions for data integrity |
| Logging | Detailed logs for monitoring and debugging |

The notification system effectively separates database operations from presentation logic, creating a maintainable and extensible notification framework.