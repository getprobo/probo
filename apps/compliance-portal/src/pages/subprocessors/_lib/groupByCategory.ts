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

export interface CategoryGroup<T> {
  category: string;
  nodes: T[];
}

// Groups subprocessor nodes into their (non-empty) categories, ordered by the
// localized category label. Presentational only — filtering happens server-side.
export function groupByCategory<T extends { readonly category: string }>(
  nodes: readonly T[],
  getLabel: (category: string) => string,
): CategoryGroup<T>[] {
  const groups = new Map<string, T[]>();

  for (const node of nodes) {
    const existing = groups.get(node.category);
    if (existing) {
      existing.push(node);
    } else {
      groups.set(node.category, [node]);
    }
  }

  return [...groups.entries()]
    .map(([category, groupNodes]) => ({ category, nodes: groupNodes }))
    .sort((a, b) => getLabel(a.category).localeCompare(getLabel(b.category)));
}
