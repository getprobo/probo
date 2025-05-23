import { describe, it, expect } from "vitest";
import { sprintf } from "./string";

describe("strings", () => {
    it("should format a string", () => {
        expect(sprintf("Hello %s", "world")).toBe("Hello world");
        expect(sprintf("Hello %s %s", "John", "Doe")).toBe("Hello John Doe");
        expect(sprintf("%2s %1s", "world", "Hello")).toBe("Hello world");
    });
});
