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
    const totalProductSlides = productSlideElements.length;
    const visibleSlides = 5;

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