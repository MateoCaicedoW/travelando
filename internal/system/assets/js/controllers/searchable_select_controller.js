import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['input'];

  connect() {
    this.form = this.element.closest('form');
    if (this.form && this.element.dataset.required === 'true') {
      this.validate = this.validate.bind(this);
      this.form.addEventListener('submit', this.validate, true);
    }
  }

  disconnect() {
    this.form?.removeEventListener('submit', this.validate, true);
  }

  clearValidity() {
    this.inputTarget.setCustomValidity('');
  }

  selectValue() {
    if (this.inputTarget.value) this.inputTarget.select();
  }

  validate(event) {
    if (this.element.querySelector('input[type="hidden"]')?.value) {
      this.clearValidity();
      return;
    }

    this.inputTarget.setCustomValidity('Selecciona una opción de la lista.');
    this.inputTarget.reportValidity();
    event.preventDefault();
    event.stopImmediatePropagation();
  }
}
