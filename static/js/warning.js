document.addEventListener('DOMContentLoaded', function () {
  var postForm = document.querySelector('.new-post form');
  var commentForm = document.querySelector('form[action="/newcomment"]');

  // Post Creation Form Validation
  if (postForm) {
    postForm.addEventListener('submit', function (e) {
      let valid = true;

      // Category validation
      var checked = postForm.querySelectorAll('.category-pills input[type="checkbox"]:checked').length;
      var categoryWarning = document.getElementById('category-warning');
      if (checked === 0) {
        categoryWarning.style.display = 'block';
        valid = false;
      } else {
        categoryWarning.style.display = 'none';
      }

      // Title validation
      var titleInput = postForm.querySelector('input[name="title"]');
      var titleWarning = document.getElementById('title-warning');
      if (!titleInput.value.trim() || titleInput.value.trim().length < 5 || titleInput.value.trim().length > 100) {
        if (!titleWarning) {
          titleWarning = document.createElement('div');
          titleWarning.id = 'title-warning';
          titleWarning.style.color = '#ff4500';
          titleWarning.style.marginBottom = '0.7rem';
          titleInput.parentNode.insertBefore(titleWarning, titleInput.nextSibling);
        }
        titleWarning.textContent = 'Title must be between 5 and 100 characters.';
        titleWarning.style.display = 'block';
        valid = false;
      } else if (titleWarning) {
        titleWarning.style.display = 'none';
      }

      // Content validation
      var contentInput = postForm.querySelector('textarea[name="content"]');
      var contentWarning = document.getElementById('content-warning');
      if (!contentInput.value.trim() || contentInput.value.trim().length < 10 || contentInput.value.trim().length > 1000) {
        if (!contentWarning) {
          contentWarning = document.createElement('div');
          contentWarning.id = 'content-warning';
          contentWarning.style.color = '#ff4500';
          contentWarning.style.marginBottom = '0.7rem';
          contentInput.parentNode.insertBefore(contentWarning, contentInput.nextSibling);
        }
        contentWarning.textContent = 'Description must be between 10 and 1000 characters.';
        contentWarning.style.display = 'block';
        valid = false;
      } else if (contentWarning) {
        contentWarning.style.display = 'none';
      }

      if (!valid) {
        e.preventDefault();
      }
    });
  }

  // Comment Form Validation
  if (commentForm) {
    commentForm.addEventListener('submit', function (e) {
      let valid = true;

      var commentInput = commentForm.querySelector('textarea[name="content"]');
      var commentWarning = document.getElementById('comment-warning');

      if (!commentInput.value.trim() || commentInput.value.trim().length < 5 || commentInput.value.trim().length > 500) {
        if (!commentWarning) {
          commentWarning = document.createElement('div');
          commentWarning.id = 'comment-warning';
          commentWarning.style.color = '#ff4500';
          commentWarning.style.marginTop = '0.5rem';
          commentInput.parentNode.insertBefore(commentWarning, commentInput.nextSibling);
        }
        commentWarning.textContent = 'Comment must be between 5 and 500 characters.';
        commentWarning.style.display = 'block';
        valid = false;
      } else if (commentWarning) {
        commentWarning.style.display = 'none';
      }

      if (!valid) {
        e.preventDefault();
        commentInput.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    });
  }

  // Warning for image size
  document.querySelector('form').addEventListener('submit', function (e) {
    const fileInput = document.getElementById('image');
    if (fileInput) {
      const file = fileInput.files[0];
      if (file && file.size > 20 * 1024 * 1024) { // 20MB
        alert("Image must be 20MB or smaller.");
        e.preventDefault();
      }
    }
  });
});