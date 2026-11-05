import { Controller } from '@hotwired/stimulus';
import Swal from 'sweetalert2';

export default class extends Controller {
  async confirm(event) {
    if (!event.detail.question) return;
    event.preventDefault();
    const result = await this.show(event.detail.question);
    if (result.isConfirmed) event.detail.issueRequest(true);
  }

  async submit(event) {
    const form = event.target;
    const question = form.dataset.confirmationMessage;
    if (!question) return;

    event.preventDefault();
    const result = await this.show(question);
    if (result.isConfirmed) form.submit();
  }

  show(question) {
    return Swal.fire({
      title: 'Confirmar acción',
      text: question,
      icon: 'warning',
      showCancelButton: true,
      reverseButtons: true,
      focusCancel: true,
      buttonsStyling: false,
      confirmButtonText: 'Confirmar',
      cancelButtonText: 'Cancelar',
      customClass: {
        actions: 'gap-4',
        popup: 'rounded-xl',
        title: 'text-ink',
        htmlContainer: 'text-muted',
        confirmButton: 'rounded-lg bg-brand-dark px-5 py-2.5 text-sm font-bold text-white hover:brightness-90',
        cancelButton: 'rounded-lg border border-line bg-white px-5 py-2.5 text-sm font-bold text-ink hover:bg-cream',
      },
    });
  }
}
