document.addEventListener('DOMContentLoaded', function() {
  // Инициализация всех функций после загрузки DOM
  initMobileMenu();
  initSmoothScroll();
  initFormValidation();
  initScrollAnimations();
  replaceEmptyLinks();
  initAuthButton();
  initCartLink();
  initCartButtons();
  updateCartCounter();

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
  // Исключаем формы авторизации/регистрации, у них своя логика в auth.js
  const forms = document.querySelectorAll('form:not(.auth-form)');
  
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

// Замена кнопки Войти -> Кабинет при активной сессии
async function initAuthButton(){
  try{
    const r = await fetch('/api/me', { credentials: 'include' });
    if(!r.ok) return; // не авторизован — ничего не меняем
    let nav = document.querySelector('.nav-menu .header-actions');
    if(!nav) nav = document.querySelector('.header-actions');
    if(!nav) return;
    // Если мы уже на странице кабинета, кнопку скрываем
    if (window.location && /^\/cabinet\/?/i.test(window.location.pathname || '')) {
      const existingCab = nav.querySelector('a[href="/cabinet/"]');
      if(existingCab){ existingCab.remove(); }
      const loginAnchor = nav.querySelector('.login-link');
      if(loginAnchor){ loginAnchor.remove(); }
      return;
    }
    // Ищем существующую ссылку входа
    const loginLink = nav.querySelector('.login-link');
    const btn = document.createElement('a');
    btn.className = 'btn-cabinet';
    btn.href = '/cabinet/';
    btn.textContent = 'Кабинет';
    if(loginLink){
      loginLink.replaceWith(btn);
    } else {
      nav.appendChild(btn);
    }
  }catch(e){
    // ignore
  }
}

// Делает ссылку Корзина кликабельной и ведущей на /cart/
function initCartLink(){
  const cart = document.querySelector('.header-actions .cart-link');
  if(cart){
    // Сделаем обычной ссылкой на /cart/
    cart.setAttribute('href', '/cart/');
  }
}

// Обработчики кнопок добавления в корзину на карточках предложений
// Делаем доступной глобально, чтобы вызывать после динамической отрисовки
window.initCartButtons = function initCartButtons(){
  const buttons = document.querySelectorAll('.offer-card .cart-btn');
  if(!buttons || buttons.length===0) return;
  buttons.forEach(btn => {
    btn.addEventListener('click', function(){
      const card = btn.closest('.offer-card');
      if(!card) return;
      const titleEl = card.querySelector('h4');
      const priceEl = card.querySelector('.new-price');
      const imgEl = card.querySelector('img');
      const title = titleEl ? titleEl.textContent.trim() : 'Товар';
      const priceStr = priceEl ? priceEl.textContent.replace(/[^0-9]/g,'') : '0';
      const price = parseInt(priceStr||'0', 10);
      const image = imgEl ? imgEl.getAttribute('src') : '';
      const id = (title + '|' + image).toLowerCase();

      const item = { id, title, price, image, qty: 1 };
      addToCart(item);
      updateCartCounter();
    });
  });
}

function readCart(){
  try{
    const raw = localStorage.getItem('cartItems');
    return raw ? JSON.parse(raw) : [];
  }catch{ return []; }
}

function writeCart(items){
  localStorage.setItem('cartItems', JSON.stringify(items));
}

function addToCart(newItem){
  const items = readCart();
  const idx = items.findIndex(i => i.id === newItem.id);
  if(idx >= 0){
    items[idx].qty = (items[idx].qty||1) + (newItem.qty||1);
  } else {
    items.push(newItem);
  }
  writeCart(items);
}

function getCartCount(){
  const items = readCart();
  return items.reduce((sum, i) => sum + (i.qty||1), 0);
}

function updateCartCounter(){
  const link = document.querySelector('.header-actions .cart-link');
  if(!link) return;
  let badge = link.querySelector('.cart-count-badge');
  if(!badge){
    badge = document.createElement('span');
    badge.className = 'cart-count-badge';
    badge.style.cssText = 'margin-left:4px;background:#e53935;color:#fff;border-radius:10px;padding:0 5px;font-size:10px;line-height:14px;display:inline-block;vertical-align:top;';
    link.appendChild(badge);
  }
  const count = getCartCount();
  if(count>0){
    badge.textContent = count;
    badge.style.display = 'inline-block';
  } else {
    badge.style.display = 'none';
  }
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
    center: [59.820540, 30.370800], // Координаты: Московское шоссе 46Б, 
    zoom: 15
  });

  myMap.behaviors.disable('scrollZoom');

  const myPlacemark = new ymaps.Placemark([59.820540, 30.370800], {
    hintContent: 'Межрегионсталь',
    balloonContent: 'Санкт-Петербург, ул. Московское шоссе 46Б'
  });

  myMap.geoObjects.add(myPlacemark);
}

// Обновленный код для бургер-меню
document.addEventListener('DOMContentLoaded', () => {
  const burgerBtn = document.querySelector('.burger-btn');
  const navMenu = document.querySelector('.nav-menu');
  const body = document.body;

  // Переключение меню
  burgerBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    navMenu.classList.toggle('active');
    body.classList.toggle('menu-open');
    
    // Меняем иконку
    const icon = burgerBtn.querySelector('i');
    if (navMenu.classList.contains('active')) {
      icon.className = 'fas fa-times';
    } else {
      icon.className = 'fas fa-bars';
    }
  });

  // Закрытие меню при клике на ссылку
  const navLinks = document.querySelectorAll('.nav-menu a');
  navLinks.forEach(link => {
    link.addEventListener('click', () => {
      navMenu.classList.remove('active');
      body.classList.remove('menu-open');
      const icon = burgerBtn.querySelector('i');
      icon.className = 'fas fa-bars';
    });
  });

  // Закрытие меню при клике вне его
  document.addEventListener('click', (e) => {
    if (navMenu.classList.contains('active') && 
        !e.target.closest('.nav-menu') && 
        !e.target.closest('.burger-btn')) {
      navMenu.classList.remove('active');
      body.classList.remove('menu-open');
      const icon = burgerBtn.querySelector('i');
      icon.className = 'fas fa-bars';
    }
  });

  // Закрытие меню при нажатии Escape
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && navMenu.classList.contains('active')) {
      navMenu.classList.remove('active');
      body.classList.remove('menu-open');
      const icon = burgerBtn.querySelector('i');
      icon.className = 'fas fa-bars';
    }
  });
});

// Применение ссылок на соцсети глобально (иконки в любом месте страницы)
document.addEventListener('DOMContentLoaded', async () => {
  try{
    const r = await fetch('/api/social', { cache: 'no-store' });
    if(!r.ok) return;
    const s = await r.json();
    const map = { vk: s.vk_link||'', telegram: s.telegram_link||'', whatsapp: s.wp_link||'' };
    const containers = document.querySelectorAll('.social-links');
    containers.forEach(container => {
      const anchors = Array.from(container.querySelectorAll('a'));
      anchors.forEach((a, idx) => {
        const isVK = !!a.querySelector('.fa-vk');
        const isTG = !!a.querySelector('.fa-telegram, .fa-telegram-plane');
        const isWA = !!a.querySelector('.fa-whatsapp');
        let url = '';
        if(isVK) url = map.vk; else if(isTG) url = map.telegram; else if(isWA) url = map.whatsapp;
        // fallback by order VK, TG, WA
        if(!url){ if(idx===0) url = map.vk; else if(idx===1) url = map.telegram; else if(idx===2) url = map.whatsapp; }
        if(url && url !== '#') { a.href = url; a.target = '_blank'; a.rel = 'noopener'; a.style.pointerEvents=''; a.style.opacity=''; a.addEventListener('click', function(ev){ ev.stopPropagation(); try{ window.open(url, '_blank'); }catch(_){} }); }
        else { a.href = '#'; a.removeAttribute('target'); a.removeAttribute('rel'); a.style.pointerEvents='none'; a.style.opacity='0.5'; }
      });
    });
  }catch(e){ /* ignore */ }
});

document.body.classList.add('loaded');
