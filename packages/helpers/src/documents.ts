type Translator = (s: string) => string;

export const documentTypes = ["OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE"] as const;

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
