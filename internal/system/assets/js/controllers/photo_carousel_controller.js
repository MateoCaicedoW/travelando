import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['track'];

  get photos() {
    return Array.from(this.trackTarget.querySelectorAll('img'));
  }

  scroll(event) {
    const track = this.trackTarget;
    const maxScroll = track.scrollWidth - track.clientWidth;
    const nextScroll = track.scrollLeft + Number(event.currentTarget.dataset.carouselDirection) * track.clientWidth;
    track.scrollTo({
      left: nextScroll > maxScroll ? 0 : nextScroll < 0 ? maxScroll : nextScroll,
      behavior: 'auto',
    });
  }

  async open(event) {
    event.preventDefault();
    if (this.opening) return;
    this.opening = true;

    try {
      const photos = this.photos;
      const selectedIndex = Number(event.currentTarget.dataset.photoCarouselIndex);
      await Promise.all(photos.map((photo) => {
        photo.loading = 'eager';
        return photo.decode().catch(() => {});
      }));

      const dataSource = [];
      let index = -1;
      photos.forEach((photo, photoIndex) => {
        if (!photo.naturalWidth || !photo.naturalHeight) return;
        if (photoIndex === selectedIndex) index = dataSource.length;
        dataSource.push({
          src: photo.currentSrc || photo.src,
          msrc: photo.currentSrc || photo.src,
          width: photo.naturalWidth,
          height: photo.naturalHeight,
          alt: photo.alt,
        });
      });
      if (index < 0) return;

      const { default: PhotoSwipe } = await import('photoswipe');
      const fitLevels = new Map();
      const zoomLevel = (zoom, scale) => {
        fitLevels.set(zoom.index, zoom.fit);
        return zoom.fit * scale;
      };
      const gallery = new PhotoSwipe({
        dataSource,
        index,
        initialZoomLevel: (zoom) => zoomLevel(zoom, 1),
        secondaryZoomLevel: (zoom) => zoomLevel(zoom, 2),
        maxZoomLevel: (zoom) => zoomLevel(zoom, 4),
        zoomAnimationDuration: 120,
        imageClickAction: false,
        doubleTapAction: 'zoom',
        wheelToZoom: true,
        closeTitle: 'Cerrar galería',
        zoomTitle: 'Ampliar imagen',
        arrowPrevTitle: 'Foto anterior',
        arrowNextTitle: 'Foto siguiente',
        errorMsg: 'No se pudo cargar la imagen',
        indexIndicatorSep: ' / ',
      });
      let zoomStep = 0;
      gallery.on('change', () => { zoomStep = 0; });
      gallery.init();
      gallery.element.addEventListener('dblclick', (doubleClick) => {
        if (!doubleClick.target.closest('.pswp__zoom-wrap')) return;
        doubleClick.preventDefault();
        const slide = gallery.currSlide;
        const fit = fitLevels.get(slide.index);
        if (!fit) return;
        const factor = [2, 4, 1][zoomStep];
        zoomStep = (zoomStep + 1) % 3;
        slide.zoomTo(fit * factor, { x: doubleClick.clientX, y: doubleClick.clientY }, 120);
      });
    } finally {
      this.opening = false;
    }
  }
}
