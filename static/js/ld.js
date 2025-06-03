// Debug version of ld.js with console logging

// syncSection: highlights the correct arrow based on data-user-vote
function syncSection(section) {
  // First remove active class from all buttons in this section
  section.querySelectorAll('.vote-btn').forEach(btn => btn.classList.remove('active'));
  
  // Get the user's vote value and parse it as integer
  const uv = parseInt(section.dataset.userVote, 10) || 0;
  
  // Add active class to the appropriate button based on data-is-like attribute
  if (uv === 1) {
    // Find the like button (has data-is-like="1")
    const likeBtn = section.querySelector('.vote-btn[data-is-like="1"]');
    if (likeBtn) likeBtn.classList.add('active');
  } else if (uv === -1) {
    // Find the dislike button (has data-is-like="0")
    const dislikeBtn = section.querySelector('.vote-btn[data-is-like="0"]');
    if (dislikeBtn) dislikeBtn.classList.add('active');
  }
}
// Initialize vote buttons and sync on DOM ready
document.addEventListener('DOMContentLoaded', () => {
  console.log('DOM loaded, initializing vote buttons');
  
  // Initial sync of all vote sections
  document.querySelectorAll('.vote-section').forEach(section => {
    console.log('Initial sync for section:', section);
    syncSection(section);
  });
  
  // Attach click handlers
  document.querySelectorAll('.vote-btn').forEach(btn => {
    btn.addEventListener('click', async (e) => {
      console.log('Button clicked:', btn);
      console.log('Button data:', {
        postId: btn.dataset.postId,
        commentId: btn.dataset.commentId,
        isLike: btn.dataset.isLike
      });
      
      const postId    = btn.dataset.postId;
      const commentId = btn.dataset.commentId;
      const isLikeVal = btn.dataset.isLike;   // "1" or "0"
      let url;

      // build request URL
      if (commentId) {
        url = `/comments/like?comment_id=${commentId}&is_like=${isLikeVal}`;
      } else if (postId) {
        url = `/posts/like?post_id=${postId}&is_like=${isLikeVal}`;
      } else {
        console.error('Missing postId or commentId');
        return;
      }
      console.log('Request URL:', url);

      // determine new vote state: 0 (unlike) if same as before, otherwise 1 or -1
      const parent    = btn.closest('.vote-section');
      const prevUV    = parseInt(parent.dataset.userVote, 10) || 0;
      const clickedUV = isLikeVal === '1' ? 1 : -1;
      const newUV     = (prevUV === clickedUV) ? 0 : clickedUV;
      
      console.log('Vote state:', { prevUV, clickedUV, newUV });

      try {
        // Update UI immediately for better user feedback
        console.log('Setting new userVote value:', newUV);
        parent.dataset.userVote = newUV;
        syncSection(parent);

        // send vote to server
        console.log('Sending request to server...');
        const res = await fetch(url, {
          method: 'POST',
          credentials: 'same-origin'
        });
        
        if (!res.ok) {
          console.error('Server response not OK:', res.status);
          // If server request fails, revert the UI
          parent.dataset.userVote = prevUV;
          syncSection(parent);
          return;
        }
        
        const data = await res.json();
        console.log('Server response:', data);

        // update ALL matching sections
        if (postId) {
          document.querySelectorAll(`.vote-section[data-post-id="${postId}"]`)
            .forEach(section => {
              console.log('Updating post section:', section);
              section.querySelector('.vote-count.likes').textContent    = data.Likes_count;
              section.querySelector('.vote-count.dislikes').textContent = data.Dislikes_count;
              section.dataset.userVote = newUV;
              syncSection(section);
            });
        } else {
          document.querySelectorAll(`.vote-section[data-comment-id="${commentId}"]`)
            .forEach(section => {
              console.log('Updating comment section:', section);
              section.querySelector('.vote-count.likes').textContent    = data.Likes_count;
              section.querySelector('.vote-count.dislikes').textContent = data.Dislikes_count;
              section.dataset.userVote = newUV;
              syncSection(section);
            });
        }
      } catch (error) {
        console.error('Error updating vote:', error);
        // Revert UI on error
        parent.dataset.userVote = prevUV;
        syncSection(parent);
      }
    });
  });
});