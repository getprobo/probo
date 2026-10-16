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

import { taskPriorities, type TaskPriority } from "@probo/tasks";

export function resolveDropPriority(
  dragged: TaskPriority,
  above?: TaskPriority,
  below?: TaskPriority,
): TaskPriority | undefined {
  if (above === dragged || below === dragged) {
    return undefined;
  }

  const draggedIndex = taskPriorities.indexOf(dragged);
  if (above == null && below != null) {
    return taskPriorities.indexOf(below) <= draggedIndex ? below : undefined;
  }
  if (below == null && above != null) {
    return taskPriorities.indexOf(above) >= draggedIndex ? above : undefined;
  }
  if (above != null && below != null) {
    const distanceAbove = Math.abs(taskPriorities.indexOf(above) - draggedIndex);
    const distanceBelow = Math.abs(taskPriorities.indexOf(below) - draggedIndex);
    return distanceAbove <= distanceBelow ? above : below;
  }

  return undefined;
}

export function destinationRank(
  items: readonly { id: string; rank: number; priority: TaskPriority }[],
  order: readonly string[],
  newIndex: number,
  dragged: { id: string; rank: number; priority: TaskPriority },
  targetPriority: TaskPriority | undefined,
): { rank: number; send: boolean } {
  if (targetPriority == null) {
    return { rank: items[newIndex]?.rank ?? dragged.rank, send: true };
  }

  const nextId = order[newIndex + 1];
  const next = nextId == null ? undefined : items.find(item => item.id === nextId);
  if (next != null && next.priority === targetPriority) {
    return { rank: next.rank, send: true };
  }

  const peers = items.filter(item =>
    item.priority === targetPriority && item.id !== dragged.id,
  );
  const rank = peers.length === 0
    ? 1
    : Math.max(...peers.map(item => item.rank)) + 1;
  return { rank, send: false };
}
