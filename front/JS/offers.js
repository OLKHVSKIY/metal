document.addEventListener('DOMContentLoaded', () => {
    const toggleBtn = document.querySelector('.toggle-btn');
    const hiddenCards = document.querySelectorAll('.offer-card.hidden');

    toggleBtn.addEventListener('click', () => {
        if (toggleBtn.getAttribute('data-state') === 'show') {
            hiddenCards.forEach(card => {
                card.classList.remove('hidden');
            });
            toggleBtn.textContent = 'Скрыть';
            toggleBtn.setAttribute('data-state', 'hide');
        } else {
            hiddenCards.forEach(card => {
                card.classList.add('hidden');
            });
            toggleBtn.textContent = 'Показать ещё';
            toggleBtn.setAttribute('data-state', 'show');
        }
    });
});