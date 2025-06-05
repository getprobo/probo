import { describe, it, expect } from "vitest";
import { isEmpty } from "./array";

describe("array", () => {
    it("should check if an array is empty", () => {
        expect(isEmpty([])).toBe(true);
        expect(isEmpty([1, 2, 3])).toBe(false);
        expect(isEmpty(null)).toBe(true);
        expect(isEmpty(undefined)).toBe(true);
        expect(isEmpty([[], false])).toBe(true);
    });
});
