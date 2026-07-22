// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { CSSProperties, PointerEvent } from "react";

// Max backdrop translate (px) at the tracking surface edge. Negated against the
// pointer so the blur moves opposite the cursor, symmetric about center.
const BACKDROP_PARALLAX_PX = 12;

const BACKDROP_EASE_MS = 150;
const BACKDROP_EASE = `transform ${BACKDROP_EASE_MS}ms ease-out`;

// Set on the tracking surface while the pointer is over it.
const PARALLAX_ACTIVE = "data-backdrop-parallax";

// Kept until pointer movement has been idle for BACKDROP_EASE_MS after enter
// (and after each move during that window) so transition is not flipped to
// none while a retargeted ease is still interpolating.
const PARALLAX_EASING = "data-backdrop-parallax-easing";

const BACKDROP_VARIANT = "data-backdrop-variant";

const enterEaseTimeouts = new WeakMap<HTMLElement, number>();

function clearEaseEnd(surface: HTMLElement) {
  const previous = enterEaseTimeouts.get(surface);
  if (previous != null) {
    window.clearTimeout(previous);
    enterEaseTimeouts.delete(surface);
  }
  surface.removeAttribute(PARALLAX_EASING);
}

// (Re)start the idle timer that ends eased tracking. Call on enter and on every
// move while still easing so a late retarget can finish before transition:none.
function scheduleEaseEnd(surface: HTMLElement) {
  surface.setAttribute(PARALLAX_EASING, "");
  const previous = enterEaseTimeouts.get(surface);
  if (previous != null) {
    window.clearTimeout(previous);
  }
  enterEaseTimeouts.set(
    surface,
    window.setTimeout(() => {
      surface.removeAttribute(PARALLAX_EASING);
      enterEaseTimeouts.delete(surface);
    }, BACKDROP_EASE_MS),
  );
}

// Per-surface Figma specs. Both use a width-based square centered on the frame
// (top/left 50% + translate -50%). Logo Tile overflows (~inset -42px on a
// ~157px media → scale 1.53); Subprocessor Card fills the header width at 1×.
export type BlurBackdropVariant = "logoTile" | "card";

const blurBackdropByVariant = {
  // Figma "Logo Card" Background Image: opacity 6%, blur 8px, inset ≈ -42px.
  logoTile: {
    className:
      "pointer-events-none absolute top-1/2 left-1/2 aspect-square w-full object-cover opacity-[0.06] blur-[8px]",
    scale: 1.53,
  },
  // Figma "Subprocessor Card" Background Image: opacity 10%, blur 14px, width square.
  card: {
    className:
      "pointer-events-none absolute top-1/2 left-1/2 aspect-square w-full object-cover opacity-10 blur-[14px]",
    scale: 1,
  },
} as const;

export function blurBackdropClassName(variant: BlurBackdropVariant): string {
  return blurBackdropByVariant[variant].className;
}

export function blurBackdropTransformRest(variant: BlurBackdropVariant): string {
  const { scale } = blurBackdropByVariant[variant];
  return scale === 1
    ? "translate(-50%, -50%)"
    : `translate(-50%, -50%) scale(${scale})`;
}

export function blurBackdropStyle(variant: BlurBackdropVariant): CSSProperties {
  return { transform: blurBackdropTransformRest(variant) };
}

// Mark the overflow-hidden media/header that owns the blur image, and which
// Figma recipe the parallax transform should preserve (scale).
export function backdropFrameProps(variant: BlurBackdropVariant) {
  return {
    "data-backdrop-frame": "",
    [BACKDROP_VARIANT]: variant,
  } as const;
}

function prefersReducedMotion(): boolean {
  return typeof window !== "undefined"
    && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

function backdropFrame(scope: HTMLElement): HTMLElement | null {
  if (scope.hasAttribute("data-backdrop-frame")) {
    return scope;
  }

  return scope.querySelector<HTMLElement>("[data-backdrop-frame]");
}

function backdropImage(frame: HTMLElement): HTMLElement | null {
  return frame.querySelector<HTMLElement>(":scope > img[aria-hidden]");
}

function backdropScale(frame: HTMLElement): number {
  const variant = frame.getAttribute(BACKDROP_VARIANT);
  if (variant === "logoTile" || variant === "card") {
    return blurBackdropByVariant[variant].scale;
  }

  return 1;
}

function backdropTransform(x: number, y: number, scale: number): string {
  const translate = `translate(calc(-50% + ${x}px), calc(-50% + ${y}px))`;
  return scale === 1 ? translate : `${translate} scale(${scale})`;
}

// Drive the blur from pointer position over the tracking surface (the whole
// card). X and Y are independent axes against that surface's bounds, so a
// left→right swipe at the top matches one at the bottom, and top→bottom gets
// a full vertical range. Attach listeners to the card; mark the blur's frame
// with backdropFrameProps(variant).
export function onBackdropPointerMove(event: PointerEvent<HTMLElement>) {
  if (prefersReducedMotion()) {
    return;
  }

  const surface = event.currentTarget;
  const frame = backdropFrame(surface);
  if (frame == null) {
    return;
  }

  const image = backdropImage(frame);
  if (image == null) {
    return;
  }

  const rect = surface.getBoundingClientRect();
  const nx = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  const ny = ((event.clientY - rect.top) / rect.height) * 2 - 1;
  const nextTransform = backdropTransform(
    -nx * BACKDROP_PARALLAX_PX,
    -ny * BACKDROP_PARALLAX_PX,
    backdropScale(frame),
  );

  const entering = !surface.hasAttribute(PARALLAX_ACTIVE);
  const easing = entering || surface.hasAttribute(PARALLAX_EASING);
  surface.setAttribute(PARALLAX_ACTIVE, "");

  if (easing) {
    // Apply transition before changing transform so the browser interpolates
    // from the frozen pose (same-frame transition+transform often snaps).
    image.style.transition = BACKDROP_EASE;
    if (entering) {
      void image.offsetWidth;
    }
    scheduleEaseEnd(surface);
  } else {
    image.style.transition = "none";
  }

  image.style.transform = nextTransform;
}

// Freeze the blur at its visible pose. Clearing flags alone is not enough:
// during enter-ease, transition is still 150ms toward the last target, so the
// image would keep sliding after leave unless we capture the computed
// transform and disable the transition.
export function onBackdropPointerLeave(event: PointerEvent<HTMLElement>) {
  const surface = event.currentTarget;
  clearEaseEnd(surface);
  surface.removeAttribute(PARALLAX_ACTIVE);

  const frame = backdropFrame(surface);
  if (frame == null) {
    return;
  }

  const image = backdropImage(frame);
  if (image == null) {
    return;
  }

  const computed = getComputedStyle(image).transform;
  image.style.transition = "none";
  if (computed !== "none") {
    image.style.transform = computed;
  }
}
