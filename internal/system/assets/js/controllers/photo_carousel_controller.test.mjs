import assert from 'node:assert/strict';
import test from 'node:test';
import PhotoCarouselController from './photo_carousel_controller.js';

test('scrolls the photo track by one page in the selected direction', () => {
  const controller = Object.create(PhotoCarouselController.prototype);
  controller.trackTarget = {
    clientWidth: 320,
    scrollWidth: 960,
    scrollLeft: 320,
    scrollTo(options) { this.options = options; },
  };

  controller.scroll({ currentTarget: { dataset: { carouselDirection: '-1' } } });
  assert.deepEqual(controller.trackTarget.options, { left: 0, behavior: 'auto' });
});

test('wraps from the last photo to the first', () => {
  const controller = Object.create(PhotoCarouselController.prototype);
  controller.trackTarget = {
    clientWidth: 320,
    scrollWidth: 960,
    scrollLeft: 640,
    scrollTo(options) { this.options = options; },
  };

  controller.scroll({ currentTarget: { dataset: { carouselDirection: '1' } } });
  assert.deepEqual(controller.trackTarget.options, { left: 0, behavior: 'auto' });
});

test('wraps from the first photo to the last', () => {
  const controller = Object.create(PhotoCarouselController.prototype);
  controller.trackTarget = {
    clientWidth: 320,
    scrollWidth: 960,
    scrollLeft: 0,
    scrollTo(options) { this.options = options; },
  };

  controller.scroll({ currentTarget: { dataset: { carouselDirection: '-1' } } });
  assert.deepEqual(controller.trackTarget.options, { left: 640, behavior: 'auto' });
});
