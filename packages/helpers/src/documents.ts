type Translator = (s: string) => string;

export const documentTypes = ["OTHER", "ISMS", "POLICY"] as const;

export function getDocumentTypeLabel(__: Translator, type: string) {
    switch (type) {
        case "OTHER":
            return __("Other");
        case "ISMS":
            return __("ISMS");
        case "POLICY":
            return __("Policy");
    }
}
