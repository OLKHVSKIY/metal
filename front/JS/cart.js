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
    const item = items.find(i => String(i.id) === String(id));
    if (item) {
        item.qty = Math.max(1, newQty);
        saveCart(items);
        renderCart();
        if (typeof updateCartCounter === 'function') { try { updateCartCounter(); } catch{} }
    }
}

function removeFromCart(id) {
    const items = readCart().filter(i => String(i.id) !== String(id));
    saveCart(items);
    renderCart();
    showNotification('Товар удален из корзины');
    if (typeof updateCartCounter === 'function') { try { updateCartCounter(); } catch{} }
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
        const qid = String(item.id||'').replace(/'/g, "\\'");
        
        return `
            <div class="cart-item" data-id="${item.id}">
                <img src="${item.image}" alt="${item.title}" class="cart-item-image">
                <div class="cart-item-info">
                    <h3 class="cart-item-title">${item.title}</h3>
                    <div class="cart-item-price">${item.price} ₽ / шт</div>
                    <div class="cart-item-meta">Артикул: ${item.id}</div>
                    <div class="quantity-controls">
                        <button class="quantity-btn" onclick="updateQuantity('${qid}', ${(item.qty || 1) - 1})">-</button>
                        <input type="number" class="quantity-input" value="${item.qty || 1}" 
                               min="1" onchange="updateQuantity('${qid}', parseInt(this.value)||1)">
                        <button class="quantity-btn" onclick="updateQuantity('${qid}', ${(item.qty || 1) + 1})">+</button>
                    </div>
                </div>
                <div class="cart-item-controls">
                    <div class="cart-item-total">${itemTotal} ₽</div>
                    <button class="cart-item-remove" onclick="removeFromCart('${qid}')">
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
    const items = readCart();
    if (!items || items.length === 0) { alert('Корзина пуста'); return; }
    function showModal(html){
        const wrap = document.createElement('div');
        wrap.className = 'ui-modal-overlay';
        const box = document.createElement('div'); box.className = 'ui-modal'; box.innerHTML = html; wrap.appendChild(box);
        document.body.appendChild(wrap);
        wrap.addEventListener('click', function(e){ if(e.target===wrap){ document.body.removeChild(wrap); }});
        return { close: function(){ try{ document.body.removeChild(wrap);}catch(_){ } } };
    }
    async function isAuth(){ try{ const r=await fetch('/api/me', { credentials:'include' }); return r.ok; }catch{ return false; } }
    const total = items.reduce((s,i)=> s + (Number(i.price)||0)*(Number(i.qty)||1), 0);
    const summaryLines = items.slice(0,5).map(i=>`${i.title} — ${i.qty} шт.`).join('<br>');
    if (!window.__cart_oneclick_busy) window.__cart_oneclick_busy = false;
    if (window.__cart_oneclick_busy) return; // prevent double click
    (async function(){
        if(!(await isAuth())){
            const m = showModal('<div class="modal-title">Быстрый заказ</div>'
                +'<p>Оставьте номер телефона, и наш менеджер свяжется с вами в течение 5 минут.</p>'
                +'<div class="modal-summary">'+ summaryLines + (items.length>5?'…':'') +'</div>'
                +'<div class="modal-total">Итого: '+ Math.round(total).toLocaleString('ru-RU') +' ₽</div>'
                +'<input id="cartOneClickPhone" class="modal-input" placeholder="+7 (___) ___-__-__"/>'
                +'<div class="modal-actions"><button id="cocCancel" class="btn secondary">Отмена</button><button id="cocSubmit" class="btn">Подтвердить</button></div>');
            document.getElementById('cocCancel').onclick = function(){ m.close(); };
            document.getElementById('cocSubmit').onclick = async function(){
                const phone = (document.getElementById('cartOneClickPhone').value||'').trim(); if(!phone){ alert('Укажите номер телефона'); return; }
                try{
                    window.__cart_oneclick_busy = true;
                    const resp = await fetch('/api/item-order/batch', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ items: items.map(it=>({ item_id: (it.id||''), title: it.title||'', qty: it.qty||1, price: it.price||0 })), phone }) });
                    if(!resp.ok){ const t=await resp.text(); throw new Error(t||'Ошибка оформления'); }
                    // clear server cart (best-effort)
                    try{ await fetch('/api/cart?all=1', { method:'DELETE' }); }catch(_){ }
                    // clear local cart
                    saveCart([]); renderCart(); if (typeof updateCartCounter==='function'){ try{ updateCartCounter(); }catch(_){} }
                    m.close(); alert('Спасибо! В течение 5 минут вам позвонит менеджер.');
                } catch(err){ alert('Ошибка: '+ (err&&err.message?err.message:String(err))); }
                finally{ window.__cart_oneclick_busy = false; }
            };
        } else {
            const m2 = showModal('<div class="modal-title">Подтверждение заказа</div>'
                +'<div class="modal-summary">'+ summaryLines + (items.length>5?'…':'') +'</div>'
                +'<div class="modal-total">Итого: '+ Math.round(total).toLocaleString('ru-RU') +' ₽</div>'
                +'<div class="modal-actions"><button id="cocCancel2" class="btn secondary">Отмена</button><button id="cocSubmit2" class="btn">Подтвердить</button></div>'
                +'<div class="modal-note">В течение 5 минут вам позвонит наш менеджер, чтобы уточнить детали.</div>');
            document.getElementById('cocCancel2').onclick = function(){ m2.close(); };
            document.getElementById('cocSubmit2').onclick = async function(){
                try{
                    window.__cart_oneclick_busy = true;
                    const resp = await fetch('/api/item-order/batch', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ items: items.map(it=>({ item_id: (it.id||''), title: it.title||'', qty: it.qty||1, price: it.price||0 })), phone: '' }) });
                    if(!resp.ok){ const t=await resp.text(); throw new Error(t||'Ошибка оформления'); }
                    // clear server cart and local cart after successful order
                    try{ await fetch('/api/cart?all=1', { method:'DELETE' }); }catch(_){ }
                    saveCart([]); renderCart(); if (typeof updateCartCounter==='function'){ try{ updateCartCounter(); }catch(_){} }
                    m2.close(); alert('Заказ оформлен! С вами свяжется менеджер.');
                } catch(err){ alert('Ошибка: '+ (err&&err.message?err.message:String(err))); }
                finally{ window.__cart_oneclick_busy = false; }
            };
        }
    })();
}

// Инициализация корзины при загрузке
document.addEventListener('DOMContentLoaded', function() {
    renderCart();
});