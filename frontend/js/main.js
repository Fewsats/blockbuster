import { initAuth } from './auth.js';
import { initVideoUpload } from './video.js';
// import { Modal } from './modal.js';

document.addEventListener('DOMContentLoaded', () => {
    initAuth();
    initVideoUpload();
});