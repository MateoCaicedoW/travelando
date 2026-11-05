import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['track'];

  scroll(event) {
    this.trackTarget.scrollBy({
      left: Number(event.currentTarget.dataset.carouselDirection) * this.trackTarget.clientWidth,
      behavior: 'auto',
    });
  }
}
