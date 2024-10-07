document.addEventListener('DOMContentLoaded', () => {
    initProfile();
});

async function checkAuth() {
    try {
        const response = await fetch('/me');
        if (response.ok) {
            console.log('me response: ', response)
            return await response.json();
        } else {
            return null;
        }
    } catch (error) {
        console.error('Auth check failed:', error);
        return null;
    }
}

async function initProfile() {
    const profileForm = document.getElementById('profileForm');
    const profileEmail = document.getElementById('profileEmail');
    const lightningAddress = document.getElementById('lightningAddress');

    // Check if user is authenticated
    const { user } = await checkAuth();
    if (!user) {
        window.location.href = '/';
        return;
    }

    // Populate form with user data
    profileEmail.value = user.email;
    lightningAddress.value = user.lightning_address || '';

    profileForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const updatedLightningAddress = lightningAddress.value;

        try {
            const response = await fetch('/auth/profile', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ lightningAddress: updatedLightningAddress }),
            });

            if (response.ok) {
                Swal.fire({
                    icon: 'success',
                    title: 'Success',
                    text: 'Profile updated successfully!',
                });
            } else {
                const data = await response.json();
                throw new Error(data.error || 'Failed to update profile');
            }
        } catch (error) {
            Swal.fire({
                icon: 'error',
                title: 'Error',
                text: error.message,
            });
        }
    });
}