import assert from 'node:assert/strict';
import test from 'node:test';
import Swal from 'sweetalert2';
import ConfirmationController from './confirmation_controller.js';

test('confirms HTMX requests through SweetAlert2', async () => {
  const originalFire = Swal.fire;
  const calls = [];
  Swal.fire = async (options) => {
    calls.push(options);
    return { isConfirmed: true };
  };

  try {
    const controller = Object.create(ConfirmationController.prototype);
    let prevented = false;
    let issued = false;
    await controller.confirm({
      preventDefault() { prevented = true; },
      detail: {
        question: '¿Cancelar reserva?',
        issueRequest(force) { issued = force; },
      },
    });

    assert.equal(prevented, true);
    assert.equal(issued, true);
    assert.equal(calls[0].text, '¿Cancelar reserva?');
    assert.equal(calls[0].showCancelButton, true);
  } finally {
    Swal.fire = originalFire;
  }
});

test('does not issue HTMX request after cancellation', async () => {
  const originalFire = Swal.fire;
  Swal.fire = async () => ({ isConfirmed: false });

  try {
    const controller = Object.create(ConfirmationController.prototype);
    let issued = false;
    await controller.confirm({
      preventDefault() {},
      detail: { question: '¿Eliminar?', issueRequest() { issued = true; } },
    });
    assert.equal(issued, false);
  } finally {
    Swal.fire = originalFire;
  }
});

test('leaves HTMX requests without a confirmation question untouched', async () => {
  const originalFire = Swal.fire;
  let shown = false;
  Swal.fire = async () => {
    shown = true;
    return { isConfirmed: true };
  };

  try {
    const controller = Object.create(ConfirmationController.prototype);
    let prevented = false;
    await controller.confirm({
      preventDefault() { prevented = true; },
      detail: { issueRequest() {} },
    });
    assert.equal(prevented, false);
    assert.equal(shown, false);
  } finally {
    Swal.fire = originalFire;
  }
});

test('submits native forms only after confirmation', async () => {
  const originalFire = Swal.fire;
  const form = { dataset: { confirmationMessage: '¿Eliminar amenidad?' }, submit() { this.submitted = true; } };
  Swal.fire = async (options) => {
    assert.equal(options.text, form.dataset.confirmationMessage);
    return { isConfirmed: true };
  };

  try {
    const controller = Object.create(ConfirmationController.prototype);
    let prevented = false;
    await controller.submit({ target: form, preventDefault() { prevented = true; } });
    assert.equal(prevented, true);
    assert.equal(form.submitted, true);
  } finally {
    Swal.fire = originalFire;
  }
});
