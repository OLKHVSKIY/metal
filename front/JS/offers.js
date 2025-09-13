document.addEventListener('DOMContentLoaded', async () => {
  const grid = document.querySelector('.offers-grid');
  const toggleBtn = document.querySelector('.toggle-btn');
  if (!grid) return;

  // Сначала скройте статический контент
  grid.innerHTML = '';

  try {
    const res = await fetch('/api/featured', { cache: 'no-store', headers: { 'Cache-Control': 'no-cache' } });
    const items = await res.json();
    if (Array.isArray(items) && items.length > 0) {
      grid.innerHTML = items.map((p, idx) => {
        const hiddenCls = idx >= 8 ? ' hidden' : '';
        
        // Генерируем цены и экономию в вашем оригинальном формате
        const oldPrice = p.old_price || Math.round(p.price * 1.25);
        const savings = oldPrice - p.price;
        
        const priceHtml = p.price && p.price > 0 ? `
          <span class="old-price">${oldPrice} ₽</span>
          <span class="new-price">${p.price} ₽</span>
          <span class="savings">Экономия: ${savings} ₽</span>
        ` : '';
        
        const typeSlug = (p.type_slug||'').toString();
        const name = (p.name||'').toString();
        const size = (p.size||'').toString();
        const href = typeSlug && name ? ('/catalog/'+encodeURIComponent(typeSlug)+'/'+encodeURIComponent(name)+'/'+encodeURIComponent(size)+'/') : '#';
        
        return `
          <a class="offer-card${hiddenCls}" href="${href}" style="text-decoration:none;color:inherit;">
            <img src="${p.img||''}" alt="${p.name||'Товар'}">
            <span class="availability">${p.in_stock ? 'Есть в наличии' : 'Под заказ'}</span>
            <h4>${p.name||'Товар'}</h4>
            <div class="price-container">${priceHtml}</div>
            <div class="button-container">
              <button class="cart-btn" type="button"><i class="fas fa-cart-plus"></i></button>
              <button class="order-btn" type="button">Заказать</button>
            </div>
          </a>
        `;
      }).join('');
      
// Принудительно показываем цены в вашем стиле
setTimeout(() => {
  const priceContainers = grid.querySelectorAll('.price-container');
  priceContainers.forEach(container => {
    container.style.display = 'flex';
    container.style.flexDirection = 'column';
    container.style.alignItems = 'center';
    container.style.gap = '4px';
    container.style.marginBottom = '10px';
  });

  const oldPrices = grid.querySelectorAll('.old-price');
  oldPrices.forEach(el => {
    el.style.color = '#888';
    el.style.textDecoration = 'line-through';
    el.style.fontWeight = '400';
    el.style.fontSize = '14px';
    el.style.margin = '0';
    el.style.padding = '0';
  });

  const newPrices = grid.querySelectorAll('.new-price');
  newPrices.forEach(el => {
    el.style.color = '#d32f2f';
    el.style.fontWeight = 'bold';
    el.style.fontSize = '18px';
    el.style.margin = '0';
    el.style.padding = '0';
  });

  const savings = grid.querySelectorAll('.savings');
  savings.forEach(el => {
    el.style.color = '#666';
    el.style.fontWeight = '400';
    el.style.fontSize = '13px';
    el.style.marginTop = '5px';
    el.style.textAlign = 'center';
    el.style.width = '100%';
  });
}, 100);
      
      // Повесим обработчики на кнопки корзины
      const attach = () => {
        const buttons = grid.querySelectorAll('.offer-card .cart-btn');
        buttons.forEach(btn => {
          btn.addEventListener('click', function(ev){
            ev.preventDefault(); 
            ev.stopPropagation();
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
            try { 
              if (typeof addToCart === 'function') addToCart(item); 
            } catch(e) {}
            try { 
              if (typeof updateCartCounter === 'function') updateCartCounter(); 
            } catch(e) {}
          });
        });
      };
      attach();
    }
  } catch (e) {
    console.warn('Failed to load featured products', e);
  }

  const refreshToggle = () => {
    if (!toggleBtn) return;
    const allCards = Array.from(grid.querySelectorAll('.offer-card'));
    toggleBtn.style.display = allCards.length > 8 ? 'inline-block' : 'none';
    
    // Центрируем кнопку
    toggleBtn.style.margin = '20px auto 0';
    toggleBtn.style.display = 'flex';
    toggleBtn.style.justifyContent = 'center';
    
    toggleBtn.setAttribute('data-state', 'show');
    toggleBtn.textContent = 'Показать ещё';
    
    toggleBtn.onclick = () => {
      const isShow = toggleBtn.getAttribute('data-state') === 'show';
      const cards = Array.from(grid.querySelectorAll('.offer-card'));
      const bestOffersSection = document.querySelector('.best-offers');
      
      if (isShow) {
        // Показываем все карточки
        cards.forEach(card => card.classList.remove('hidden'));
        toggleBtn.textContent = 'Скрыть';
        toggleBtn.setAttribute('data-state', 'hide');
      } else {
        // Скрываем дополнительные карточки
        cards.forEach((card, idx) => {
          if (idx >= 8) {
            card.classList.add('hidden');
          }
        });
        toggleBtn.textContent = 'Показать ещё';
        toggleBtn.setAttribute('data-state', 'show');
        
        // Плавная прокрутка к началу секции
        if (bestOffersSection) {
          setTimeout(() => {
            window.scrollTo({
              top: bestOffersSection.offsetTop - 100,
              behavior: 'smooth'
            });
          }, 50);
        }
      }
    };
  };
  
  // Запускаем после загрузки карточек
  setTimeout(refreshToggle, 200);
});

// Добавляем CSS для гарантии отображения в вашем стиле
const style = document.createElement('style');
style.textContent = `
  .best-offers .container {
    text-align: center;
  }
  
  .toggle-btn {
    display: inline-flex !important;
    align-items: center;
    justify-content: center;
    margin: 20px auto 0 !important;
    background-color: var(--primary-color) !important;
    color: var(--text-light) !important;
    padding: 12px 25px !important;
    border-radius: var(--border-radius) !important;
    border: none !important;
    font-weight: bold !important;
    font-size: 14px !important;
    cursor: pointer !important;
    transition: var(--transition) !important;
  }
  
  .toggle-btn:hover {
    background-color: #b71c1c !important;
    transform: translateY(-2px) !important;
  }
  
  /* СТАРАЯ ЦЕНА - серая, перечеркнутая */
  .offer-card .old-price {
    color: #888 !important;
    text-decoration: line-through !important;
    font-weight: 400 !important;
    display: block !important;
    font-size: 14px !important;
    margin: 0 !important;
    padding: 0 !important;
  }
  
  /* НОВАЯ ЦЕНА - красная, жирная */
  .offer-card .new-price {
    color: #d32f2f !important;
    font-weight: bold !important;
    display: block !important;
    font-size: 18px !important;
    margin: 0 !important;
    padding: 0 !important;
  }
  
  /* ЭКОНОМИЯ - серая, под ценами по центру */
  .offer-card .savings {
    color: #666 !important;
    font-weight: 400 !important;
    display: block !important;
    width: 100% !important;
    text-align: center !important;
    margin-top: 5px !important;
    font-size: 13px !important;
  }
  
  .price-container {
    display: flex !important;
    flex-direction: column !important;
    justify-content: center !important;
    align-items: center !important;
    gap: 4px !important;
    margin-bottom: 10px !important;
  }
  
  .best-offers {
    scroll-margin-top: 80px;
  }
  
  html {
    scroll-behavior: smooth;
  }
  
  /* Ваши оригинальные стили для карточек */
  .offer-card {
    background-color: var(--light-color) !important;
    border-radius: var(--border-radius) !important;
    padding: 15px !important;
    text-align: center !important;
    transition: var(--transition) !important;
    border: 1px solid #ddd !important;
  }
  
  .offer-card:hover {
    transform: translateY(-5px) !important;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15) !important;
  }
  
  .offer-card img {
    width: 100% !important;
    max-width: 220px !important;
    height: 220px !important;
    object-fit: contain !important;
    margin-bottom: 10px !important;
  }
  
  .offer-card .availability {
    color: var(--green-color) !important;
    font-size: 13px !important;
    font-weight: 500 !important;
    margin-bottom: 8px !important;
    display: block !important;
    width: 60% !important;
    text-align: center !important;
    margin-left: auto !important;
    margin-right: auto !important;
  }
  
  .offer-card h4 {
    font-size: 15px !important;
    margin-bottom: 8px !important;
    color: var(--text-color) !important;
  }
  
  .button-container {
    display: flex !important;
    justify-content: center !important;
    gap: 10px !important;
  }
  
  .cart-btn {
    width: 45px !important;
    height: 35px !important;
    display: flex !important;
    justify-content: center !important;
    align-items: center !important;
    background-color: var(--gray-color) !important;
    color: var(--text-color) !important;
    border: none !important;
    border-radius: var(--border-radius) !important;
    cursor: pointer !important;
    font-size: 16px !important;
    transition: var(--transition) !important;
  }
  
  .cart-btn:hover {
    background-color: var(--primary-color) !important;
    color: var(--text-light) !important;
    transform: translateY(-2px) !important;
  }
  
  .order-btn {
    flex: 1 !important;
    padding: 8px !important;
    background-color: var(--primary-color) !important;
    color: var(--text-light) !important;
    border: none !important;
    border-radius: var(--border-radius) !important;
    cursor: pointer !important;
    font-size: 14px !important;
    font-weight: bold !important;
    transition: var(--transition) !important;
  }
  
  .order-btn:hover {
    background-color: #b71c1c !important;
    transform: translateY(-2px) !important;
  }
  
  .offer-card.hidden {
    display: none !important;
  }
`;
document.head.appendChild(style);
