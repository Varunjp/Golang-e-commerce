let croppers = {};

function cropAndSave(index) {
    return new Promise(resolve => {
      const cropper = croppers[index];
      if (!cropper) return resolve(); // No cropper = no image selected
  
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

document.querySelector('form').addEventListener('submit', async function (e) {
  e.preventDefault();

  // Process all croppers that are active (images selected)
  await Promise.all([0, 1, 2].map(i => cropAndSave(i)));

  // Just submit the form — no validation needed for all 3
  this.submit();
});