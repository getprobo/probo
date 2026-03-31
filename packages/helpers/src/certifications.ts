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

export const certifications = {
    securityStandards: [
        "SOC 2",
        "ISO 27001",
        "HITRUST",
        "NIST",
        "SOC 2 Type 2",
        "SOC 2 Type 1",
    ],
    regulatoryLegal: ["HIPAA", "FERPA", "FISMA", "PIPEDA", "GDPR", "CCPA"],
    industrySpecific: ["FinTech", "MPAA", "GSMA"],
    internationalGov: ["FedRAMP", "ENS High", "IRAP", "CJIS"],
    custom: [] as string[],
} as const;

export const certificationCategoryLabel = (
    __: Translator,
    category: keyof typeof certifications,
) => {
    switch (category) {
        case "securityStandards":
            return __("Security Standards");
        case "regulatoryLegal":
            return __("Regulatory & Legal");
        case "industrySpecific":
            return __("Industry-Specific");
        case "internationalGov":
            return __("International & Government");
        default:
            return __("Custom certifications");
    }
};
