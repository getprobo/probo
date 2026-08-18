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

import { describe, expect, it } from "vitest";
import { getRiskScoreLevel } from "./risk";

describe("getRiskScoreLevel", () => {
    it("keeps the 5×5 bands when no matrix size is given", () => {
        expect(getRiskScoreLevel(4)).toBe(0);
        expect(getRiskScoreLevel(5)).toBe(1);
        expect(getRiskScoreLevel(14)).toBe(1);
        expect(getRiskScoreLevel(15)).toBe(2);
        expect(getRiskScoreLevel(25)).toBe(2);
    });

    it("scales bands to a 3×3 matrix", () => {
        const matrixSize = { rows: 3, cols: 3 };
        expect(getRiskScoreLevel(1, matrixSize)).toBe(0);
        expect(getRiskScoreLevel(2, matrixSize)).toBe(1);
        expect(getRiskScoreLevel(5, matrixSize)).toBe(1);
        expect(getRiskScoreLevel(6, matrixSize)).toBe(2);
        expect(getRiskScoreLevel(9, matrixSize)).toBe(2);
    });

    it("scales bands to a 4×4 matrix", () => {
        const matrixSize = { rows: 4, cols: 4 };
        expect(getRiskScoreLevel(3, matrixSize)).toBe(0);
        expect(getRiskScoreLevel(4, matrixSize)).toBe(1);
        expect(getRiskScoreLevel(9, matrixSize)).toBe(1);
        expect(getRiskScoreLevel(10, matrixSize)).toBe(2);
        expect(getRiskScoreLevel(16, matrixSize)).toBe(2);
    });
});
