import { Modal } from '/static/js/modal.js';

export function initAuth() {
    const userInfo = document.getElementById('userInfo');
    const authButton = document.getElementById('authButton');
    const authModal = new Modal('authModal');
    const authForm = document.getElementById('authForm');
    const messageElement = document.getElementById('message');

    function updateUserInfo(email) {
        if (email) {
            userInfo.textContent = email;
            authButton.textContent = 'Sign Out';
        } else {
            userInfo.textContent = '';
            authButton.textContent = 'Sign In / Sign Up';
        }
    }

    async function checkAuth() {
        try {
            const response = await fetch('/me');
            if (response.ok) {
                const data = await response.json();
                updateUserInfo(data.email);
            } else {
                updateUserInfo(null);
            }
        } catch (error) {
            console.error('Auth check failed:', error);
            updateUserInfo(null);
        }
    }

    authButton.addEventListener('click', async () => {
        if (authButton.textContent === 'Sign Out') {
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
        } else {
            authModal.show();
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