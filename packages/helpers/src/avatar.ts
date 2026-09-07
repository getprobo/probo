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

export const avatarColors = [
    "gold",
    "red",
    "green",
    "amber",
    "sky",
    "indigo",
] as const;

type AvatarColor = (typeof avatarColors)[number];

export function avatarInitial(name: string): string {
    const trimmed = name.trim();
    if (trimmed.length === 0) {
        return "?";
    }

    return (Array.from(trimmed)[0] ?? "?").toUpperCase();
}

export function avatarColor(seed: string): AvatarColor {
    let hash = 0;
    for (const unit of Array.from(seed)) {
        hash = Math.imul(hash, 31) + (unit.codePointAt(0) ?? 0);
    }

    const index = Math.abs(hash) % avatarColors.length;

    return avatarColors[index];
}
