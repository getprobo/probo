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
