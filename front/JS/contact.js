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
});