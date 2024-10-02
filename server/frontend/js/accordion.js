export function initAccordion() {
    const toggleButton = document.getElementById('toggleUploadForm');
    const uploadFormContainer = document.getElementById('uploadFormContainer');
    const videoList = document.getElementById('videoList');
    const accordionIcon = document.getElementById('accordionIcon');

    function toggleUploadForm() {
        const isHidden = uploadFormContainer.classList.toggle('hidden');
        accordionIcon.classList.toggle('rotate-180');
        toggleButton.classList.toggle('rounded-b-none', !isHidden);
        uploadFormContainer.classList.toggle('shadow-md', !isHidden);
    }

    toggleButton.addEventListener('click', toggleUploadForm);

    // Open by default
    if (uploadFormContainer.classList.contains('hidden')) {
        toggleUploadForm();
    }
}

export function updateAccordionState(videoCount) {
    const uploadFormContainer = document.getElementById('uploadFormContainer');
    const toggleButton = document.getElementById('toggleUploadForm');
    const accordionIcon = document.getElementById('accordionIcon');

    if (videoCount > 0) {
        uploadFormContainer.classList.add('hidden');
        toggleButton.classList.remove('rounded-b-none');
        accordionIcon.classList.remove('rotate-180');
    } else {
        uploadFormContainer.classList.remove('hidden');
        toggleButton.classList.add('rounded-b-none');
        accordionIcon.classList.add('rotate-180');
    }
}