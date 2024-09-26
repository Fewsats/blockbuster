export class Modal {
    constructor(modalId) {
        this.modal = document.getElementById(modalId);
        this.setupCloseOnOutsideClick();
    }

    show() {
        this.modal.style.display = 'block';
    }

    hide() {
        this.modal.style.display = 'none';
    }

    setupCloseOnOutsideClick() {
        window.addEventListener('click', (event) => {
            if (event.target === this.modal) {
                this.hide();
            }
        });
    }
}