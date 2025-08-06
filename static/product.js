
document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('productForm');
  if (!form) {
    console.error('Form not found!');
    return;
  }

  let croppers = {};

  // Helper: Convert file directly to base64 (no cropper)
  function convertFileToBase64(file, index) {
    return new Promise(resolve => {
      const reader = new FileReader();
      reader.onload = function (e) {
        document.getElementById(`cropped_image${index}`).value = e.target.result;

        // Preview non-image files as text/pdf icon etc. or clear preview
        const preview = document.getElementById(`preview${index}`);
        preview.src = ""; // No preview for non-image
        resolve();
      };
      reader.readAsDataURL(file);
    });
  }

  // Helper: Crop and convert image using Cropper
  function cropAndSave(index) {
    return new Promise(resolve => {
      const cropper = croppers[index];
      if (!cropper) return resolve();

      const canvas = cropper.getCroppedCanvas();
      if (!canvas) return resolve();

      canvas.toBlob(blob => {
        const reader = new FileReader();
        reader.onloadend = () => {
          document.getElementById(`cropped_image${index}`).value = reader.result;
          resolve();
        };
        reader.readAsDataURL(blob);
      }, 'image/jpeg');
    });
  }

  // Set up file input handler
  document.querySelectorAll('.crop-image').forEach(input => {
    input.addEventListener('change', function () {
      const index = this.dataset.index;
      const file = this.files[0];
      if (!file) return;

      const preview = document.getElementById(`preview${index}`);

      // Check if file is an image
      if (file.type.startsWith('image/')) {
        const reader = new FileReader();
        reader.onload = function (e) {
          preview.src = e.target.result;

          // Destroy previous cropper if any
          if (croppers[index]) {
            croppers[index].destroy();
          }

          // Initialize cropper
          croppers[index] = new Cropper(preview, {
            aspectRatio: 1,
            viewMode: 1,
            autoCropArea: 1,
            responsive: true,
          });
        };
        reader.readAsDataURL(file);
      } else {
        // If not an image, destroy cropper and convert directly
        if (croppers[index]) {
          croppers[index].destroy();
          delete croppers[index];
        }

        convertFileToBase64(file, index);
      }
    });
  });

  // Helper: process all image croppers (only those that exist)
  async function processCroppers() {
    await Promise.all([0, 1, 2].map(i => {
      if (croppers[i]) {
        return cropAndSave(i);
      }
      return Promise.resolve(); // no cropper, already set
    }));
  }

  // Handle form submission
  form.addEventListener('submit', async function (e) {
    e.preventDefault();

    // Clear previous errors
    document.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
    document.querySelectorAll('.invalid-feedback').forEach(el => el.remove());

    // Ensure cropped image base64 is ready
    await processCroppers();

    const formData = new FormData(form);

    fetch(form.action, {
      method: form.method,
      body: formData,
      headers: { 'X-Requested-With': 'XMLHttpRequest' }
    })
    .then(async response => {
      const res = await response.json();

      if (response.ok) {
        if (res.redirect) {
          window.location.href = res.redirect;
        } else {
          location.reload();
        }
      } else if (response.status === 400) {
        if (res.errors) {
          Object.keys(res.errors).forEach(field => {
            const input = form.querySelector(`[name="${field}"]`);
            if (input) {
              input.classList.add('is-invalid');
              const feedback = document.createElement('div');
              feedback.className = 'invalid-feedback';
              feedback.innerText = res.errors[field];
              input.insertAdjacentElement('afterend', feedback);
            }
          });
        }
      } else {
        alert('Something went wrong. Please try again.');
      }
    })
    .catch(() => {
      alert('Network error. Please try again.');
    });
  });
});
