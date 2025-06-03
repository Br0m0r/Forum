document.getElementById('image').addEventListener('change', function(event) {
  const preview = document.getElementById('preview');
  const file = event.target.files[0];

  if (file) {
    const validTypes = ['image/jpeg', 'image/png', 'image/gif'];
    if (!validTypes.includes(file.type)) {
      alert('Only JPEG, PNG, and GIF files are allowed.');
      event.target.value = '';
      preview.style.display = 'none';
      return;
    }

    if (file.size > 20 * 1024 * 1024) {
      alert('Image must be 20MB or smaller.');
      event.target.value = '';
      preview.style.display = 'none';
      return;
    }

    const reader = new FileReader();
    reader.onload = function(e) {
      preview.src = e.target.result;
      preview.style.display = 'block';
    };
    reader.readAsDataURL(file);
  } else {
    preview.style.display = 'none';
  }
});