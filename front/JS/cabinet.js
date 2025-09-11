// Функции для личного кабинета
document.addEventListener('DOMContentLoaded', function() {
    // Загрузка данных пользователя
    fetchUserData();
    
    // Навигация по вкладкам
    setupTabs();
    
    // Модальное окно смены пароля
    setupPasswordModal();
    
    // Обработка форм
    setupForms();
});

async function fetchUserData() {
    try {
        const response = await fetch('/api/me');
        if (!response.ok) {
            window.location.href = '/front/HTML/main.html';
            return;
        }
        
        const userData = await response.json();
        displayUserData(userData);
        
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
        window.location.href = '/front/HTML/main.html';
    }
}

function displayUserData(userData) {
    // Основная информация
    document.getElementById('userInfo').textContent = `Добро пожаловать, ${userData.name || 'Пользователь'}!`;
    document.getElementById('userName').textContent = userData.name || 'Пользователь';
    document.getElementById('userEmail').textContent = userData.email || 'Email не указан';
    document.getElementById('userPhone').textContent = userData.phone || 'Телефон не указан';
    
    // Аватарка с инициалами
    const avatar = document.getElementById('userAvatar');
    if (userData.name) {
        const initials = userData.name.split(' ').map(n => n[0]).join('').toUpperCase();
        avatar.innerHTML = initials;
    }
    
    // Заполнение формы
    if (userData.firstName) document.getElementById('firstName').value = userData.firstName;
    if (userData.lastName) document.getElementById('lastName').value = userData.lastName;
    if (userData.email) document.getElementById('email').value = userData.email;
    if (userData.phone) document.getElementById('phone').value = userData.phone;
}

function setupTabs() {
    const tabs = document.querySelectorAll('.nav-tab');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            // Убираем активный класс у всех вкладок
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Добавляем активный класс текущей вкладке
            tab.classList.add('active');
            const tabId = tab.dataset.tab + '-tab';
            document.getElementById(tabId).classList.add('active');
        });
    });
}

function setupPasswordModal() {
    const modal = document.getElementById('passwordModal');
    const openBtn = document.getElementById('changePasswordBtn');
    const closeBtn = document.getElementById('closePasswordModal');
    const cancelBtn = document.getElementById('cancelPasswordChange');
    const form = document.getElementById('passwordForm');
    
    openBtn.addEventListener('click', () => {
        modal.style.display = 'flex';
    });
    
    closeBtn.addEventListener('click', () => {
        modal.style.display = 'none';
    });
    
    cancelBtn.addEventListener('click', () => {
        modal.style.display = 'none';
    });
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });
    
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        // Здесь будет логика смены пароля
        alert('Пароль успешно изменен!');
        modal.style.display = 'none';
    });
}

function setupForms() {
    const profileForm = document.getElementById('profileForm');
    
    profileForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const formData = {
            firstName: document.getElementById('firstName').value,
            lastName: document.getElementById('lastName').value,
            email: document.getElementById('email').value,
            phone: document.getElementById('phone').value
        };
        
        try {
            const response = await fetch('/api/profile', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });
            
            if (response.ok) {
                alert('Данные успешно сохранены!');
            } else {
                alert('Ошибка сохранения данных');
            }
        } catch (error) {
            console.error('Ошибка:', error);
            alert('Ошибка сохранения данных');
        }
    });
}



// Выход из аккаунта
document.getElementById('logoutBtn').addEventListener('click', async () => {
    try {
        await fetch('/api/logout', { method: 'POST' });
        window.location.href = '/front/HTML/main.html';
    } catch (error) {
        console.error('Ошибка выхода:', error);
    }
});