document.addEventListener('DOMContentLoaded', function() {
  // Инициализация всех функций после загрузки DOM
  initMobileMenu();
  initSmoothScroll();
  initFormValidation();
  initScrollAnimations();
  replaceEmptyLinks();

  // Инициализация Яндекс.Карты
  if (typeof ymaps !== 'undefined') {
    ymaps.ready(initMap).catch(error => {
      console.error('Ошибка загрузки Яндекс.Карт:', error);
    });
  } else {
    console.error('Библиотека Яндекс.Карт не загрузилась');
  }
});

// Мобильное меню
function initMobileMenu() {
  const menuToggle = document.createElement('div');
  menuToggle.className = 'mobile-menu-toggle';
  menuToggle.innerHTML = '<span></span><span></span><span></span>';
  
  const header = document.querySelector('.header-container');
  header.appendChild(menuToggle);
  
  menuToggle.addEventListener('click', function() {
    document.body.classList.toggle('menu-open');
    const navMenu = document.querySelector('.nav-menu');
    navMenu.classList.toggle('active');
  });
}

// Плавная прокрутка
function initSmoothScroll() {
  const links = document.querySelectorAll('a[href^="#"]');
  
  links.forEach(link => {
    link.addEventListener('click', function(e) {
      e.preventDefault();
      
      const targetId = this.getAttribute('href');
      if (targetId === '#') return;
      
      const targetElement = document.querySelector(targetId);
      if (targetElement) {
        window.scrollTo({
          top: targetElement.offsetTop - 80,
          behavior: 'smooth'
        });
      }
    });
  });
}

// Валидация форм
function initFormValidation() {
  const forms = document.querySelectorAll('form');
  
  forms.forEach(form => {
    form.addEventListener('submit', function(e) {
      e.preventDefault();
      
      let isValid = true;
      const inputs = this.querySelectorAll('input[required], textarea[required]');
      
      inputs.forEach(input => {
        if (!input.value.trim()) {
          isValid = false;
          input.classList.add('error');
        } else {
          input.classList.remove('error');
        }
      });
      
      if (isValid) {
        alert('Форма успешно отправлена!');
        this.reset();
      }
    });
  });
}

// Анимации при скролле
function initScrollAnimations() {
  const animatedElements = document.querySelectorAll('.category-card, .advantage-item');
  
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('animate');
      }
    });
  }, { threshold: 0.1 });
  
  animatedElements.forEach(element => {
    observer.observe(element);
  });
}

// Замена пустых ссылок
function replaceEmptyLinks() {
  const links = document.querySelectorAll('a');
  
  links.forEach(link => {
    const href = link.getAttribute('href');
    if (!href || href === 'javascript:void(0)' || href === '') {
      link.setAttribute('href', '#');
    }
  });
}

// Функция для отправки данных форм (заглушка)
function submitForm(formData) {
  console.log('Данные формы:', formData);
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({ success: true, message: 'Форма успешно отправлена' });
    }, 1000);
  });
}

// Инициализация карты
function initMap() {
  console.log('Яндекс.Карты инициализированы');
  const myMap = new ymaps.Map("map", {
    center: [59.987208, 30.445029], // Координаты: Санкт-Петербург, ул. Коммуны, 67
    zoom: 15
  });

  myMap.behaviors.disable('scrollZoom');

  const myPlacemark = new ymaps.Placemark([59.987208, 30.445029], {
    hintContent: 'Сеть металлобаз Металл ДК',
    balloonContent: 'Санкт-Петербург, ул. Коммуны, 67'
  });

  myMap.geoObjects.add(myPlacemark);
}