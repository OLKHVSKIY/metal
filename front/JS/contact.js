document.addEventListener('DOMContentLoaded', () => {
  const feedbackBtn = document.querySelector('.feedback-btn');
  const modal = document.querySelector('#feedbackModal');
  const closeBtn = document.querySelector('.close-btn');
  const feedbackForm = document.querySelector('.feedback-form');
  const fileInput = document.querySelector('#file');
  const fileUpload = document.querySelector('.file-upload');
  const fileList = document.querySelector('.file-list');
  let selectedFiles = [];

  const updateFileList = () => {
    fileList.innerHTML = '';
    selectedFiles.forEach((file, index) => {
      const fileItem = document.createElement('div');
      fileItem.className = 'file-item';
      fileItem.innerHTML = `
        <i class="fas fa-file"></i>
        <span>${file.name}</span>
        <i class="fas fa-times" data-index="${index}"></i>
      `;
      fileList.appendChild(fileItem);
    });

    const removeButtons = fileList.querySelectorAll('.fa-times');
    removeButtons.forEach(button => {
      button.addEventListener('click', () => {
        const index = parseInt(button.getAttribute('data-index'));
        selectedFiles.splice(index, 1);
        updateFileList();
      });
    });
  };

  fileUpload.addEventListener('click', () => {
    fileInput.click();
  });

  fileInput.addEventListener('change', () => {
    const newFiles = Array.from(fileInput.files);
    const totalFiles = selectedFiles.length + newFiles.length;

    if (totalFiles > 5) {
      alert('Можно прикрепить не более 5 файлов.');
      fileInput.value = '';
      return;
    }

    selectedFiles = [...selectedFiles, ...newFiles].slice(0, 5);
    updateFileList();
    fileInput.value = '';
  });

  feedbackBtn.addEventListener('click', () => {
    modal.style.display = 'flex';
    setTimeout(() => {
      modal.classList.add('open');
    }, 10);
  });

  closeBtn.addEventListener('click', () => {
    modal.classList.remove('open');
    setTimeout(() => {
      modal.style.display = 'none';
    }, 300);
  });

  window.addEventListener('click', (event) => {
    if (event.target === modal) {
      modal.classList.remove('open');
      setTimeout(() => {
        modal.style.display = 'none';
      }, 300);
    }
  });

  feedbackForm.addEventListener('submit', (event) => {
    event.preventDefault();
    alert('Отзыв отправлен! Спасибо за ваш отклик.');
    modal.classList.remove('open');
    setTimeout(() => {
      modal.style.display = 'none';
    }, 300);
    feedbackForm.reset();
    selectedFiles = [];
    fileList.innerHTML = '';
  });

  // Load social links from backend and apply to contact page icons and footer
  (async function applySocialLinks(){
    try{
      const r = await fetch('/api/social', { cache: 'no-store' });
      if(!r.ok) return;
      const s = await r.json();
      const map = {
        vk: s.vk_link || '',
        telegram: s.telegram_link || '',
        whatsapp: s.wp_link || ''
      };
      function wire(container){
        if(!container) return;
        const a = container.querySelectorAll('a');
        a.forEach(el=>{
          const isVK = !!el.querySelector('.fa-vk');
          const isTG = !!el.querySelector('.fa-telegram, .fa-telegram-plane');
          const isWA = !!el.querySelector('.fa-whatsapp');
          let url = '';
          if(isVK) url = map.vk; else if(isTG) url = map.telegram; else if(isWA) url = map.whatsapp;
          if(url){
            el.href = url; el.target = '_blank'; el.rel = 'noopener';
            el.style.opacity = '';
            el.style.pointerEvents = '';
            el.addEventListener('click', function(e){ if(!el.href || el.href==='#'){ e.preventDefault(); } });
          } else {
            // нет ссылки — отключим клик визуально
            el.href = '#'; el.removeAttribute('target'); el.removeAttribute('rel');
            el.style.opacity = '0.5'; el.style.pointerEvents = 'none';
          }
        });
      }
      // Top block
      wire(document.querySelector('.contacts .social-links'));
      // Footer block
      wire(document.querySelector('footer .social-links'));
    }catch(e){ /* ignore */ }
  })();
});