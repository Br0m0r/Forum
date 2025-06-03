document.addEventListener('DOMContentLoaded', function() {
  // Get all edit comment buttons
  const editCommentButtons = document.querySelectorAll('.edit-comment-btn');
  
  // Add click event to each button
  editCommentButtons.forEach(button => {
    button.addEventListener('click', function() {
      const commentId = this.getAttribute('data-comment-id');
      
      // Find the closest container to this button
      // Updated: Look for post-content instead of user-comment
      const commentContainer = this.closest('.post-content');
      
      if (commentContainer) {
        // Find the edit form within this specific container
        const editForm = commentContainer.querySelector(`#edit-comment-form-${commentId}`);
        
        if (editForm) {
          // Toggle the form's display
          if (editForm.style.display === 'block') {
            editForm.style.display = 'none';
          } else {
            // Hide all other edit forms first
            document.querySelectorAll('.edit-form').forEach(form => {
              form.style.display = 'none';
            });
            
            // Show this specific form
            editForm.style.display = 'block';
          }
        }
      }
    });
  });
});