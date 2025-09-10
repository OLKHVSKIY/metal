document.addEventListener('DOMContentLoaded', () => {
    const authModal = document.querySelector('#authModal');
    const authCloseBtn = document.querySelector('.auth-close-btn');
    const loginLink = document.querySelector('.login-link');
    const authTabs = document.querySelectorAll('.auth-tab');
    const loginForm = document.querySelector('#loginForm');
    const registerForm = document.querySelector('#registerForm');
    const togglePasswordButtons = document.querySelectorAll('.toggle-password');
    const forgotPasswordLink = document.querySelector('.forgot-password');
    const loginIdentifierInput = document.getElementById('login-identifier');
    const registerPhoneInput = document.getElementById('register-phone');

    // Автоматическое определение типа ввода (логин или телефон)
    loginIdentifierInput.addEventListener('input', function(e) {
        const value = e.target.value.replace(/\D/g, '');
        
        // Если ввод начинается с цифры, предполагаем что это телефон
        if (/^\d/.test(e.target.value)) {
            // Применяем маску для телефона
            let phoneValue = value;
            if (phoneValue.startsWith('7') || phoneValue.startsWith('8')) {
                phoneValue = phoneValue.substring(1);
            }
            
            if (phoneValue.length > 0) {
                let formattedValue = '+7 (' + phoneValue;
                if (formattedValue.length > 7) formattedValue = formattedValue.substring(0, 7) + ') ' + formattedValue.substring(7);
                if (formattedValue.length > 12) formattedValue = formattedValue.substring(0, 12) + '-' + formattedValue.substring(12);
                if (formattedValue.length > 15) formattedValue = formattedValue.substring(0, 15) + '-' + formattedValue.substring(15);
                
                e.target.value = formattedValue;
            }
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
    registerForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        // Валидация формы
        const phone = document.getElementById('register-phone').value;
        const username = document.getElementById('register-username').value;
        const password = document.getElementById('register-password').value;
        
        if (!phone || !username || !password) {
            alert('Пожалуйста, заполните все поля');
            return;
        }
        
        // Проверка формата телефона
        const phoneRegex = /^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$/;
        if (!phoneRegex.test(phone)) {
            alert('Пожалуйста, введите корректный номер телефона');
            return;
        }
        
        // Имитация успешной регистрации
        alert('Регистрация прошла успешно!');
        
        // Переключение на вкладку входа
        authTabs[0].click();
        
        // Заполнение поля входа
        document.getElementById('login-identifier').value = username;
    });

    // Обработка формы входа
    loginForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const identifier = document.getElementById('login-identifier').value;
        const password = document.getElementById('login-password').value;
        
        if (!identifier || !password) {
            alert('Пожалуйста, заполните все поля');
            return;
        }
        
        // Определяем, что ввел пользователь - логин или телефон
        let authMethod = 'login';
        let credentials = {};
        
        // Проверяем, является ли ввод телефоном
        const phoneRegex = /^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$/;
        const isPhone = phoneRegex.test(identifier);
        
        if (isPhone) {
            // Вход по телефону
            authMethod = 'phone';
            credentials.phone = identifier;
            credentials.password = password;
            alert('Вход по телефону выполнен!');
        } else {
            // Вход по логину
            authMethod = 'login';
            credentials.username = identifier;
            credentials.password = password;
            alert('Вход выполнен!');
        }
        
        // Сохраняем данные для запоминания пароля
        if (document.getElementById('remember-me').checked) {
            localStorage.setItem('rememberedUser', JSON.stringify(credentials));
        }
        
        authModal.classList.remove('open');
        setTimeout(() => {
            authModal.style.display = 'none';
            resetForms();
        }, 300);
    });

    // Функция сброса форм
    function resetForms() {
        loginForm.reset();
        registerForm.reset();
    }

    // Проверяем сохраненные данные при загрузке
    const rememberedUser = localStorage.getItem('rememberedUser');
    if (rememberedUser) {
        const userData = JSON.parse(rememberedUser);
        document.getElementById('login-identifier').value = userData.username || userData.phone;
        document.getElementById('remember-me').checked = true;
    }
});