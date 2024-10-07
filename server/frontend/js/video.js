import { updateAccordionState } from './accordion.js';

export function initVideoUpload() {
    const uploadForm = document.getElementById('uploadForm');
    const uploadEmailInput = document.getElementById('email');
    const titleInput = document.getElementById('title');
    const descriptionInput = document.getElementById('description');
    const titleError = document.getElementById('titleError');
    const descriptionError = document.getElementById('descriptionError');
    const uploadButton = document.getElementById('uploadButton');
    const uploadSpinner = document.getElementById('uploadSpinner');

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

    function setUploadingState(isUploading) {
        uploadButton.disabled = isUploading;
        uploadButton.querySelector('span').textContent = isUploading ? 'Uploading...' : 'Upload Video';
        uploadSpinner.classList.toggle('hidden', !isUploading);
    }

    uploadForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const isTitleValid = validateField(titleInput, titleError, 10);
        const isDescriptionValid = validateField(descriptionInput, descriptionError, 25);

        if (!isTitleValid || !isDescriptionValid) {
            return;
        }

        setUploadingState(true);

        const formData = new FormData(uploadForm);

        // Convert price from USD to cents
        const priceInUSD = formData.get('price_in_usd');
        const priceInCents = Math.round(parseFloat(priceInUSD) * 100);
        formData.set('price_in_cents', priceInCents);
        formData.delete('price_in_usd');

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
        } finally {
            setUploadingState(false);
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
                // Update accordion state after populating the video list
                updateAccordionState(r.videos.length);
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
            updateAccordionState(0);
            return;
        }

        videoList.innerHTML = videos.map((video, index) => `
            <div class="bg-gray-100 rounded-lg shadow-md p-4 flex flex-col space-y-2 cursor-pointer" 
                 onclick="redirectToVideo('${video.l402_info_uri}')">
                <div class="flex items-center space-x-4">
                    <img src="${video.cover_url}" alt="${video.title}" class="w-16 h-16 rounded-md object-cover">
                    <div class="flex-1">
                        <h4 class="text-lg font-semibold">${video.title}</h4>
                    </div>
                    <div class="flex justify-end">
                        <button id="copyButton${index}" 
                            onclick="event.stopPropagation(); copyL402Uri('${video.l402_info_uri}', 'copyButton${index}')" 
                            class="bg-indigo-600 text-white py-2 px-4 rounded-md text-sm hover:bg-indigo-700 
                            focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 
                            transition duration-150 ease-in-out relative">
                            Copy L402 URI
                        </button>
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
        updateAccordionState(0);
    }

    fetchUserVideos();
}

function copyL402Uri(url, buttonId) {
    navigator.clipboard.writeText(url).then(() => {
        showNotification('Copied!', buttonId, 'success');
    }).catch(err => {
        console.error('Failed to copy L402 URI: ', err);
        showNotification('Failed to copy', buttonId, 'error');
    });
}

function showNotification(message, buttonId, type = 'success') {
    const button = document.getElementById(buttonId);
    const notification = document.createElement('div');
    notification.textContent = message;
    notification.className = `absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-full 
        px-2 py-1 rounded text-xs text-white bg-gray-800
        opacity-0 transition-opacity duration-300 pointer-events-none`;
    
    button.style.position = 'relative';
    button.appendChild(notification);

    // Position the notification higher above the button
    notification.style.top = '-0.1rem';

    setTimeout(() => {
        notification.style.opacity = '1';
        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                button.removeChild(notification);
            }, 300);
        }, 2000);
    }, 10);
}

// Make copyL402Uri globally accessible
window.copyL402Uri = copyL402Uri;

function redirectToVideo(uri) {
    window.location.href = `http://videos.l402.org/?uri=${encodeURIComponent(uri)}`;
}

// Make redirectToVideo globally accessible
window.redirectToVideo = redirectToVideo;