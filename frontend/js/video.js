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

export function initVideoList() {
    const videoList = document.getElementById('videoList');
    const userVideos = document.getElementById('userVideos');

    async function fetchUserVideos() {
        try {
            const response = await fetch('/user/videos');
            if (response.ok) {
                const r = await response.json();
                displayVideos(r.videos);
            } else {
                throw new Error('Failed to fetch user videos');
            }
        } catch (error) {
            console.error('Error fetching user videos:', error);
            userVideos.style.display = 'none';
        }
    }

    function displayVideos(videos) {
        console.log('videos', videos);
        if (!Array.isArray(videos) || videos.length === 0) {
            videoList.innerHTML = '<p class="text-gray-500">You haven\'t uploaded any videos yet.</p>';
            return;
        }

        videoList.innerHTML = videos.map(video => `
            <div class="flex items-center justify-between border-b border-gray-200 pb-4">
                <div class="flex items-center space-x-4">
                    <img src="${video.cover_url}" alt="${video.title}" class="w-24 h-16 object-cover rounded">
                    <div>
                        <h4 class="font-semibold">${video.title}</h4>
                        <p class="text-sm text-gray-500">${video.description}</p>
                    </div>
                </div>
                <div class="text-right">
                    <p class="text-sm font-semibold">$${(video.price_in_cents / 100).toFixed(2)}</p>
                    <p class="text-xs text-gray-500">${video.total_views} views</p>
                </div>
            </div>
        `).join('');
    }

    fetchUserVideos();
}