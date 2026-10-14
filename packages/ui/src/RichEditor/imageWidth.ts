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

export const minImageWidth = 10;
export const imageWidthStep = 10;

const snapDistance = 3;

export function imageWidthPercent(value: string | null | undefined): number | null {
  if (value == null) {
    return null;
  }

  const match = /^(\d{1,3})%$/.exec(value.trim());
  if (match == null) {
    return null;
  }

  const width = Number(match[1]);
  if (!Number.isInteger(width) || width < 1 || width > 100) {
    return null;
  }

  return width;
}

export function dragImageWidth(input: {
  side: "left" | "right";
  clientX: number;
  startX: number;
  startPercent: number;
  slotLeft: number;
  slotRight: number;
  slotWidth: number;
  targets: readonly number[];
}): number {
  if (input.slotWidth <= 0) {
    return snapImageWidth(input.startPercent, input.targets);
  }

  const pastOuterEdge = input.side === "right"
    ? input.clientX > input.slotRight
    : input.clientX < input.slotLeft;
  if (pastOuterEdge) {
    return 100;
  }

  const dx = input.clientX - input.startX;
  const signed = input.side === "right" ? dx : -dx;

  return snapImageWidth(input.startPercent + (signed * 2 / input.slotWidth) * 100, input.targets);
}

export function keyboardImageWidth(input: {
  side: "left" | "right";
  key: string;
  width: number;
  targets: readonly number[];
}): number | null {
  const delta = keyboardImageDelta(input.side, input.key);
  if (delta == null) {
    return null;
  }

  return snapImageWidth(input.width + delta, input.targets);
}

function keyboardImageDelta(side: "left" | "right", key: string): number | null {
  if (key === "Enter" || key === " ") {
    return imageWidthStep;
  }

  if (key === "ArrowRight") {
    return side === "right" ? imageWidthStep : -imageWidthStep;
  }

  if (key === "ArrowLeft") {
    return side === "left" ? imageWidthStep : -imageWidthStep;
  }

  return null;
}

export function snapImageWidth(width: number, targets: readonly number[]): number {
  const clamped = Math.min(100, Math.max(minImageWidth, Math.round(width)));
  let snapped = clamped;
  let nearest = snapDistance + 1;

  for (const target of [...targets, 100]) {
    if (!Number.isInteger(target) || target < minImageWidth || target > 100) {
      continue;
    }

    const distance = Math.abs(clamped - target);
    if (distance > 0 && distance <= snapDistance && distance < nearest) {
      snapped = target;
      nearest = distance;
    }
  }

  return snapped;
}
