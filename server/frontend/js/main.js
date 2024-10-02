import { initAuth } from './auth.js';
import { initVideoUpload, initVideoList } from './video.js';
import { initIntercom } from './intercom.js';
import { initAccordion } from './accordion.js';

document.addEventListener('DOMContentLoaded', () => {
    initAuth();
    initAccordion(); // Move this before initVideoList
    initVideoUpload();
    initVideoList();
    initIntercom();
});