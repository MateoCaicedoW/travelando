import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  connect() {
    if (this.element.dataset.flashLevel === 'success') {
      this.dismissTimeout = window.setTimeout(() => this.dismiss(), 6000);
    }
  }

  disconnect() {
    window.clearTimeout(this.dismissTimeout);
    window.clearTimeout(this.removeTimeout);
  }

  dismiss() {
    window.clearTimeout(this.dismissTimeout);
    this.element.classList.add('translate-y-2', 'opacity-0');
    this.removeTimeout = window.setTimeout(() => this.element.remove(), 200);
  }
}
