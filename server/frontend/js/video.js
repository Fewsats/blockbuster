import { updateAccordionState } from './accordion.js';

function showSwalNotification(message, type = 'success') {
    const toast = Swal.mixin({
        toast: true,
        position: 'top',
        showConfirmButton: false,
        timer: 3000,
        timerProgressBar: true,
        didOpen: (toast) => {
            toast.addEventListener('mouseenter', Swal.stopTimer)
            toast.addEventListener('mouseleave', Swal.resumeTimer)
        }
    });

    return toast.fire({
        icon: type,
        title: message
    });
}

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

        const formData = new FormData();
        formData.append('title', titleInput.value);
        formData.append('description', descriptionInput.value);
        formData.append('price_in_cents', Math.round(parseFloat(document.getElementById('price_in_usd').value) * 100));
        formData.append('email', uploadEmailInput.value);
        formData.append('cover_image', document.getElementById('coverImageInput').files[0]);

        try {
            // Step 1: Send form with upload video request (excluding video file)
            const response = await fetch('/video/upload', {
                method: 'POST',
                body: formData,
            });
            const data = await response.json();
            if (response.ok) {
                // Step 2: Get the signed URL (uploadURL) from the response
                const { upload_url: uploadURL, video_id: videoId } = data;

                // Step 3: Upload the video file to the signed URL
                const videoFile = formData.get('video');
                const uploadFormData = new FormData();
                uploadFormData.append('file', document.getElementById('videoInput').files[0]);

                const uploadResponse = await fetch(uploadURL, {
                    method: 'POST',
                    body: uploadFormData,
                });

                if (uploadResponse.ok) {
                    const l402Uri = `l402://blockbuster.fewsats.com/video/info/${videoId}`;
                    
                    Swal.fire({
                        title: 'Video uploaded successfully!',
                        html: `Your L402 URI: <br><strong>${l402Uri}</strong>`,
                        icon: 'success',
                        showCancelButton: true,
                        confirmButtonText: 'Copy L402 URI',
                        cancelButtonText: 'Close'
                    }).then((result) => {
                        if (result.isConfirmed) {
                            navigator.clipboard.writeText(l402Uri).then(() => {
                                showSwalNotification('Copied!').then(() => {
                                    setTimeout(() => window.location.reload(), 3000);
                                });
                            }).catch(err => {
                                console.error('Failed to copy: ', err);
                                showSwalNotification('Failed to copy', 'error').then(() => {
                                    setTimeout(() => window.location.reload(), 3000);
                                });
                            });
                        } else {
                            window.location.reload();
                        }
                    });

                } else {
                    throw new Error('Failed to upload video to streams storage');
                }
            } else {
                throw new Error(data.error || 'Failed to initiate video upload');
            }
        } catch (error) {
            Swal.fire('Error', error.message, 'error');
        } finally {
            setUploadingState(false);
        }
    });

    populateEmailField();
}

export function initVideoList() {
    const videoList = document.getElementById('videoList');

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
            <div class="bg-gray-100 rounded-lg shadow-md p-4 mb-4">
                <div class="flex items-center space-x-4 cursor-pointer" onclick="toggleAccordion(${index})">
                    <img src="${video.cover_url}" alt="${video.title}" class="w-16 h-16 rounded-md object-cover">
                    <div class="flex-1">
                        <h4 class="text-lg font-semibold">${video.title}</h4>
                    </div>
                    <div class="flex justify-end">
                        <button id="copyButton${index}" 
                            onclick="event.stopPropagation(); copyL402Uri('${video.l402_info_uri}')"
                            class="mr-1 bg-indigo-600 text-white py-2 px-4 rounded-md text-sm hover:bg-indigo-700 
                            focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 
                            transition duration-150 ease-in-out relative">
                            Copy L402 URI
                        </button>

                        <button id="postOnXButton${index}"
                            onclick="event.stopPropagation(); postOnX('${video.title}', ${video.price_in_cents}, '${video.external_id}')"
                            class="mr-1 bg-black text-white py-2 px-4 rounded-md text-sm hover:bg-indigo-700 
                            focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 
                            transition duration-150 ease-in-out relative">
                            Post on X
                        </button>
                    </div>
                    <div class="text-right">
                        <p class="text-sm font-semibold">$${(video.price_in_cents / 100).toFixed(2)}</p>
                        <p class="text-xs text-gray-500">${video.total_views} views</p>
                        <p class="text-xs text-gray-500">${video.total_purchases} purchases</p>
                    </div>
                    <svg class="w-6 h-6 transform transition-transform duration-200" id="accordionIcon${index}" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                    </svg>
                </div>
                <div class="mt-4 hidden" id="videoDetails${index}">
                    <form onsubmit="updateVideo(event, '${video.external_id}', ${index})" class="space-y-4">
                        <div>
                            <label for="title${index}" class="block text-sm font-medium text-gray-700">Title</label>
                            <input type="text" id="title${index}" name="title" value="${video.title}" 
                                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm 
                                focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 
                                px-3 py-2">
                        </div>
                        <div>
                            <label for="description${index}" class="block text-sm font-medium text-gray-700">Description</label>
                            <textarea id="description${index}" name="description" rows="3" 
                                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm 
                                focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 
                                px-3 py-2">${video.description}</textarea>
                        </div>
                        <div>
                            <label for="price${index}" class="block text-sm font-medium text-gray-700">Price (in cents)</label>
                            <input type="number" id="price${index}" name="price" value="${video.price_in_cents}" 
                                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm 
                                focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 
                                px-3 py-2">
                        </div>
                        <button type="submit" 
                            class="w-full bg-indigo-600 text-white py-2 px-4 rounded-md text-sm 
                            hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 
                            focus:ring-offset-2 transition duration-150 ease-in-out">
                            Update Video
                        </button>
                    </form>
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

function redirectToVideo(uri) {
    window.location.href = `http://videos.l402.org/?uri=${encodeURIComponent(uri)}`;
}

// Make redirectToVideo globally accessible
window.redirectToVideo = redirectToVideo;

function initDropzone(dropzoneId, inputId, fileType) {
    const dropzone = document.getElementById(dropzoneId);
    const input = document.getElementById(inputId);

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropzone.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        dropzone.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropzone.addEventListener(eventName, unhighlight, false);
    });

    function highlight() {
        dropzone.classList.add('dragover');
    }

    function unhighlight() {
        dropzone.classList.remove('dragover');
    }

    dropzone.addEventListener('drop', handleDrop, false);

    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        input.files = files;
        updateDropzoneText(dropzone, files[0], fileType);
    }

    dropzone.addEventListener('click', () => input.click());

    input.addEventListener('change', () => {
        updateDropzoneText(dropzone, input.files[0], fileType);
    });
}

function updateDropzoneText(dropzone, file, fileType) {
    const previewContainer = dropzone.querySelector('.preview-container');
    
    if (file) {
        if (fileType === 'image/') {
            const img = previewContainer.querySelector('.preview-image');
            img.src = URL.createObjectURL(file);
            img.onload = () => URL.revokeObjectURL(img.src);
            img.classList.remove('hidden');
        } else if (fileType === 'video/') {
            const video = previewContainer.querySelector('.preview-video');
            video.src = URL.createObjectURL(file);
            video.onloadedmetadata = () => {
                URL.revokeObjectURL(video.src);
            };
            video.classList.remove('hidden');
        }
    } else {
        const previewElement = previewContainer.querySelector('.preview-image, .preview-video');
        if (previewElement) {
            previewElement.classList.add('hidden');
            previewElement.src = '';
        }
    }
}

initDropzone('coverImageDropzone', 'coverImageInput', 'image/');
initDropzone('videoDropzone', 'videoInput', 'video/');

function copyL402Uri(url) {
    navigator.clipboard.writeText(url).then(() => {
        showSwalNotification('Copied!');
    }).catch(err => {
        console.error('Failed to copy L402 URI: ', err);
        showSwalNotification('Failed to copy', 'error');
    });
}

// Make copyL402Uri globally accessible
window.copyL402Uri = copyL402Uri;


function postOnX(title, priceInCents, videoId) {
    const extensionUrl = 'SHORT URL TO CHROME EXTENSION'; 
    const priceInUSD = (priceInCents / 100).toFixed(2);
    const videoUrl = `https://blockbuster.fewsats.com/video/${videoId}`;
    
    const postText = `Get access to my latest content: "${title}"

💰 Price: $${priceInUSD} USD

One extension away from the best exclusive content: ${extensionUrl}

${videoUrl}`;

    const encodedPost = encodeURIComponent(postText);
    const xPostUrl = `https://twitter.com/intent/tweet?text=${encodedPost}`;

    window.open(xPostUrl, '_blank');
}

// Make postOnX globally accessible
window.postOnX = postOnX;

function toggleAccordion(index) {
    const detailsElement = document.getElementById(`videoDetails${index}`);
    const iconElement = document.getElementById(`accordionIcon${index}`);
    
    detailsElement.classList.toggle('hidden');
    iconElement.classList.toggle('rotate-180');
}

// Make toggleAccordion globally accessible
window.toggleAccordion = toggleAccordion;

async function updateVideo(event, externalId, index) {
    event.preventDefault();
    const form = event.target;
    const title = form.title.value;
    const description = form.description.value;
    const priceInCents = parseInt(form.price.value);

    try {
        const response = await fetch(`/video/${externalId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, description, price_in_cents: priceInCents }),
        });

        if (response.ok) {
            const updatedVideo = await response.json();
            updateVideoInList(index, updatedVideo);
            showSwalNotification('Video updated successfully');
        } else {
            throw new Error('Failed to update video');
        }
    } catch (error) {
        console.error('Error updating video:', error);
        showSwalNotification('Failed to update video', 'error');
    }
}

function updateVideoInList(index, updatedVideo) {
    const videoElement = document.querySelector(`#videoList > div:nth-child(${index + 1})`);
    if (videoElement) {
        videoElement.querySelector('h4').textContent = updatedVideo.title;
        videoElement.querySelector('p').textContent = updatedVideo.description.substring(0, 100) + (updatedVideo.description.length > 100 ? '...' : '');
        videoElement.querySelector('.text-sm.font-semibold').textContent = `$${(updatedVideo.price_in_cents / 100).toFixed(2)}`;
    }
}

// Make updateVideo globally accessible
window.updateVideo = updateVideo;

