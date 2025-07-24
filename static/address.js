document.addEventListener('DOMContentLoaded',() =>{
    const form = document.getElementById('addressform')

    if (!form){
      console.error('Form not found!');
      return;
    }

    form.addEventListener('submit',function(e){
      e.preventDefault();

      // clear previous errors
      document.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
      document.querySelectorAll('.invalid-feedback').forEach(el => el.remove());

      const formData = new FormData(form)

      fetch(form.action,{
        method: form.method,
        body: formData,
        headers:  {'X-Requested-With': 'XMLHttpRequest'}
      })
      .then(async response =>{
        const res = await response.json();

        if(response.ok){
          if (res.redirect){
            window.location.href = res.redirect;
          }else{
            location.reload();
          }
        }else if (response.status === 400){
          if (res.errors){
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
        }else{
          alert('Something went wrong. Please try again.');
        }

      })
      .catch(() =>{
        alert('Network error. Please try again.');
      });

    });

  });