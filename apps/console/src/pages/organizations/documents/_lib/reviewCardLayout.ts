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

export type ReviewCardSize = {
  key: string;
  anchor: number;
  height: number;
};

export const reviewCardGap = 8;

// placeReviewCards keeps cards in anchor order without overlap. The active
// card sits on its anchor when it can, and the cards above it move up.
export function placeReviewCards(
  cards: ReadonlyArray<ReviewCardSize>,
  activeKey: string | null,
): Record<string, number> {
  const sorted = [...cards].sort((left, right) =>
    left.anchor - right.anchor || left.key.localeCompare(right.key),
  );
  const tops = sorted.map(card => card.anchor);
  const active = sorted.findIndex(card => card.key === activeKey);

  if (active >= 0) {
    let ceiling = sorted[active].anchor - reviewCardGap;
    for (let index = active - 1; index >= 0; index -= 1) {
      const card = sorted[index];
      tops[index] = Math.min(card.anchor, ceiling - card.height);
      ceiling = tops[index] - reviewCardGap;
    }
  }

  let floor = 0;
  sorted.forEach((card, index) => {
    const preferred = index <= active ? tops[index] : card.anchor;
    tops[index] = Math.max(preferred, floor);
    floor = tops[index] + card.height + reviewCardGap;
  });

  return Object.fromEntries(sorted.map((card, index) => [card.key, tops[index]]));
}
