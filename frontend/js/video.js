export function initVideoUpload() {
    const uploadForm = document.getElementById('uploadForm');
    const uploadEmailInput = document.getElementById('uploadEmail');

    async function populateEmailField() {
        try {
            const response = await fetch('/me');
            if (response.ok) {
                const data = await response.json();
                uploadEmailInput.value = data.email;
                uploadEmailInput.disabled = true;
            }
        } catch (error) {
            console.error('Failed to fetch user email:', error);
        }
    }

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(uploadForm);
        try {
            const response = await fetch('/video/upload', {
                method: 'POST',
                body: formData,
            });
            const data = await response.json();
            if (response.ok) {
                alert('Video uploaded successfully!');
                uploadForm.reset();
            } else {
                throw new Error(data.error || 'Failed to upload video');
            }
        } catch (error) {
            alert(error.message);
        }
    });

    populateEmailField();
}