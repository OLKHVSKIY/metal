document.addEventListener('DOMContentLoaded', () => {
    const gostTitles = document.querySelectorAll('.gost-title');
    
    gostTitles.forEach(title => {
        title.addEventListener('click', () => {
            const sublist = title.nextElementSibling;
            const isActive = sublist.classList.contains('active');
            
            // Закрываем все открытые списки
            document.querySelectorAll('.gost-sublist.active').forEach(list => {
                list.classList.remove('active');
                list.previousElementSibling.classList.remove('active');
            });
            
            // Открываем/закрываем текущий список
            if (!isActive) {
                sublist.classList.add('active');
                title.classList.add('active');
            }
        });
    });
});