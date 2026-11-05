import htmx from 'htmx.org';
import { Application } from '@hotwired/stimulus';
import ModalController from './controllers/modal_controller';
import PriceEstimateController from './controllers/price_estimate_controller';

const application = Application.start();
application.register('modal', ModalController);
application.register('price-estimate', PriceEstimateController);
window.Stimulus = application;

htmx.config.responseHandling.unshift(
  { code: '400', swap: true, error: false },
  { code: '401', swap: true, error: false },
  { code: '422', swap: true, error: false },
);
window.htmx = htmx;

function initializeFlash() {
  const flash = document.querySelector('[data-global-flash]');
  if (!flash || flash.dataset.flashInitialized) return;
  flash.dataset.flashInitialized = 'true';

  let timeout;
  const dismiss = () => {
    flash.classList.add('translate-y-2', 'opacity-0');
    window.setTimeout(() => flash.remove(), 200);
  };

  flash.querySelector('[data-flash-dismiss]')?.addEventListener('click', () => {
    window.clearTimeout(timeout);
    dismiss();
  });
  if (flash.dataset.flashLevel === 'success') {
    timeout = window.setTimeout(dismiss, 6000);
  }
}

htmx.onLoad(initializeFlash);

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initializeFlash, { once: true });
} else {
  initializeFlash();
}
