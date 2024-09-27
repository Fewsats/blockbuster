export function initVideoUpload() {
    const uploadForm = document.getElementById('uploadForm');
    const uploadEmailInput = document.getElementById('email');
    const titleInput = document.getElementById('title');
    const descriptionInput = document.getElementById('description');
    const titleError = document.getElementById('titleError');
    const descriptionError = document.getElementById('descriptionError');

    function validateField(input, errorElement, minLength) {
        if (input.value.length < minLength) {
            errorElement.classList.remove('hidden');
            return false;
        } else {
            errorElement.classList.add('hidden');
            return true;
        }
    }

    titleInput.addEventListener('input', () => validateField(titleInput, titleError, 10));
    descriptionInput.addEventListener('input', () => validateField(descriptionInput, descriptionError, 25));

    async function populateEmailField() {
        try {
            const response = await fetch('/me');
            if (response.ok) {
                const { user } = await response.json();
                uploadEmailInput.value = user.email;
                uploadEmailInput.readOnly = true; 
            }
        } catch (error) {
            console.error('Failed to fetch user email:', error);
        }
    }

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const isTitleValid = validateField(titleInput, titleError, 10);
        const isDescriptionValid = validateField(descriptionInput, descriptionError, 25);

        if (!isTitleValid || !isDescriptionValid) {
            return;
        }

        const formData = new FormData(uploadForm);

        try {
            // Create a new FormData object excluding the video field
            const metadataFormData = new FormData();
            for (const [key, value] of formData.entries()) {
                if (key !== 'video') {
                    metadataFormData.append(key, value);
                }
            }

            // Step 1: Send form with upload video request (excluding video file)
            const response = await fetch('/video/upload', {
                method: 'POST',
                body: metadataFormData,
            });
            const data = await response.json();
            if (response.ok) {
                // Step 2: Get the signed URL (uploadURL) from the response
                const { upload_url: uploadURL } = data;

                // Step 3: Upload the video file to the signed URL
                const videoFile = formData.get('video');
                const uploadFormData = new FormData();
                uploadFormData.append('file', videoFile);

                const uploadResponse = await fetch(uploadURL, {
                    method: 'POST',
                    body: uploadFormData,
                });

                if (uploadResponse.ok) {
                    alert('Video uploaded successfully!');
                    uploadForm.reset();
                } else {
                    throw new Error('Failed to upload video to streams storage');
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
                displaySignInMessage();
                throw new Error('Failed to fetch user videos');
            }
        } catch (error) {
            console.error('Error fetching user videos:', error);
            displaySignInMessage();
        }
    }

    function displayVideos(videos) {
        if (!Array.isArray(videos) || videos.length === 0) {
            videoList.innerHTML = '<p class="text-gray-500">You haven\'t uploaded any videos yet.</p>';
            return;
        }

        videoList.innerHTML = videos.map(video => `
            <div class="bg-gray-100 rounded-lg shadow-md p-4 flex flex-col space-y-2">
                <div class="flex items-center space-x-4">
                    <img src="${video.cover_url}" alt="${video.title}" class="w-16 h-16 rounded-md object-cover">
                    <div class="flex-1">
                        <h4 class="text-lg font-semibold">${video.title}</h4>
                        <p class="text-sm text-gray-600">${video.description}</p>
                    </div>
                    <div class="flex justify-end">
                        <button onclick="copyL402Uri('${video.l402_info_uri}')" class="bg-indigo-600 text-white py-1 px-2 rounded-md text-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2">Copy L402 URI</button>
                    </div>
                    <div class="text-right">
                        <p class="text-sm font-semibold">$${(video.price_in_cents / 100).toFixed(2)}</p>
                        <p class="text-xs text-gray-500">${video.total_views} views</p>
                    </div>
                </div>
            </div>
        `).join('');
    }

    function displaySignInMessage() {
        videoList.innerHTML = '<p class="text-gray-500">To view the list of your uploaded videos, please sign in first.</p>';
    }

    fetchUserVideos();
}

function copyL402Uri(url) {
    navigator.clipboard.writeText(url).then(() => {
        alert('L402 URI copied to clipboard!');
    }).catch(err => {
        console.error('Failed to copy L402 URI: ', err);
        alert('Failed to copy L402 URI. Please try again.');
    });
}

// Make copyL402Uri globally accessible
window.copyL402Uri = copyL402Uri;