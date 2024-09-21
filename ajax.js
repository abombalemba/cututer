document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('linkForm');
    
    form.addEventListener('submit', (event) => {
        event.preventDefault();

        const originalURLInput = document.getElementById('original_url');
        const shortUrlOutput = document.getElementById('shortUrlOutput');

        fetch('/api', {
            method: 'POST',
            body: new FormData(form)
        }).then((response) => response.json()).then((data) => {
            shortUrlOutput.value = data.short_url;
        }).catch((error) => console.log(':( errror: ', error));
    });
});
