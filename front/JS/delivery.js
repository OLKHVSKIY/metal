// delivery.js
document.addEventListener('DOMContentLoaded', function() {
    let selectedVehicle = null;
    let officeCoords = [59.820540, 30.370800];
    let destinationCoords = null;
    let map = null;
    let officePlacemark = null;
    let destinationPlacemark = null;
    let currentRoute = null;

    // Инициализация карты
    function initDeliveryMap() {
        if (typeof ymaps !== 'undefined') {
            ymaps.ready(function() {
                map = new ymaps.Map("delivery-map", {
                    center: officeCoords,
                    zoom: 10,
                    controls: ['zoomControl', 'fullscreenControl']
                });

                // Добавляем метку офиса
                officePlacemark = new ymaps.Placemark(officeCoords, {
                    hintContent: 'Наш склад',
                    balloonContent: 'г. СПб, ул. Московское шоссе 46Б, офис 310'
                }, {
                    preset: 'islands#redDotIcon',
                    iconColor: '#d32f2f'
                });

                map.geoObjects.add(officePlacemark);
                map.behaviors.disable('scrollZoom');

                // Обработчик клика по карте
                map.events.add('click', function(e) {
                    if (!selectedVehicle) {
                        showInstruction('Сначала выберите тип машины!');
                        return;
                    }

                    const coords = e.get('coords');
                    destinationCoords = coords;
                    calculateRoute(coords);
                });

                setTimeout(() => {
                    showInstruction('📍 Выберите тип машины, затем кликните на карте для построения маршрута');
                }, 1000);
            });
        }
    }

    // Показать инструкцию
    function showInstruction(text) {
        const instruction = document.getElementById('mapInstruction');
        instruction.textContent = text;
        instruction.classList.add('show');
        
        setTimeout(() => {
            instruction.classList.remove('show');
        }, 3000);
    }

    // Удалить все маршруты и метки назначения
    function clearMap() {
        if (currentRoute) {
            map.geoObjects.remove(currentRoute);
            currentRoute = null;
        }
        
        if (destinationPlacemark) {
            map.geoObjects.remove(destinationPlacemark);
            destinationPlacemark = null;
        }
    }

    // Расчет маршрута
    function calculateRoute(coords) {
        if (!selectedVehicle || !map) return;

        clearMap();

        // Добавляем новую метку назначения
        destinationPlacemark = new ymaps.Placemark(coords, {
            hintContent: 'Адрес доставки',
            balloonContent: 'Выбранный адрес доставки'
        }, {
            preset: 'islands#blueDotIcon'
        });

        map.geoObjects.add(destinationPlacemark);

        // Строим маршрут
        ymaps.route([officeCoords, coords]).then(function(route) {
            currentRoute = route;
            
            route.options.set({
                strokeColor: '#d32f2f',
                strokeWidth: 4,
                strokeOpacity: 0.7
            });
            
            map.geoObjects.add(route);
            
            const distance = route.getLength() / 1000;
            const isSpb = checkIfInSPB(coords);
            let price = calculatePrice(selectedVehicle.price, distance, isSpb);
            
            showDeliveryResult(price, distance, isSpb);
            
        }).catch(function(error) {
            console.error('Ошибка построения маршрута:', error);
            showInstruction('Ошибка построения маршрута');
        });
    }

    // Проверка, находится ли точка в СПб
    function checkIfInSPB(coords) {
        const spbBounds = {
            north: 60.1,
            south: 59.7,
            west: 29.4,
            east: 30.8
        };
        
        return coords[0] >= spbBounds.south && coords[0] <= spbBounds.north &&
               coords[1] >= spbBounds.west && coords[1] <= spbBounds.east;
    }

    // Расчет стоимости доставки
    function calculatePrice(basePrice, distance, isSpb) {
        if (!isSpb) return null;

        let price = basePrice + (distance * 50);
        const weightMultiplier = selectedVehicle.capacity / 5;
        price *= weightMultiplier;
        
        return Math.round(price);
    }

    // Показать блок с результатами
    function showDeliveryResult(price, distance, isSpb) {
        const resultBlock = document.getElementById('deliveryResult');
        resultBlock.style.display = 'block';

        document.getElementById('resultVehicle').textContent = selectedVehicle.title;
        document.getElementById('resultDistance').textContent = `${distance.toFixed(1)} км`;
        document.getElementById('resultCapacity').textContent = `до ${selectedVehicle.capacity} тонн`;
        document.getElementById('resultRegion').textContent = isSpb ? 'Санкт-Петербург' : 'За пределами СПб';

        if (isSpb && price) {
            document.getElementById('resultPrice').textContent = `${price} ₽`;
            document.getElementById('resultNote').textContent = 'Окончательная стоимость может отличаться';
        } else {
            document.getElementById('resultPrice').textContent = 'Уточните у менеджера';
            document.getElementById('resultNote').textContent = 'Для доставки за пределы СПб свяжитесь с нами';
        }
    }

    // Обработчики выбора машины
    function setupVehicleSelection() {
        const vehicleCards = document.querySelectorAll('.vehicle-card');
        
        vehicleCards.forEach(card => {
            card.addEventListener('click', function() {
                vehicleCards.forEach(c => c.classList.remove('selected'));
                this.classList.add('selected');
                
                selectedVehicle = {
                    title: this.querySelector('.vehicle-title').textContent,
                    capacity: parseFloat(this.dataset.capacity),
                    price: parseFloat(this.dataset.price)
                };
                
                showInstruction('✅ Теперь кликните на карте для построения маршрута');
                
                clearMap();
                const resultBlock = document.getElementById('deliveryResult');
                resultBlock.style.display = 'none';
                
                if (destinationCoords) {
                    calculateRoute(destinationCoords);
                }
            });
        });
    }

    // Инициализация
    function init() {
        initDeliveryMap();
        setupVehicleSelection();
    }

    // Запуск
    if (typeof ymaps !== 'undefined') {
        ymaps.ready(init);
    }
});