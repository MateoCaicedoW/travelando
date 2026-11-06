import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['sidebar', 'backdrop', 'content', 'mobileButton', 'headerRow', 'collapseRow', 'collapseButton', 'collapseIcon', 'expandIcon', 'brandLabel', 'portalSection', 'navigationLink', 'linkLabel', 'logoutButton', 'logoutLabel', 'accountActions'];

  connect() {
    this.handleKeydown = this.handleKeydown.bind(this);
    this.handleBreakpointChange = this.handleBreakpointChange.bind(this);
    this.desktopBreakpoint = window.matchMedia('(min-width: 1024px)');
    this.desktopBreakpoint.addEventListener('change', this.handleBreakpointChange);
    document.addEventListener('keydown', this.handleKeydown);
    this.closeMobile();
  }

  disconnect() {
    this.desktopBreakpoint.removeEventListener('change', this.handleBreakpointChange);
    document.removeEventListener('keydown', this.handleKeydown);
  }

  toggleCompact() {
    const compact = this.sidebarTarget.dataset.compact !== 'true';
    this.sidebarTarget.dataset.compact = String(compact);
    this.sidebarTarget.classList.toggle('w-20', compact);
    this.sidebarTarget.classList.toggle('w-64', !compact);
    this.headerRowTarget.classList.toggle('flex-col', compact);
    this.headerRowTarget.classList.toggle('justify-center', compact);
    this.headerRowTarget.classList.toggle('justify-between', !compact);
    this.collapseRowTarget.classList.toggle('justify-center', compact);
    this.collapseRowTarget.classList.toggle('justify-end', !compact);
    this.brandLabelTarget.hidden = compact;
    if (this.hasPortalSectionTarget) this.portalSectionTarget.hidden = compact;
    if (this.hasLogoutLabelTarget) this.logoutLabelTarget.hidden = compact;
    if (this.hasLogoutButtonTarget) this.logoutButtonTarget.classList.toggle('justify-center', compact);
    if (this.hasAccountActionsTarget) this.accountActionsTarget.classList.toggle('flex-col', compact);
    this.linkLabelTargets.forEach((label) => { label.hidden = compact; });
    this.navigationLinkTargets.forEach((link) => link.classList.toggle('justify-center', compact));
    this.collapseIconTarget.classList.toggle('hidden', compact);
    this.expandIconTarget.classList.toggle('hidden', !compact);
    this.collapseButtonTarget.setAttribute('aria-label', compact ? 'Expandir navegación' : 'Contraer navegación');
    this.collapseButtonTarget.title = compact ? 'Expandir navegación' : 'Contraer navegación';
  }

  toggleMobile() {
    if (this.sidebarTarget.classList.contains('hidden')) {
      this.openMobile();
    } else {
      this.closeMobile();
    }
  }

  openMobile() {
    this.sidebarTarget.classList.remove('hidden');
    this.backdropTarget.classList.remove('hidden');
    this.contentTarget.inert = true;
    this.mobileButtonTarget.setAttribute('aria-expanded', 'true');
    this.sidebarTarget.querySelector('a')?.focus();
  }

  closeMobile() {
    const restoreFocus = !this.desktopBreakpoint.matches && (this.sidebarTarget.contains(document.activeElement) || this.backdropTarget === document.activeElement);
    this.sidebarTarget.classList.add('hidden');
    this.backdropTarget.classList.add('hidden');
    this.contentTarget.inert = false;
    this.mobileButtonTarget.setAttribute('aria-expanded', 'false');
    if (restoreFocus) this.mobileButtonTarget.focus();
  }

  handleKeydown(event) {
    if (this.sidebarTarget.classList.contains('hidden') || this.desktopBreakpoint.matches) return;
    if (event.key === 'Escape') {
      this.closeMobile();
      return;
    }
    if (event.key !== 'Tab') return;
    const focusable = Array.from(this.sidebarTarget.querySelectorAll('a, button:not([disabled])'))
      .filter((element) => element.offsetParent !== null);
    if (focusable.length === 0) return;
    if (event.shiftKey && document.activeElement === focusable[0]) {
      event.preventDefault();
      focusable.at(-1).focus();
    } else if (!event.shiftKey && document.activeElement === focusable.at(-1)) {
      event.preventDefault();
      focusable[0].focus();
    }
  }

  handleBreakpointChange(event) {
    if (event.matches) this.closeMobile();
  }
}
