import { describe, it, expect } from "vitest";
import { isEmpty } from "./array";
import { fileSize } from "./file";

describe("file", () => {
    it("should display a file size in a human readable format", () => {
        const fakeTranslator = (s: string) => s;
        expect(fileSize(fakeTranslator, 4911)).toMatchInlineSnapshot(
            `"4.8 KB"`,
        );
        expect(fileSize(fakeTranslator, 20)).toMatchInlineSnapshot(`"20 B"`);
    });
});
