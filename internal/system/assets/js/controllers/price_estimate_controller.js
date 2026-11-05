import { Controller } from '@hotwired/stimulus';
import { estimatePrice } from '../price_estimate.mjs';

export default class extends Controller {
  static targets = ['type', 'amount', 'unit'];
  static values = {
    includedGuests: Number,
    extraGuestPrice: Number,
    dayPrice: Number,
    nightPrice: Number,
  };

  connect() {
    this.form = this.element.querySelector('form');
    this.update = this.update.bind(this);
    this.form?.addEventListener('input', this.update);
    this.form?.addEventListener('change', this.update);
    this.update();
  }

  disconnect() {
    this.form?.removeEventListener('input', this.update);
    this.form?.removeEventListener('change', this.update);
  }

  update() {
    if (!this.form) return;
    const values = new FormData(this.form);
    const stayType = values.get('stay_type');
    const estimate = estimatePrice({
      stayType,
      guests: values.get('guests'),
      dateFrom: values.get('date_from'),
      dateTo: values.get('date_to'),
      includedGuests: this.includedGuestsValue,
      extraGuestPrice: this.extraGuestPriceValue,
      dayPrice: this.dayPriceValue,
      nightPrice: this.nightPriceValue,
    });

    this.typeTarget.textContent = stayType === 'overnight' ? 'Estadía' : 'Pasadía';
    if (!estimate) {
      this.amountTarget.textContent = '—';
      this.unitTarget.textContent = 'Revisa huéspedes y fechas';
      return;
    }

    this.amountTarget.textContent = '$' + new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 }).format(estimate.total);
    this.unitTarget.textContent = estimate.unit === 'stay' ? 'COP · estadía completa' : estimate.unit === 'night' ? 'COP / noche' : 'COP · pasadía';
  }
}
