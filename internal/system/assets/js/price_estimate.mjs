const maxAmountCOP = 1_000_000_000_000;

export function estimatePrice(input) {
  const guests = Number(input.guests);
  const { includedGuests, extraGuestPrice, dayPrice, nightPrice } = input;
  if (
    !Number.isSafeInteger(guests) || guests < 1 ||
    !Number.isSafeInteger(includedGuests) || includedGuests < 1 ||
    !Number.isSafeInteger(extraGuestPrice) || extraGuestPrice < 0 || extraGuestPrice > maxAmountCOP ||
    !Number.isSafeInteger(dayPrice) || dayPrice < 0 || dayPrice > maxAmountCOP ||
    !Number.isSafeInteger(nightPrice) || nightPrice < 0 || nightPrice > maxAmountCOP
  ) return null;

  let unit = 'day';
  let units = 1;
  let basePrice = dayPrice;
  if (input.stayType === 'overnight') {
    unit = 'night';
    basePrice = nightPrice;
    if (input.dateFrom && input.dateTo) {
      const start = parseDate(input.dateFrom);
      const end = parseDate(input.dateTo);
      if (start === null || end === null) return null;
      units = (end - start) / 86_400_000;
      if (!Number.isInteger(units) || units < 1 || units > 365) return null;
      unit = 'stay';
    }
  } else if (input.stayType !== 'day_use') {
    return null;
  }

  const extraGuests = Math.max(0, guests - includedGuests);
  if (basePrice > Math.floor(maxAmountCOP / units)) return null;
  const baseTotal = basePrice * units;
  if (extraGuests > 0 && extraGuestPrice > 0 && extraGuests > Math.floor((maxAmountCOP - baseTotal) / extraGuestPrice)) return null;
  const total = baseTotal + extraGuests * extraGuestPrice;
  if (!Number.isSafeInteger(total) || total > maxAmountCOP) return null;
  return { total, unit };
}

function parseDate(value) {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) return null;
  const timestamp = Date.UTC(Number(match[1]), Number(match[2]) - 1, Number(match[3]));
  const date = new Date(timestamp);
  if (date.getUTCFullYear() !== Number(match[1]) || date.getUTCMonth() !== Number(match[2]) - 1 || date.getUTCDate() !== Number(match[3])) return null;
  return timestamp;
}
