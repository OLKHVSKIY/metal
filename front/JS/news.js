document.addEventListener('DOMContentLoaded', () => {
    const newsGrid = document.querySelector('.news-grid');
    const modal = document.querySelector('#newsModal');
    const closeBtn = document.querySelector('.modal .close-btn');
    const newsTitle = document.querySelector('.news-title');
    const newsDate = document.querySelector('.news-date');
    const newsFullText = document.querySelector('.news-full-text');
    const timelineLinks = document.querySelectorAll('.timeline-link');
    let newsData = [];

    function formatDateToRu(iso) {
        // iso expected YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS
        if (!iso) return '';
        const parts = iso.substring(0,10).split('-');
        if (parts.length !== 3) return iso;
        return parts[2] + '.' + parts[1] + '.' + parts[0];
    }

    function withCacheBuster(url) {
        if (!url) return '';
        const sep = url.includes('?') ? '&' : '?';
        return url + sep + 'v=' + Date.now();
    }

    function toProxiedImage(url) {
        if (!url) return '';
        // If it is an absolute external URL, go through backend proxy to avoid hotlink/cdn cache issues
        const isExternal = /^https?:\/\//i.test(url);
        return isExternal ? ('/api/image-proxy?u=' + encodeURIComponent(url)) : url;
    }

    async function loadNews(year) {
        try {
            const url = year && year !== 'all' ? '/api/news?year=' + encodeURIComponent(year) : '/api/news';
            const res = await fetch(url, { cache: 'no-store', headers: { 'Cache-Control': 'no-cache' } });
            const items = await res.json();
            newsData = items.map(n => ({
                id: n.id,
                year: parseInt((n.published_at||'').substring(0,4)) || 0,
                date: formatDateToRu(n.published_at),
                title: n.title,
                shortText: n.short_text,
                fullText: n.full_text,
                imageUrl: n.image_url || ''
            }));
            renderNews(year||'all');
        } catch (e) {
            console.error('Failed to load news', e);
            newsGrid.innerHTML = '<p>Не удалось загрузить новости.</p>';
        }
    }

    function renderNews(year) {
        newsGrid.innerHTML = '';
        const filteredNews = year === 'all' ? newsData : newsData.filter(news => news.year === parseInt(year));
        
        filteredNews.forEach((news, index) => {
            const newsBlock = document.createElement('div');
            newsBlock.classList.add('news-block');
            newsBlock.dataset.id = news.id;
            const imageSrc = withCacheBuster(toProxiedImage(news.imageUrl));
            const imageHtml = news.imageUrl ? `<div class="news-image" style="margin-bottom:10px"><img src="${imageSrc}" alt="" style="width:100%;height:160px;object-fit:cover;border-radius:8px;" onerror="this.style.display='none'"/></div>` : '';
            newsBlock.innerHTML = `
                ${imageHtml}
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
            loadNews(year);
        });
    });

    newsGrid.addEventListener('click', (e) => {
        const block = e.target.closest('.news-block');
        if (!block) return;
        const newsId = parseInt(block.dataset.id);
        if (!isNaN(newsId)) {
            window.location.href = `/back/news/${newsId}`;
        }
    });

    // Modal no longer used, keep listeners noop for safety
    closeBtn.addEventListener('click', () => {});

    window.addEventListener('click', () => {});

    // Initial load
    loadNews('all');
});