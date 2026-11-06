import { Controller } from '@hotwired/stimulus';

const MAX_PHOTOS = 5;
const MAX_BYTES = 8 * 1024 * 1024; // mismo límite que el backend
const MAX_DIMENSION = 1920;        // lado mayor en px
const JPEG_QUALITY = 0.82;
const SKIP_BELOW = 300 * 1024;     // si ya es chica y pequeña en px, no se toca

export default class extends Controller {
  static targets = ['input', 'preview', 'stack'];

  connect() {
    this.files = [];
    this.urls = [];
    this.pending = 0;
    this.onClick = (event) => {
      if (event.target === this.inputTarget) return;
      if (event.target.closest('button')) return;
      this.inputTarget.click();
    };
    this.element.addEventListener('click', this.onClick);
    this.inputTarget.addEventListener('change', () => this.addFiles(this.inputTarget.files));
  }

  disconnect() {
    this.element.removeEventListener('click', this.onClick);
    this.revokeUrls();
  }

  async addFiles(fileList) {
    const incoming = Array.from(fileList || []);
    if (incoming.length === 0) return;

    if (this.files.length + this.pending + incoming.length > MAX_PHOTOS) {
      this.setError('Solo se pueden subir 5 fotos a la vez.');
      return;
    }
    this.setError('');

    this.pending += incoming.length;
    this.element.setAttribute('aria-busy', 'true');
    try {
      const processed = await Promise.all(incoming.map((f) => this.compress(f)));
      const tooBig = processed.some((f) => f.size > MAX_BYTES);
      if (tooBig) {
        this.setError('Cada foto debe pesar máximo 8MB.');
        return;
      }
      this.files = this.files.concat(processed);
      this.sync();
    } finally {
      this.pending -= incoming.length;
      if (this.pending === 0) this.element.removeAttribute('aria-busy');
    }
  }

  async compress(file) {
    if (!file.type.startsWith('image/')) return file;
    if (file.type === 'image/gif' || file.type === 'image/svg+xml') return file;

    try {
      // 'from-image' aplica la orientación EXIF (fotos de celular rotadas)
      const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' });
      const scale = Math.min(1, MAX_DIMENSION / Math.max(bitmap.width, bitmap.height));

      if (scale === 1 && file.size < SKIP_BELOW) {
        bitmap.close();
        return file;
      }

      const width = Math.round(bitmap.width * scale);
      const height = Math.round(bitmap.height * scale);
      const canvas = document.createElement('canvas');
      canvas.width = width;
      canvas.height = height;

      const ctx = canvas.getContext('2d');
      ctx.fillStyle = '#fff'; // PNG con transparencia no queda negro en JPEG
      ctx.fillRect(0, 0, width, height);
      ctx.drawImage(bitmap, 0, 0, width, height);
      bitmap.close();

      const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/jpeg', JPEG_QUALITY));
      canvas.width = canvas.height = 0; // libera memoria del canvas

      // Si no mejoró, se queda el original
      if (!blob || blob.size >= file.size) return file;

      const name = file.name.replace(/\.[^.]+$/, '') + '.jpg';
      return new File([blob], name, { type: 'image/jpeg', lastModified: Date.now() });
    } catch {
      // Formato no decodificable en el navegador (ej. HEIC en Chrome): se sube tal cual
      return file;
    }
  }

  setError(text) {
    const msg = this.element.querySelector('.photo-dropzone-error');
    if (msg) msg.textContent = text;
  }

  remove(index) {
    this.files.splice(index, 1);
    this.sync();
  }

  dragover(event) {
    event.preventDefault();
    this.element.classList.add('border-brand');
  }

  dragleave() {
    this.element.classList.remove('border-brand');
  }

  drop(event) {
    event.preventDefault();
    this.element.classList.remove('border-brand');
    this.addFiles(event.dataTransfer.files);
  }

  sync() {
    const dt = new DataTransfer();
    this.files.forEach((file) => dt.items.add(file));
    this.inputTarget.files = dt.files;
    this.render();
  }

  revokeUrls() {
    this.urls.forEach((u) => URL.revokeObjectURL(u));
    this.urls = [];
  }

  render() {
    const target = this.hasStackTarget ? this.stackTarget : this.previewTarget;
    target.querySelectorAll('[data-new-photo]').forEach((el) => el.remove());
    this.revokeUrls(); // el original nunca liberaba los object URLs

    this.files.forEach((file, index) => {
      const url = URL.createObjectURL(file);
      this.urls.push(url);

      const wrapper = document.createElement('div');
      wrapper.className = 'relative';
      wrapper.dataset.newPhoto = 'true';

      const img = document.createElement('img');
      img.src = url;
      img.className = 'h-24 w-full rounded-lg object-cover';

      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'absolute right-1 top-1 rounded-full bg-black/60 px-2 py-1 text-xs text-white';
      button.textContent = '×';
      button.addEventListener('click', (event) => {
        event.stopPropagation();
        this.remove(index);
      });

      wrapper.appendChild(img);
      wrapper.appendChild(button);
      target.appendChild(wrapper);
    });
  }
}