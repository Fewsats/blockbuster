import { initAuth } from './auth.js';
import { initVideoUpload, initVideoList } from './video.js';
// import { Modal } from './modal.js';

document.addEventListener('DOMContentLoaded', () => {
    initAuth();
    initVideoUpload();
    initVideoList();
});