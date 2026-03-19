export function documentTypeLabel(type: string, __: (s: string) => string) {
  switch (type) {
    case "POLICY":
      return __("Policy");
    case "GOVERNANCE":
      return __("Governance");
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
    default:
      return __("Other");
  }
}
