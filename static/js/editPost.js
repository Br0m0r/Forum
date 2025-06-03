document.addEventListener('DOMContentLoaded', function() {
  // Get all edit post buttons
  const editPostButtons = document.querySelectorAll('.edit-post-btn');
  
  // Add click event to each button
  editPostButtons.forEach(button => {
    button.addEventListener('click', function() {
      const postId = this.getAttribute('data-post-id');
      
      // Find the closest post-content container to this button
      const postContentContainer = this.closest('.post-content');
      
      // Find the edit form within this specific post content container
      const editForm = postContentContainer.querySelector(`#edit-post-form-${postId}`);
      
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
    });
  });
});