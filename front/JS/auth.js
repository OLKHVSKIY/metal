document.addEventListener('DOMContentLoaded', () => {
    const authModal = document.querySelector('#authModal');
    const authCloseBtn = document.querySelector('.auth-close-btn');
    const loginLink = document.querySelector('.login-link');
    const authTabs = document.querySelectorAll('.auth-tab');
    const loginForm = document.querySelector('#loginForm');
    const registerForm = document.querySelector('#registerForm');
    const togglePasswordButtons = document.querySelectorAll('.toggle-password');
    const forgotPasswordLink = document.querySelector('.forgot-password');
    const loginEmailInput = document.getElementById('login-email');
    const registerEmailInput = document.getElementById('register-email');
    const registerPhoneInput = document.getElementById('register-phone');

    // Функция проверки email
    function isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    // Автоматическое определение типа ввода (email или телефон)
    loginEmailInput.addEventListener('input', function(e) {
        const value = e.target.value;
        
        // Если ввод начинается с цифры, предполагаем что это телефон
        if (/^\d/.test(value)) {
            // Применяем маску для телефона
            const phoneValue = value.replace(/\D/g, '');
            let formattedValue = phoneValue;
            
            if (phoneValue.startsWith('7') || phoneValue.startsWith('8')) {
                formattedValue = phoneValue.substring(1);
            }
            
            if (formattedValue.length > 0) {
                formattedValue = '+7 (' + formattedValue;
                if (formattedValue.length > 7) formattedValue = formattedValue.substring(0, 7) + ') ' + formattedValue.substring(7);
                if (formattedValue.length > 12) formattedValue = formattedValue.substring(0, 12) + '-' + formattedValue.substring(12);
                if (formattedValue.length > 15) formattedValue = formattedValue.substring(0, 15) + '-' + formattedValue.substring(15);
                
                e.target.value = formattedValue;
            }
        }
    });

    // Валидация email при вводе (регистрация)
    registerEmailInput.addEventListener('blur', function() {
        if (this.value && !isValidEmail(this.value)) {
            this.style.borderColor = 'red';
            showError(this, 'Введите корректный email адрес');
        } else {
            this.style.borderColor = '#e9ecef';
            hideError(this);
        }
    });

    // Маска для телефона в форме регистрации
    if (registerPhoneInput) {
        registerPhoneInput.addEventListener('input', function(e) {
            let value = e.target.value.replace(/\D/g, '');
            if (value.startsWith('7') || value.startsWith('8')) {
                value = value.substring(1);
            }
            if (value.length > 0) {
                value = '+7 (' + value;
                if (value.length > 7) value = value.substring(0, 7) + ') ' + value.substring(7);
                if (value.length > 12) value = value.substring(0, 12) + '-' + value.substring(12);
                if (value.length > 15) value = value.substring(0, 15) + '-' + value.substring(15);
            }
            e.target.value = value;
        });
    }

    // Функция показа ошибки
    function showError(input, message) {
        hideError(input);
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.style.color = 'red';
        errorDiv.style.fontSize = '12px';
        errorDiv.style.marginTop = '5px';
        errorDiv.textContent = message;
        input.parentNode.appendChild(errorDiv);
    }

    // Функция скрытия ошибки
    function hideError(input) {
        const errorDiv = input.parentNode.querySelector('.error-message');
        if (errorDiv) {
            errorDiv.remove();
        }
    }

    // Открытие модального окна при клике на "Войти"
    loginLink.addEventListener('click', (e) => {
        e.preventDefault();
        authModal.style.display = 'flex';
        setTimeout(() => {
            authModal.classList.add('open');
        }, 10);
    });

    // Закрытие модального окна
    authCloseBtn.addEventListener('click', () => {
        authModal.classList.remove('open');
        setTimeout(() => {
            authModal.style.display = 'none';
            resetForms();
        }, 300);
    });

    // Закрытие модального окна при клике вне его
    window.addEventListener('click', (event) => {
        if (event.target === authModal) {
            authModal.classList.remove('open');
            setTimeout(() => {
                authModal.style.display = 'none';
                resetForms();
            }, 300);
        }
    });

    // Переключение между вкладками Вход/Регистрация
    authTabs.forEach(tab => {
        tab.addEventListener('click', () => {
            authTabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            
            if (tab.dataset.tab === 'login') {
                loginForm.style.display = 'flex';
                registerForm.style.display = 'none';
            } else {
                loginForm.style.display = 'none';
                registerForm.style.display = 'flex';
            }
        });
    });

    // Переключение видимости пароля
    togglePasswordButtons.forEach(button => {
        button.addEventListener('click', () => {
            const input = button.closest('.password-group').querySelector('input[type="password"], input[type="text"]');
            const icon = button.querySelector('i');
            
            if (input.type === 'password') {
                input.type = 'text';
                icon.className = 'fas fa-eye-slash';
            } else {
                input.type = 'password';
                icon.className = 'fas fa-eye';
            }
        });
    });

    // Обработка "Забыли пароль?"
    forgotPasswordLink.addEventListener('click', (e) => {
        e.preventDefault();
        alert('Функция восстановления пароля будет доступна в ближайшее время!');
    });

    // Обработка формы регистрации
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        // Валидация формы
        const email = document.getElementById('register-email').value;
        const phone = document.getElementById('register-phone').value;
        const password = document.getElementById('register-password').value;
        
        if (!email || !phone || !password) {
            alert('Пожалуйста, заполните все поля');
            return;
        }
        
        // Проверка формата email
        if (!isValidEmail(email)) {
            alert('Пожалуйста, введите корректный email адрес');
            registerEmailInput.style.borderColor = 'red';
            showError(registerEmailInput, 'Введите корректный email адрес');
            return;
        }
        
        // Проверка формата телефона
        const phoneRegex = /^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$/;
        if (!phoneRegex.test(phone)) {
            alert('Пожалуйста, введите корректный номер телефона');
            return;
        }
        
        try {
            const res = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, phone, password })
            });
            if(!res.ok){
                const text = await res.text();
                alert('Ошибка регистрации: ' + text);
                return;
            }
            window.location.href = '/cabinet/';
        } catch(err){
            alert('Ошибка сети: ' + err.message);
        }
    });

    // Обработка формы входа
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const identifier = document.getElementById('login-email').value;
        const password = document.getElementById('login-password').value;
        
        if (!identifier || !password) {
            alert('Пожалуйста, заполните все поля');
            return;
        }
        
        // Определяем, что ввел пользователь - email или телефон
        let authMethod = 'email';
        let credentials = {};
        
        // Проверяем, является ли ввод телефоном
        const phoneRegex = /^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$/;
        const isPhone = phoneRegex.test(identifier);
        
        if (isPhone) {
            // Вход по телефону
            authMethod = 'phone';
            credentials.phone = identifier;
            credentials.password = password;
            
            // Проверяем, есть ли @ в номере телефона (не должно быть)
            if (identifier.includes('@')) {
                alert('Номер телефона не может содержать символ @');
                return;
            }
        } else {
            // Вход по email
            authMethod = 'email';
            
            // Проверяем наличие @ в email
            if (!identifier.includes('@')) {
                alert('Email должен содержать символ @');
                loginEmailInput.style.borderColor = 'red';
                showError(loginEmailInput, 'Email должен содержать символ @');
                return;
            }
            
            // Проверяем формат email
            if (!isValidEmail(identifier)) {
                alert('Пожалуйста, введите корректный email адрес');
                loginEmailInput.style.borderColor = 'red';
                showError(loginEmailInput, 'Введите корректный email адрес');
                return;
            }
            
            credentials.email = identifier;
            credentials.password = password;
        }
        
        try{
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(credentials)
            });
            if(!res.ok){
                const text = await res.text();
                alert('Ошибка входа: ' + text);
                return;
            }
            if (document.getElementById('remember-me').checked) {
                localStorage.setItem('rememberedUser', JSON.stringify(credentials));
            }
            window.location.href = '/cabinet/';
        } catch(err){
            alert('Ошибка сети: ' + err.message);
        }
    });

    // Функция сброса форм
    function resetForms() {
        loginForm.reset();
        registerForm.reset();
        // Сбрасываем стили ошибок
        const errorMessages = document.querySelectorAll('.error-message');
        errorMessages.forEach(error => error.remove());
        loginEmailInput.style.borderColor = '#e9ecef';
        registerEmailInput.style.borderColor = '#e9ecef';
    }

    // Проверяем сохраненные данные при загрузке
    const rememberedUser = localStorage.getItem('rememberedUser');
    if (rememberedUser) {
        const userData = JSON.parse(rememberedUser);
        document.getElementById('login-email').value = userData.email || userData.phone || '';
        document.getElementById('remember-me').checked = true;
    }
});