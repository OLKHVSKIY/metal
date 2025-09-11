// Функции для работы с корзиной
function readCart() {
    try {
        const raw = localStorage.getItem('cartItems');
        return raw ? JSON.parse(raw) : [];
    } catch {
        return [];
    }
}

function saveCart(items) {
    localStorage.setItem('cartItems', JSON.stringify(items));
}

function updateQuantity(id, newQty) {
    const items = readCart();
    const item = items.find(i => i.id === id);
    if (item) {
        item.qty = Math.max(1, newQty);
        saveCart(items);
        renderCart();
    }
}

function removeFromCart(id) {
    const items = readCart().filter(i => i.id !== id);
    saveCart(items);
    renderCart();
    showNotification('Товар удален из корзины');
}

function showNotification(message) {
    const notification = document.createElement('div');
    notification.className = 'cart-notification';
    notification.innerHTML = `
        <i class="fas fa-check-circle" style="color: var(--primary-color); margin-right: 10px;"></i>
        ${message}
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

function renderCart() {
    const items = readCart();
    const list = document.getElementById('cartList');
    const summary = document.getElementById('cartSummary');
    
    if (items.length === 0) {
        list.innerHTML = `
            <div class="cart-empty">
                <div class="cart-empty-icon">
                    <i class="fas fa-shopping-cart"></i>
                </div>
                <p class="cart-empty-text">Ваша корзина пока пуста</p>
                <a href="/front/HTML/catalog.html" class="cart-empty-btn">
                    <i class="fas fa-store"></i>
                    Перейти к покупкам
                </a>
            </div>
        `;
        summary.innerHTML = '';
        return;
    }
    
    let total = 0;
    list.innerHTML = items.map(item => {
        const itemTotal = (item.price || 0) * (item.qty || 1);
        total += itemTotal;
        
        return `
            <div class="cart-item" data-id="${item.id}">
                <img src="${item.image}" alt="${item.title}" class="cart-item-image">
                <div class="cart-item-info">
                    <h3 class="cart-item-title">${item.title}</h3>
                    <div class="cart-item-price">${item.price} ₽ / шт</div>
                    <div class="cart-item-meta">Артикул: ${item.id}</div>
                    <div class="quantity-controls">
                        <button class="quantity-btn" onclick="updateQuantity(${item.id}, ${(item.qty || 1) - 1})">-</button>
                        <input type="number" class="quantity-input" value="${item.qty || 1}" 
                               min="1" onchange="updateQuantity(${item.id}, parseInt(this.value))">
                        <button class="quantity-btn" onclick="updateQuantity(${item.id}, ${(item.qty || 1) + 1})">+</button>
                    </div>
                </div>
                <div class="cart-item-controls">
                    <div class="cart-item-total">${itemTotal} ₽</div>
                    <button class="cart-item-remove" onclick="removeFromCart(${item.id})">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');
    
    summary.innerHTML = `
        <h3 class="summary-title">Итог заказа</h3>
        <div class="summary-row">
            <span class="summary-label">Товары (${items.length})</span>
            <span class="summary-value">${total} ₽</span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Доставка</span>
            <span class="summary-value">Бесплатно</span>
        </div>
        <div class="summary-row">
            <span class="summary-label">Итого</span>
            <span class="summary-value">${total} ₽</span>
        </div>
        <button class="checkout-btn" onclick="checkout()">
            <i class="fas fa-credit-card"></i>
            Оформить заказ
        </button>
    `;
}

function checkout() {
    alert('Функция оформления заказа будет реализована в ближайшее время!');
}

// Инициализация корзины при загрузке
document.addEventListener('DOMContentLoaded', function() {
    renderCart();
});