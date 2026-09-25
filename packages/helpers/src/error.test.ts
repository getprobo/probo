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

import { formatError, graphqlErrorField, toFieldErrors } from "./error";

describe("graphqlErrorField", () => {
    it("reads the validation field from extensions", () => {
        expect(graphqlErrorField({
            message: "too long",
            extensions: { code: "INVALID", field: "name" },
        })).toBe("name");
    });

    it("returns undefined without a field extension", () => {
        expect(graphqlErrorField({ message: "boom", extensions: { code: "FORBIDDEN" } })).toBeUndefined();
        expect(graphqlErrorField(new Error("boom"))).toBeUndefined();
        expect(graphqlErrorField(null)).toBeUndefined();
    });
});

describe("toFieldErrors", () => {
    it("maps payload errors keyed by field", () => {
        expect(toFieldErrors([
            { message: "too long", extensions: { field: "name" } },
            { message: "required", extensions: { field: "logo_file" } },
        ])).toEqual({
            name: "too long",
            logo_file: "required",
        });
    });

    it("accepts a single error", () => {
        expect(toFieldErrors({
            message: "too long",
            extensions: { field: "name" },
        })).toEqual({ name: "too long" });
    });

    it("ignores errors without a field", () => {
        expect(toFieldErrors([
            { message: "internal", extensions: { code: "INTERNAL_SERVER_ERROR" } },
        ])).toBeUndefined();
        expect(toFieldErrors(null)).toBeUndefined();
    });

    it("maps field errors nested under source.errors", () => {
        expect(toFieldErrors({
            message: "Unexpected error",
            source: {
                errors: [
                    { message: "too long", extensions: { code: "INVALID", field: "name" } },
                ],
            },
        })).toEqual({ name: "too long" });
    });
});

describe("formatError", () => {
    it("joins GraphQL messages onto the title", () => {
        expect(formatError("Update failed", { message: "too long" })).toBe("Update failed: too long.");
    });
});
