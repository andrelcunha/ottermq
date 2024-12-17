document.addEventListener('DOMContentLoaded', function() {
    const triangles = document.querySelectorAll('.triangle');

    triangles.forEach(triangle => {
        triangle.addEventListener('click', function() {
            const sectionContent = triangle.parentElement.nextElementSibling;
            if (sectionContent.classList.contains('hidden')) {
                sectionContent.classList.remove('hidden');
                this.textContent = '▾';
            } else {
                sectionContent.classList.add('hidden');
                this.textContent = '▸';
            }
        });
    });
});
