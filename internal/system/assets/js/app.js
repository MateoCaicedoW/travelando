import htmx from 'htmx.org';
import { Application } from '@hotwired/stimulus';
import 'basecoat-css/basecoat';
import 'basecoat-css/combobox';
import 'basecoat-css/dropdown-menu';
import FlashController from './controllers/flash_controller';
import PortalNavigationController from './controllers/portal_navigation_controller';
import ConfirmationController from './controllers/confirmation_controller';
import ModalController from './controllers/modal_controller';
import PhotoCarouselController from './controllers/photo_carousel_controller';
import PriceEstimateController from './controllers/price_estimate_controller';
import SearchableSelectController from './controllers/searchable_select_controller';

const application = Application.start();
application.register('portal-navigation', PortalNavigationController);
application.register('confirmation', ConfirmationController);
application.register('flash', FlashController);
application.register('modal', ModalController);
application.register('photo-carousel', PhotoCarouselController);
application.register('price-estimate', PriceEstimateController);
application.register('searchable-select', SearchableSelectController);
window.Stimulus = application;

htmx.config.responseHandling.unshift(
  { code: '400', swap: true, error: false },
  { code: '401', swap: true, error: false },
  { code: '409', swap: true, error: false },
  { code: '422', swap: true, error: false },
);
window.htmx = htmx;
