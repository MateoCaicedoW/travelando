import assert from 'node:assert/strict';
import test from 'node:test';
import { estimatePrice } from './price_estimate.mjs';

test('adds fixed deposit once to day-use and multi-night estimates', () => {
  const cabin = { guests: 2, includedGuests: 2, extraGuestPrice: 0, depositCOP: 50000, dayPrice: 100000, nightPrice: 200000 };
  assert.equal(estimatePrice({ ...cabin, stayType: 'day_use', dateFrom: '2026-11-15' }).total, 150000);
  assert.equal(estimatePrice({ ...cabin, stayType: 'overnight', dateFrom: '2026-11-15', dateTo: '2026-11-17' }).total, 450000);
});

test('rejects estimates above COP limit after adding deposit', () => {
  assert.equal(estimatePrice({
    stayType: 'day_use', guests: 1, includedGuests: 1, extraGuestPrice: 0,
    depositCOP: 1, dayPrice: 1_000_000_000_000, nightPrice: 0,
  }), null);
});
