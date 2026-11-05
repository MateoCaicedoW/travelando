import assert from 'node:assert/strict';
import test from 'node:test';
import AdminNavigationController from './admin_navigation_controller.js';

function fakeClassList(...initial) {
  const values = new Set(initial);
  return {
    add: (value) => values.add(value),
    remove: (value) => values.delete(value),
    contains: (value) => values.has(value),
    toggle(value, force = !values.has(value)) {
      if (force) values.add(value);
      else values.delete(value);
      return force;
    },
  };
}

test('collapses admin sidebar to labeled icon links and expands it again', () => {
  const controller = Object.create(AdminNavigationController.prototype);
  const labels = [{}, {}];
  const links = [{ classList: fakeClassList() }, { classList: fakeClassList() }];
  controller.collapseRowTarget = { classList: fakeClassList('justify-end') };
  controller.sidebarTarget = { dataset: {}, classList: fakeClassList('w-64') };
  controller.brandLabelTarget = {};
  controller.portalSectionTarget = {};
  controller.logoutLabelTarget = {};
  controller.logoutButtonTarget = { classList: fakeClassList() };
  controller.collapseIconTarget = { classList: fakeClassList() };
  controller.expandIconTarget = { classList: fakeClassList('hidden') };
  controller.collapseButtonTarget = { setAttribute(_name, value) { this.label = value; } };
  Object.defineProperties(controller, {
    linkLabelTargets: { value: labels },
    linkTargets: { value: links },
  });

  controller.toggleCompact();
  assert.equal(controller.sidebarTarget.dataset.compact, 'true');
  assert.equal(controller.sidebarTarget.classList.contains('w-20'), true);
  assert.equal(labels.every((label) => label.hidden), true);
  assert.equal(controller.collapseRowTarget.classList.contains('justify-center'), true);
  assert.equal(links.every((link) => link.classList.contains('justify-center')), true);
  assert.equal(controller.collapseButtonTarget.label, 'Expandir navegación');

  controller.toggleCompact();
  assert.equal(controller.sidebarTarget.classList.contains('w-64'), true);
  assert.equal(labels.some((label) => label.hidden), false);
  assert.equal(controller.collapseRowTarget.classList.contains('justify-end'), true);
  assert.equal(controller.collapseButtonTarget.label, 'Contraer navegación');
});

test('opens mobile navigation, makes page inert, and restores focus on close', () => {
  const originalDocument = globalThis.document;
  const document = { activeElement: null };
  globalThis.document = document;
  try {
    const controller = Object.create(AdminNavigationController.prototype);
    const firstLink = { focus() { document.activeElement = this; } };
    controller.desktopBreakpoint = { matches: false };
    controller.sidebarTarget = {
      classList: fakeClassList('hidden'),
      contains: (element) => element === firstLink,
      querySelector: () => firstLink,
    };
    controller.backdropTarget = { classList: fakeClassList('hidden') };
    controller.contentTarget = { inert: false };
    controller.mobileButtonTarget = {
      setAttribute(_name, value) { this.expanded = value; },
      focus() { document.activeElement = this; },
    };

    controller.openMobile();
    assert.equal(controller.sidebarTarget.classList.contains('hidden'), false);
    assert.equal(controller.backdropTarget.classList.contains('hidden'), false);
    assert.equal(controller.contentTarget.inert, true);
    assert.equal(controller.mobileButtonTarget.expanded, 'true');
    assert.equal(document.activeElement, firstLink);

    controller.closeMobile();
    assert.equal(controller.sidebarTarget.classList.contains('hidden'), true);
    assert.equal(controller.contentTarget.inert, false);
    assert.equal(controller.mobileButtonTarget.expanded, 'false');
    assert.equal(document.activeElement, controller.mobileButtonTarget);
  } finally {
    if (originalDocument === undefined) delete globalThis.document;
    else globalThis.document = originalDocument;
  }
});
