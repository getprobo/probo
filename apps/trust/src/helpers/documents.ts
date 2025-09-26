export function documentTypeLabel(type: string, __: (s: string) => string) {
  switch (type) {
    case "POLICY":
      return __("Policy");
    case "ISMS":
      return __("Security");
    default:
      return __("Other");
  }
}
