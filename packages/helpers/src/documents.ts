// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

type Translator = (s: string) => string;

export const documentTypes = ["OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE", "STATEMENT_OF_APPLICABILITY", "DATA_LIST"] as const;

export function getDocumentTypeLabel(__: Translator, type: string) {
    switch (type) {
        case "OTHER":
            return __("Other");
        case "GOVERNANCE":
            return __("Governance");
        case "POLICY":
            return __("Policy");
        case "PROCEDURE":
            return __("Procedure");
        case "PLAN":
            return __("Plan");
        case "REGISTER":
            return __("Register");
        case "RECORD":
            return __("Record");
        case "REPORT":
            return __("Report");
        case "TEMPLATE":
            return __("Template");
        case "STATEMENT_OF_APPLICABILITY":
            return __("Statement of Applicability");
        case "DATA_LIST":
            return __("Data List");
    }
}

export const documentWriteModes = ["AUTHORED", "GENERATED"] as const;

export function getDocumentWriteModeLabel(__: Translator, writeMode: string) {
    switch (writeMode) {
        case "AUTHORED":
            return __("Authored");
        case "GENERATED":
            return __("Generated");
    }
}

export const documentClassifications = [
    "PUBLIC",
    "INTERNAL",
    "CONFIDENTIAL",
    "SECRET",
] as const;

export function getDocumentClassificationLabel(__: Translator, classification: string) {
    switch (classification) {
        case "PUBLIC":
            return __("Public");
        case "INTERNAL":
            return __("Internal");
        case "CONFIDENTIAL":
            return __("Confidential");
        case "SECRET":
            return __("Secret");
    }
}
