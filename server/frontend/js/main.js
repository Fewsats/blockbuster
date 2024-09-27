import { initAuth } from './auth.js';
import { initVideoUpload, initVideoList } from './video.js';
import { initIntercom } from './intercom.js';

document.addEventListener('DOMContentLoaded', () => {
    initAuth();
    initVideoUpload();
    initVideoList();
    initIntercom();
});