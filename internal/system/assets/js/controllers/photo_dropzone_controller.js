import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['input', 'preview', 'stack'];

  connect() {
    this.files = [];
    this.element.addEventListener('click', (event) => {
      if (event.target === this.inputTarget) return;
      if (event.target.closest('button')) return;
      this.inputTarget.click();
    });
    this.inputTarget.addEventListener('change', () => this.addFiles(this.inputTarget.files));
  }

  addFiles(fileList) {
    this.files = this.files.concat(Array.from(fileList || []));
    this.sync();
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

  render() {
    const target = this.hasStackTarget ? this.stackTarget : this.previewTarget;
    target.querySelectorAll('[data-new-photo]').forEach((el) => el.remove());
    this.files.forEach((file, index) => {
      const url = URL.createObjectURL(file);
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