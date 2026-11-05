# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Guests comparing cabins and looking for an available day-use or overnight stay.
- Cabin owners managing their cabins, availability, and external booking blocks.
- A Travelando administrator managing owners, cabins, inquiries, reservations, and commissions.

## Product Purpose

Travelando helps guests find available cabins and contact their owner through WhatsApp, while giving Travelando a private way to follow inquiries through reservation outcomes. Owners can add cabins, maintain availability, and report external bookings. Success means guests find a suitable available stay, reach its owner, and reservations, external blocks, and commission payments remain distinct.

## Positioning

Availability search is connected to WhatsApp inquiries and a private follow-up flow. Owners use cédula-and-PIN access limited to their assigned cabins; each cabin inherits its owner's WhatsApp contact.

## Operating Context

- Spanish-language catalog with availability search by dates, stay type (pasadía or estadía), and guest count. Cabin settings define each stay's entry and exit times; guests do not select hours.
- Guest requests open a prefilled WhatsApp message to the cabin contact.
- The private admin manages owners, assigned cabins, inquiry outcomes, full commission payments, and reservation dates.
- Owners sign in with their cédula and six-digit PIN, can create cabins published immediately, and can toggle assigned cabins active or inactive.
- The administrator configures one global commission rate for all new reservations.
- When SMTP is configured, confirmed reservations can trigger an email to the guest.

## Capabilities and Constraints

- Prices are represented in Colombian pesos (COP); Colombia is implied by the currency, but exact service areas are undecided.
- Each cabin belongs to an owner; its public contact name and WhatsApp come from that owner. Owners can access only assigned cabins and reservations, including their guests' contact details; other owners' reservations remain private.
- Admin can deactivate an owner without deleting their history; their cabins stay stored and hidden until the owner is reactivated.
- Confirmed reservations occupy their configured time range plus each cabin's cleaning hours (default 3, configurable down to 0); manual admin blocks and owner-reported external bookings block full dates. External bookings do not appear as Travelando reservations and do not incur commission.
- Commission rate is fixed per reservation. Reservation edits recalculate commission at that rate; post-payment changes are recorded as adjustments, yielding an outstanding balance or owner credit.
- Specific accessibility requirements and standards have not been established.

## Brand Commitments

- Product name: Travelando.
- Spanish is the product language.
- WhatsApp is the established cabin-contact channel.

## Evidence on Hand

- Working Go web application and source for the guest catalog, inquiry flow, private admin, and reservation follow-up.
- No verified testimonials, customer counts, performance claims, or marketing case studies identified; do not invent them.

## Product Principles

- Show availability for the guest's actual dates and stay type before sending an inquiry.
- Keep the path from cabin discovery to its contact direct through WhatsApp.
- Make inquiry outcomes and confirmed reservations trackable without requiring guest accounts.
