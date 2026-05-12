export type { DocumentTemplate } from "./types";

import { acceptableUseTemplate } from "./acceptableUse";
import { dataClassificationTemplate } from "./dataClassification";
import { incidentResponseTemplate } from "./incidentResponse";
import { securityPolicyTemplate } from "./securityPolicy";
import type { DocumentTemplate } from "./types";

export const documentTemplates: DocumentTemplate[] = [
  securityPolicyTemplate,
  acceptableUseTemplate,
  incidentResponseTemplate,
  dataClassificationTemplate,
];
