## Changed Header

- The header from the homepage was different from the header on the “My Activity” page.  
- Copied the header from `home.html` to `myposts.html`.  
- Updated `handlers.go`, inside the `myPostsHandler` function, to make it work.

## Added Buttons

- Added “My Activity” and “Logout” buttons.  
- Modified the headers in the HTML files and in `style.css`.  
- Converted the “Logged in as” text into a badge.

## Added `globalHeader.html`

## Added Notifications functionality in GlobalHeader

- **sqlite.go**: 
    added notificationsTable
- **models.go**: 
    TemplateData.Notifications and .NotifCount for handlers to populate.
    added Notification model
    added DisplayText() helper ("peos liked your post” style messages)
- **notifications.go**:
    created
    functions:
        Create: adds a like/dislike/comment notification
        UnreadCount: returns the number of unread notices for the badge
        List: fetches all notifications (joined with the initiator’s username)
        MarkAllRead: marks everything read once viewed
- **likes_dislikes.go**:
    import notifications package.
    Fetch the post owner’s ID.
    Call notifications.Create(...) (if they aren’t liking their own post)
- **post.go**:
    (inside newcomment)
    Capture the newly-inserted comment’s ID
    Fetch the post owner’s ID (you already have postID)
    Call notifications.Create(...) with commentID
- **routes.go**:
    added new routes
- **handlers.go**:
    added new handlers
- **notifications.html** :
    created new file
