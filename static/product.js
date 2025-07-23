document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('productForm');
  if (!form) {
    console.error('Form not found!');
    return;
  }

  let croppers = {};

  // helper: process a single cropper
  function cropAndSave(index) {
    return new Promise(resolve => {
      const cropper = croppers[index];
      if (!cropper) return resolve();

      const canvas = cropper.getCroppedCanvas();
      if (!canvas) return resolve();

      canvas.toBlob(blob => {
        const reader = new FileReader();
        reader.onloadend = () => {
          console.log(`Base64 data for image ${index}:`, reader.result.slice(0, 100));
          document.getElementById(`cropped_image${index}`).value = reader.result;
          resolve();
        };
        reader.readAsDataURL(blob);
      }, 'image/jpeg');
    });
  }

  // helper: process all croppers
  async function processCroppers() {
    await Promise.all([0, 1, 2].map(i => cropAndSave(i)));
  }

  // setup cropper for each image input
  document.querySelectorAll('.crop-image').forEach(input => {
    input.addEventListener('change', function () {
      const index = this.dataset.index;
      const file = this.files[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = function (e) {
        const img = document.getElementById(`preview${index}`);
        img.src = e.target.result;

        if (croppers[index]) croppers[index].destroy();

        croppers[index] = new Cropper(img, {
          aspectRatio: 1,
          viewMode: 1,
          autoCropArea: 1,
          responsive: true,
        });
      };
      reader.readAsDataURL(file);
    });
  });

  // handle form submission
  form.addEventListener('submit', async function (e) {
    e.preventDefault();
 

    // clear previous errors
    document.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
    document.querySelectorAll('.invalid-feedback').forEach(el => el.remove());

    // process croppers before submission
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
