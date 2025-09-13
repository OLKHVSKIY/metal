document.addEventListener('DOMContentLoaded', () => {
    const slides = document.querySelector('.slides');
    const slideElements = document.querySelectorAll('.slide');
    const prevButton = document.querySelector('.prev-slide');
    const nextButton = document.querySelector('.next-slide');
    const dots = document.querySelectorAll('.dot');
    let currentIndex = 0;
    const totalSlides = slideElements.length;
    let autoSlideInterval;

    function updateSlide(index) {
        slides.style.transform = `translateX(-${index * 100}%)`;
        dots.forEach(dot => dot.classList.remove('active'));
        dots[index].classList.add('active');
        currentIndex = index;
    }

    function startAutoSlide() {
        autoSlideInterval = setInterval(() => {
            currentIndex = (currentIndex + 1) % totalSlides;
            updateSlide(currentIndex);
        }, 6000);
    }

    function stopAutoSlide() {
        clearInterval(autoSlideInterval);
    }

    nextButton.addEventListener('click', () => {
        stopAutoSlide();
        currentIndex = (currentIndex + 1) % totalSlides;
        updateSlide(currentIndex);
        startAutoSlide();
    });

    prevButton.addEventListener('click', () => {
        stopAutoSlide();
        currentIndex = (currentIndex - 1 + totalSlides) % totalSlides;
        updateSlide(currentIndex);
        startAutoSlide();
    });

    dots.forEach((dot, index) => {
        dot.addEventListener('click', () => {
            stopAutoSlide();
            updateSlide(index);
            startAutoSlide();
        });
    });

    startAutoSlide();
});

/* product-slider */

document.addEventListener('DOMContentLoaded', () => {
    const productSlides = document.querySelector('.product-slides');
    const productSlideElements = document.querySelectorAll('.product-slide');
    const prevProductButton = document.querySelector('.prev-product-slide');
    const nextProductButton = document.querySelector('.next-product-slide');
    let currentProductIndex = 0;
    let totalProductSlides = productSlideElements.length;
    const visibleSlides = 5;

    // If we have less than 15 slides (all catalog types), ensure we render them from catalog.html map
    if (totalProductSlides < 15) {
        const types = [
          {slug:'armatura', title:'Арматура', img:'/img/catalog/catalog1.jpeg'},
          {slug:'truba-profilnaya', title:'Труба профильная', img:'/img/catalog/catalog2.jpeg'},
          {slug:'sortovoy-prokat', title:'Сортовой прокат', img:'/img/catalog/catalog3.jpg'},
          {slug:'truba-kruglaya', title:'Труба круглая', img:'/img/catalog/catalog4.jpeg'},
          {slug:'listovoy-prokat', title:'Листовой прокат', img:'/img/catalog/catalog5.jpeg'},
          {slug:'profnastil', title:'Профнастил', img:'/img/catalog/catalog6.jpeg'},
          {slug:'kovanye-izdeliya', title:'Кованые изделия', img:'/img/catalog/catalog7.png'},
          {slug:'shtaketnik-metallicheskiy', title:'Штакетник металлический', img:'/img/catalog/catalog8.webp'},
          {slug:'setka-metallicheskaia', title:'Сетка металлическая', img:'/img/catalog/catalog9.jpeg'},
          {slug:'stroymaterialy', title:'Стройматериалы', img:'/img/catalog/catalog10.png'},
          {slug:'zabory', title:'Заборы', img:'/img/catalog/catalog11.png'},
          {slug:'krepezh', title:'Крепеж', img:'/img/catalog/catalog12.webp'},
          {slug:'petli', title:'Петли', img:'/img/catalog/catalog13.png'},
          {slug:'fitingi', title:'Фитинги', img:'/img/catalog/catalog14.jpg'},
          {slug:'vintovye-svai', title:'Винтовые сваи', img:'/img/catalog/catalog15.jpg'},
          {slug:'zaglushki-dlya-profilnyh-trub', title:'Заглушки', img:'/img/catalog/catalog16.png'}
        ];
        productSlides.innerHTML = types.map(t => `
          <div class="product-slide">
            <a class="product-card" href="/catalog/${t.slug}/">
              <img src="${t.img}" alt="${t.title}">
              <h4>${t.title}</h4>
            </a>
          </div>
        `).join('');
        totalProductSlides = types.length;
    }

    function updateProductSlide(index) {
        if (index < 0) {
            index = 0;
        } else if (index > totalProductSlides - visibleSlides) {
            index = totalProductSlides - visibleSlides;
        }
        productSlides.style.transform = `translateX(-${index * 20}%)`;
        currentProductIndex = index;
    }

    nextProductButton.addEventListener('click', () => {
        if (currentProductIndex < totalProductSlides - visibleSlides) {
            updateProductSlide(currentProductIndex + 1);
        }
    });

    prevProductButton.addEventListener('click', () => {
        if (currentProductIndex > 0) {
            updateProductSlide(currentProductIndex - 1);
        }
    });
});