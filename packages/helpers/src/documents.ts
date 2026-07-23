// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type Translator = (s: string) => string;

export const documentTypes = ["OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE", "STATEMENT_OF_APPLICABILITY"] as const;

export function getDocumentTypeLabel(t: Translator, type: string) {
    switch (type) {
        case "OTHER":
            return t("helpers.documentType.other");
        case "GOVERNANCE":
            return t("helpers.documentType.governance");
        case "POLICY":
            return t("helpers.documentType.policy");
        case "PROCEDURE":
            return t("helpers.documentType.procedure");
        case "PLAN":
            return t("helpers.documentType.plan");
        case "REGISTER":
            return t("helpers.documentType.register");
        case "RECORD":
            return t("helpers.documentType.record");
        case "REPORT":
            return t("helpers.documentType.report");
        case "TEMPLATE":
            return t("helpers.documentType.template");
        case "STATEMENT_OF_APPLICABILITY":
            return t("helpers.documentType.statementOfApplicability");
    }
}

export const documentWriteModes = ["AUTHORED", "GENERATED"] as const;

export function getDocumentWriteModeLabel(t: Translator, writeMode: string) {
    switch (writeMode) {
        case "AUTHORED":
            return t("helpers.documentWriteMode.authored");
        case "GENERATED":
            return t("helpers.documentWriteMode.generated");
    }
}

export const documentClassifications = [
    "PUBLIC",
    "INTERNAL",
    "CONFIDENTIAL",
    "SECRET",
] as const;

export function getDocumentClassificationLabel(t: Translator, classification: string) {
    switch (classification) {
        case "PUBLIC":
            return t("helpers.documentClassification.public");
        case "INTERNAL":
            return t("helpers.documentClassification.internal");
        case "CONFIDENTIAL":
            return t("helpers.documentClassification.confidential");
        case "SECRET":
            return t("helpers.documentClassification.secret");
    }
}
