document.addEventListener('DOMContentLoaded', () => {
    const newsGrid = document.querySelector('.news-grid');
    const modal = document.querySelector('#newsModal');
    const closeBtn = document.querySelector('.modal .close-btn');
    const newsTitle = document.querySelector('.news-title');
    const newsDate = document.querySelector('.news-date');
    const newsFullText = document.querySelector('.news-full-text');
    const timelineLinks = document.querySelectorAll('.timeline-link');

    const newsData = [
        {
            id: 1,
            year: 2024,
            date: '15 января 2024',
            title: 'Рост производства стали в России на 5%',
            shortText: 'По итогам 2023 года производство стали в России выросло на 5%, достигнув рекордных показателей.',
            fullText: 'По итогам 2023 года производство стали в России выросло на 5%, достигнув рекордных показателей. Это связано с увеличением внутреннего спроса и экспорта. Крупные компании, такие как НЛМК и Severstal, внесли основной вклад в рост. Ожидается дальнейшее увеличение в 2024 году благодаря новым инвестициям в модернизацию производства.'
        },
        {
            id: 2,
            year: 2024,
            date: '20 марта 2024',
            title: 'Новые инвестиции в металлургию',
            shortText: 'Правительство России выделило дополнительные средства на развитие металлургической отрасли.',
            fullText: 'Правительство России выделило дополнительные средства на развитие металлургической отрасли. Инвестиции направлены на модернизацию оборудования и внедрение экологичных технологий. Это поможет снизить углеродный след и повысить конкурентоспособность российских металлов на мировом рынке.'
        },
        {
            id: 3,
            year: 2024,
            date: '10 июня 2024',
            title: 'Увеличение экспорта металлопроката',
            shortText: 'Экспорт металлопроката из России вырос на 10% в первом полугодии 2024 года.',
            fullText: 'Экспорт металлопроката из России вырос на 10% в первом полугодии 2024 года. Основные направления — Китай, Турция и страны ЕС. Рост обусловлен высоким спросом на российскую сталь и благоприятными ценами.'
        },
        {
            id: 4,
            year: 2024,
            date: '05 августа 2024',
            title: 'Инновации в обработке металла',
            shortText: 'Внедрение новых технологий обработки металла повышает эффективность производства.',
            fullText: 'Внедрение новых технологий обработки металла повышает эффективность производства. Компании переходят на автоматизированные линии, что снижает затраты и улучшает качество продукции. Ожидается рост производства на 7% к концу года.'
        },
        {
            id: 5,
            year: 2024,
            date: '25 ноября 2024',
            title: 'Снижение цен на сырье для металлургии',
            shortText: 'Снижение цен на железную руду благоприятно сказывается на отрасли.',
            fullText: 'Снижение цен на железную руду благоприятно сказывается на отрасли. Это позволяет металлургическим компаниям снизить себестоимость продукции и увеличить прибыль. Аналитики прогнозируют стабильные цены в 2025 году.'
        },
        {
            id: 6,
            year: 2025,
            date: '12 февраля 2025',
            title: 'Рекордный импорт оборудования для металлургии',
            shortText: 'Россия импортировала рекордное количество оборудования для металлургии.',
            fullText: 'Россия импортировала рекордное количество оборудования для металлургии. Это позволит модернизировать заводы и увеличить производство стали на 8% в 2025 году. Основные поставщики — Китай и Германия.'
        },
        {
            id: 7,
            year: 2025,
            date: '08 мая 2025',
            title: 'Новые экологические стандарты в металлопромышленности',
            shortText: 'Введены новые ГОСТы для снижения выбросов в металлургии.',
            fullText: 'Введены новые экологические стандарты в металлопромышленности. Компании обязаны внедрять зеленые технологии, что приведет к снижению выбросов CO2 на 15%. Это повысит конкурентоспособность на глобальном рынке.'
        },
        {
            id: 8,
            year: 2025,
            date: '22 июля 2025',
            title: 'Рост спроса на арматуру в строительстве',
            shortText: 'Спрос на арматуру вырос на 20% из-за строительного бума.',
            fullText: 'Спрос на арматуру вырос на 20% из-за строительного бума в России. Крупные проекты инфраструктуры стимулируют производство, что приводит к созданию новых рабочих мест в отрасли.'
        },
        {
            id: 9,
            year: 2025,
            date: '09 сентября 2025',
            title: 'Новые партнерства с Китаем в сталелитейной торговле',
            shortText: 'Россия и Китай заключили новые соглашения о поставках стали.',
            fullText: 'Россия и Китай заключили новые соглашения о поставках стали. Это усилит экспорт и стабилизирует цены на рынке. Ожидается рост объемов торговли на 25% в ближайший год.'
        }
    ];

    function renderNews(year) {
        newsGrid.innerHTML = '';
        const filteredNews = year === 'all' ? newsData : newsData.filter(news => news.year === parseInt(year));
        
        filteredNews.forEach((news, index) => {
            const newsBlock = document.createElement('div');
            newsBlock.classList.add('news-block');
            newsBlock.dataset.id = news.id;
            newsBlock.innerHTML = `
                <p class="news-date">${news.date}</p>
                <h4 class="news-title">${news.title}</h4>
                <p class="news-short-text">${news.shortText}</p>
            `;
            newsGrid.appendChild(newsBlock);
            setTimeout(() => {
                newsBlock.classList.add('visible');
            }, index * 100);
        });
    }

    timelineLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            timelineLinks.forEach(l => l.classList.remove('active'));
            link.classList.add('active');
            const year = link.dataset.year;
            renderNews(year);
        });
    });

    newsGrid.addEventListener('click', (e) => {
        const block = e.target.closest('.news-block');
        if (block) {
            const newsId = parseInt(block.dataset.id);
            const news = newsData.find(n => n.id === newsId);

            if (news) {
                newsTitle.textContent = news.title;
                newsDate.textContent = news.date;
                newsFullText.innerHTML = news.fullText;

                modal.style.display = 'flex';
                setTimeout(() => {
                    modal.classList.add('open');
                }, 10);
            }
        }
    });

    closeBtn.addEventListener('click', () => {
        modal.classList.remove('open');
        setTimeout(() => {
            modal.style.display = 'none';
        }, 300);
    });

    window.addEventListener('click', (event) => {
        if (event.target === modal) {
            modal.classList.remove('open');
            setTimeout(() => {
                modal.style.display = 'none';
            }, 300);
        }
    });

    // Initial render
    renderNews('all');
});