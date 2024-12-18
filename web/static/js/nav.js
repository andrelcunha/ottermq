document.addEventListener('DOMContentLoaded', function() {
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('nav a');

    navLinks.forEach(link => {
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
    const username = document.getElementById('username');
    const logoutContainer = document.querySelector('.logout-container');
    if (username.textContent === '') {
        logoutContainer.style.display = 'none'
    }
});
