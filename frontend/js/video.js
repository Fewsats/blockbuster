export function initVideoUpload() {
    const uploadForm = document.getElementById('uploadForm');
    const uploadEmailInput = document.getElementById('email');

    async function populateEmailField() {
        try {
            const response = await fetch('/me');
            if (response.ok) {
                const data = await response.json();
                uploadEmailInput.value = data.email;
                uploadEmailInput.readOnly = true;  // Change from disabled to readOnly
            }
        } catch (error) {
            console.error('Failed to fetch user email:', error);
        }
    }

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(uploadForm);

        try {
            // Step 1: Send form with upload video request
            const response = await fetch('/video/upload', {
                method: 'POST',
                body: formData,
            });
            const data = await response.json();
            if (response.ok) {
                // Step 2: Get the signed URL (uploadURL) from the response
                const { upload_url: uploadURL } = data;

                // Step 3: Upload the video file to the signed URL
                const videoFile = formData.get('video');
                const uploadResponse = await fetch(uploadURL, {
                    method: 'PUT',
                    body: videoFile,
                    headers: {
                        'Content-Type': videoFile.type,
                    },
                });

                if (uploadResponse.ok) {
                    alert('Video uploaded successfully!');
                    uploadForm.reset();
                } else {
                    throw new Error('Failed to upload video to storage');
                }
            } else {
                throw new Error(data.error || 'Failed to initiate video upload');
            }
        } catch (error) {
            alert(error.message);
        }
    });

    populateEmailField();
}