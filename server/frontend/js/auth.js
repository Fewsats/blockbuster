import { Modal } from '/static/js/modal.js';

export function initAuth() {
    const userInfoContainer = document.getElementById('userInfoContainer');
    const userInfo = document.getElementById('userInfo');
    const userEmail = document.getElementById('userEmail');
    const userEmailSpan = userEmail.querySelector('span');
    const userInitials = userEmail.querySelector('div');
    const userDropdown = document.getElementById('userDropdown');
    const authButton = document.getElementById('authButton');
    const signOutButton = document.getElementById('signOutButton');
    const authModal = new Modal('authModal');
    const authForm = document.getElementById('authForm');
    const messageElement = document.getElementById('message');

    function updateUserInfo(email) {
        if (email) {
            userEmailSpan.textContent = email;
            userInitials.textContent = email[0].toUpperCase();
            authButton.style.display = 'none';
            userInfo.style.display = 'block';
        } else {
            authButton.style.display = 'block';
            userInfo.style.display = 'none';
        }
    }

    async function checkAuth() {
        try {
            const response = await fetch('/me');
            if (response.ok) {
                const { user } = await response.json();
                updateUserInfo(user.email);
            } else {
                updateUserInfo(null);
            }
        } catch (error) {
            console.error('Auth check failed:', error);
            updateUserInfo(null);
        }
    }

    userEmail.addEventListener('click', () => {
        userDropdown.classList.toggle('hidden');
    });

    document.addEventListener('click', (event) => {
        if (!userInfoContainer.contains(event.target)) {
            userDropdown.classList.add('hidden');
        }
    });

    authButton.addEventListener('click', () => {
        authModal.show();
    });

    signOutButton.addEventListener('click', async (e) => {
        e.preventDefault();
        try {
            const response = await fetch('/auth/logout');
            if (response.ok) {
                updateUserInfo(null);
            } else {
                throw new Error('Logout failed');
            }
        } catch (error) {
            console.error('Logout failed:', error);
        }
    });

    authForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('login_email').value;
        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email }),
            });
            const data = await response.json();
            if (response.ok) {
                messageElement.textContent = data.message;
                messageElement.className = 'mt-4 text-sm text-center text-green-600';
                setTimeout(() => authModal.hide(), 3000);
            } else {
                throw new Error(data.error || 'An error occurred');
            }
        } catch (error) {
            messageElement.textContent = error.message;
            messageElement.className = 'mt-4 text-sm text-center text-red-600';
        }
    });

    checkAuth();
}