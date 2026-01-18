// Native JavaScript: Theme Toggle
document.addEventListener('DOMContentLoaded', () => {
    const themeToggle = document.getElementById('theme-toggle');
    const body = document.body;
    const icon = themeToggle.querySelector('i');

    themeToggle.addEventListener('click', () => {
        if (body.classList.contains('dark-theme')) {
            body.classList.remove('dark-theme');
            body.classList.add('light-theme');
            icon.classList.remove('fa-moon');
            icon.classList.add('fa-sun');
            
            // Just for demonstration, changing root variables in JS
            document.documentElement.style.setProperty('--bg-dark', '#f1f5f9');
            document.documentElement.style.setProperty('--card-bg', 'rgba(255, 255, 255, 0.8)');
            document.documentElement.style.setProperty('--text-main', '#1e293b');
            document.documentElement.style.setProperty('--border', 'rgba(0,0,0,0.1)');
        } else {
            body.classList.remove('light-theme');
            body.classList.add('dark-theme');
            icon.classList.remove('fa-sun');
            icon.classList.add('fa-moon');
            
            // Revert to dark
            document.documentElement.style.removeProperty('--bg-dark');
            document.documentElement.style.removeProperty('--card-bg');
            document.documentElement.style.removeProperty('--text-main');
            document.documentElement.style.removeProperty('--border');
        }
    });
});

// jQuery: Form result interaction
$(document).ready(function() {
    // Listen for HTMX afterSwap event
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'response-container') {
            const $container = $('#response-container');
            
            // If the response contains an error message, shake the container
            if ($container.find('.error-message').length > 0) {
                $('.login-container').addClass('shake');
                setTimeout(() => {
                    $('.login-container').removeClass('shake');
                }, 500);
            }
            
            // Log successful swap using jQuery
            console.log("HTMX swapped content into #response-container");
        }
    });

    // Native JS logging of a "REST-like" activity (simulated)
    console.log("Hydra Test Portal initialized.");
});
