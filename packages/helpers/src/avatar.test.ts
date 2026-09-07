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

import { describe, it, expect } from "vitest";
import { avatarColor, avatarColors, avatarInitial } from "./avatar";

describe("avatarInitial", () => {
    it("uses the first letter", () => {
        expect(avatarInitial("Neytiri")).toBe("N");
        expect(avatarInitial("Jane Doe")).toBe("J");
    });

    it("uses the first code point for emoji", () => {
        expect(avatarInitial("😀 Smith")).toBe("😀");
    });

    it("returns a question mark for a blank name", () => {
        expect(avatarInitial("   ")).toBe("?");
    });
});

describe("avatarColor", () => {
    it("is stable for the same seed", () => {
        expect(avatarColor("identity-1")).toBe(avatarColor("identity-1"));
    });

    it("picks a palette color by modulo", () => {
        expect(avatarColors).toContain(avatarColor("identity-1"));
    });

    it("is stable when the display name is used as the seed", () => {
        expect(avatarColor("Neytiri")).toBe(avatarColor("Neytiri"));
    });
});
