type Translator = (s: string) => string;

export const documentTypes = ["OTHER", "ISMS", "POLICY", "PROCEDURE"] as const;

export function getDocumentTypeLabel(__: Translator, type: string) {
    switch (type) {
        case "OTHER":
            return __("Other");
        case "ISMS":
            return __("ISMS");
        case "POLICY":
            return __("Policy");
        case "PROCEDURE":
            return __("Procedure");
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
