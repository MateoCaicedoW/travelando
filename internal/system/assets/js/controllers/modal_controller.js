import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  connect() {
    this.boundHandleKeydown = this.handleKeydown.bind(this);
    this.element.classList.add('modal-animate-in');
    document.body.classList.add('overflow-hidden');
    document.addEventListener('keydown', this.boundHandleKeydown);
    this.focusTimeout = window.setTimeout(() => this.getFocusableElements()[0]?.focus(), 50);
  }

  disconnect() {
    document.removeEventListener('keydown', this.boundHandleKeydown);
    document.body.classList.remove('overflow-hidden');
    window.clearTimeout(this.focusTimeout);
    window.clearTimeout(this.closeTimeout);
  }

  handleKeydown(event) {
    if (event.key === 'Escape') {
      event.preventDefault();
      this.close();
      return;
    }
    if (event.key !== 'Tab') return;

    const focusable = this.getFocusableElements();
    if (focusable.length === 0) {
      event.preventDefault();
      this.element.querySelector('[role="dialog"]')?.focus();
      return;
    }
    const first = document.activeElement === focusable[0];
    const last = document.activeElement === focusable.at(-1);
    if ((event.shiftKey && first) || (!event.shiftKey && last)) {
      event.preventDefault();
      focusable[event.shiftKey ? focusable.length - 1 : 0].focus();
    }
  }

  getFocusableElements() {
    const selector = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';
    return Array.from(this.element.querySelectorAll(selector))
      .filter((element) => !element.disabled && element.offsetParent !== null);
  }

  close() {
    if (this.element.classList.contains('modal-animate-out')) return;
    this.element.classList.remove('modal-animate-in');
    this.element.classList.add('modal-animate-out');
    document.body.classList.remove('overflow-hidden');
    this.closeTimeout = window.setTimeout(() => {
      this.element.remove();
    }, 150);
  }
}
