---
name: Travelando
description: A warm, direct visual system for finding a cabin escape.
colors:
  trail-orange: "#ff5d06"
  burnt-trail: "#c24200"
  warm-graphite: "#004928"
  soft-olive-gray: "#3d624b"
  cabin-cream: "#fffdeb"
  sandstone-line: "#d5dfd5"
  pine-green: "#004928"
  pine-mist: "#e8f1e9"
  white: "#ffffff"
typography:
  display:
    fontFamily: '"Barlow Semi Condensed", sans-serif'
    fontSize: "3.75rem"
    fontWeight: 900
    lineHeight: 1.06
    letterSpacing: "-0.04em"
    fontStyle: italic
  headline:
    fontFamily: '"Barlow Semi Condensed", sans-serif'
    fontSize: "2.25rem"
    fontWeight: 800
    lineHeight: 1.2
    fontStyle: italic
  title:
    fontFamily: '"Barlow Semi Condensed", sans-serif'
    fontSize: "1.25rem"
    fontWeight: 800
    lineHeight: 1.4
    fontStyle: italic
  body:
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", "Noto Sans", Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"'
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", "Noto Sans", Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"'
    fontSize: "0.75rem"
    fontWeight: 800
    lineHeight: 1.5
    letterSpacing: "0.16em"
rounded:
  lg: "8px"
  xl: "12px"
  2xl: "16px"
  full: "9999px"
spacing:
  1: "0.25rem"
  2: "0.5rem"
  3: "0.75rem"
  4: "1rem"
  5: "1.25rem"
  6: "1.5rem"
  8: "2rem"
  12: "3rem"
components:
  button-primary:
    backgroundColor: "{colors.burnt-trail}"
    textColor: "{colors.white}"
    rounded: "{rounded.lg}"
    height: "48px"
    padding: "12px 20px"
  button-strong:
    backgroundColor: "{colors.burnt-trail}"
    textColor: "{colors.white}"
    rounded: "{rounded.lg}"
    height: "48px"
    padding: "12px 20px"
  button-secondary:
    backgroundColor: "{colors.white}"
    textColor: "{colors.warm-graphite}"
    rounded: "{rounded.lg}"
    height: "44px"
    padding: "8px 16px"
  input:
    backgroundColor: "{colors.white}"
    textColor: "{colors.warm-graphite}"
    rounded: "{rounded.xl}"
    padding: "12px 16px"
  card:
    backgroundColor: "{colors.white}"
    rounded: "{rounded.2xl}"
    padding: "20px"
  chip-success:
    backgroundColor: "{colors.pine-green}"
    textColor: "{colors.white}"
    rounded: "{rounded.full}"
    padding: "4px 10px"
---

# Design System: Travelando

## Overview

**Creative North Star: "Naranja en el sendero"**

Travelando feels like a friendly, practical starting point for finding an escape. Bright orange, deep forest green, and pale cream make the brand unmistakable; oblique, condensed headings carry the character of the identity while body text stays familiar and readable.

The guest catalog is welcoming and image-led where cabin photography exists. Administrative screens reuse the same surfaces and controls, with denser tables and lists that prioritize scanability. Pine green marks successful or confirmed states. The visual mood is warm and direct, not a polished luxury-resort aesthetic.

**Key Characteristics:**
- Pale cream surfaces, orange brand/actions, and forest-green text.
- Oblique, extra-bold condensed headings with neutral sans-serif body text.
- Rounded, thin-bordered cards with restrained depth.
- Responsive, utility-led layouts shared by guest and admin pages.

## Colors

The palette follows the supplied identity: bright orange `#ff5d06`, forest green `#004928`, and cream `#fffdeb`.

### Primary
- **Trail Orange:** Primary brand color, brand mark, and discovery accents.
- **Burnt Trail:** High-contrast primary actions and emphasis text.

### Secondary
- **Pine Green:** Confirmed and successful states.
- **Pine Mist:** Soft success notices.

### Neutral
- **Warm Graphite:** Deep forest-green primary text and important values.
- **Soft Olive Gray:** Supporting green copy and secondary labels.
- **Cabin Cream:** Warm page bands, badges, and low-emphasis surfaces.
- **Sandstone Line:** Card, field, and section borders.
- **White:** Card and input surfaces, plus text on filled actions.

**The Orange-Action Rule.** Trail Orange leads the identity; Burnt Trail carries primary actions with white text for accessible contrast. Forest Green supports typography and confirmed states.

## Typography

**Display Font:** Barlow Semi Condensed, italic 800/900 (SIL Open Font License 1.1), used for headings and wordmark.
**Body Font:** System sans-serif stack (`-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica Neue, Noto Sans, Arial, sans-serif`)
**Label/Mono Font:** No distinct label or mono family; labels use the body stack.

**Character:** Bold, oblique condensed headings echo the supplied Tablet Gothic specimen. Body text stays neutral and comfortably spaced; small uppercase labels add hierarchy without becoming ornamental.

### Hierarchy
- **Display** (black weight, responsive 36–60px, tight 1.06 line-height): Homepage hero only.
- **Headline** (extra-bold, 30–36px): Page and section headings.
- **Title** (extra-bold, 18–24px): Card names and section titles.
- **Body** (regular, 16px, 1.5 line-height): Descriptions and form content; prose uses a readable measure.
- **Label** (extra-bold, 12px, tracked and often uppercase): Field labels, metadata, and section eyebrows.

## Layout

Pages center within a wide 80rem content container. Mobile gutters are 1.25rem and widen to 2rem from the small breakpoint (640px). The catalog hero and cabin detail switch to split layouts at large screens (1024px); cabin results expand from one column to two at medium screens (768px), then three at extra-large screens (1280px). Admin forms and cards use the same spacing rhythm, while wide tables scroll horizontally rather than compressing their columns. Common section padding grows from 3rem to 4rem at small screens.

## Elevation & Depth

Depth is mostly structural: white surfaces sit on white or cream backgrounds, separated by sandstone borders. A small shadow supports the inquiry report; cabin cards lift slightly and gain a medium shadow on hover. Form focus uses an orange-tinted ring rather than elevation.

**The Border-First Rule.** Keep cards flat at rest; reserve shadow for hover feedback or the report surface that needs a clearer boundary.

## Shapes

Cards use generously rounded corners; buttons and fields use smaller rounded corners, while badges and the compact brand tile use pill or compact rounded silhouettes. Thin neutral borders define surfaces. Images are clipped to card corners and use wide, landscape ratios.

## Components

### Buttons
- **Character:** Direct and sturdy, with bold labels and clear contrast.
- **Shape:** Gently rounded; primary buttons use the `lg` radius token.
- **Primary:** Burnt Trail fill with white text for high-emphasis actions; see `button-primary` token.
- **Accent:** Trail Orange leads the brand mark, focus, and discovery details.
- **Hover / Focus:** The fill darkens on hover. Keyboard focus gets a visible, offset outline in the deep orange.
- **Secondary:** White or transparent surface, sandstone border, graphite text; hover shifts the border toward orange.

### Chips
- **Style:** Compact pills with cream/orange for neutral status and pine/white for confirmed status.
- **State:** Status communicates meaning through both label and color; success uses `chip-success`.

### Cards / Containers
- **Corner Style:** Generously rounded using the `2xl` radius token.
- **Background:** White, with cream reserved for highlighted bands and details.
- **Shadow Strategy:** Border-first; hover lift is limited to interactive cabin cards.
- **Border:** Thin sandstone line.
- **Internal Padding:** Commonly 1.25rem, increasing to 1.5rem on wider layouts.

### Inputs / Fields
- **Style:** White fill, sandstone stroke, softly rounded corners, and consistent horizontal padding.
- **Focus:** Orange border and subtle orange ring; keyboard focus remains visible.
- **Error / Disabled:** Errors use red text and pale red surfaces. No distinct disabled treatment is established.

### Navigation
- **Style:** A cream header and footer frame a centered content width. The brand pairs the orange Travelando icon with a bold wordmark; navigation links stay brief and collapse secondary exploration on narrow screens.
- **States:** Links are graphite at rest and shift toward deep orange on hover; the admin link uses a light bordered pill.

## Do's and Don'ts

### Do:
- **Do** use Trail Orange as the primary brand color and Burnt Trail for high-emphasis actions with white text.
- **Do** use Pine Green and Pine Mist for successful or confirmed states.
- **Do** preserve the shared card, form, and button language across guest and admin screens.
- **Do** let responsive grids add columns at established breakpoints rather than shrinking content too far.

### Don't:
- **Don't** turn the cabin catalog into a polished luxury-resort aesthetic.
- **Don't** add decorative colors or use display type for body copy.
- **Don't** use shadow as the default way to separate cards; borders and tonal surfaces do that work.
