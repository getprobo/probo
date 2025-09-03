/**
 * @generated SignedSource<<48427eb821d3a0f4e53430bb384d0305>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityDataProtectionImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityLawfulBasis = "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
export type ProcessingActivitySpecialOrCriminalData = "NO" | "POSSIBLE" | "YES";
export type ProcessingActivityTransferImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityTransferSafeguards = "ADEQUACY_DECISION" | "BINDING_CORPORATE_RULES" | "CERTIFICATION_MECHANISMS" | "CODES_OF_CONDUCT" | "DEROGATIONS" | "STANDARD_CONTRACTUAL_CLAUSES";
export type UpdateProcessingActivityInput = {
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessment?: ProcessingActivityDataProtectionImpactAssessment | null | undefined;
  dataSubjectCategory?: string | null | undefined;
  id: string;
  internationalTransfers?: boolean | null | undefined;
  lawfulBasis?: ProcessingActivityLawfulBasis | null | undefined;
  location?: string | null | undefined;
  name?: string | null | undefined;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData?: ProcessingActivitySpecialOrCriminalData | null | undefined;
  transferImpactAssessment?: ProcessingActivityTransferImpactAssessment | null | undefined;
  transferSafeguards?: ProcessingActivityTransferSafeguards | null | undefined;
};
export type ProcessingActivityGraphUpdateMutation$variables = {
  input: UpdateProcessingActivityInput;
};
export type ProcessingActivityGraphUpdateMutation$data = {
  readonly updateProcessingActivity: {
    readonly processingActivity: {
      readonly consentEvidenceLink: string | null | undefined;
      readonly dataProtectionImpactAssessment: ProcessingActivityDataProtectionImpactAssessment;
      readonly dataSubjectCategory: string | null | undefined;
      readonly id: string;
      readonly internationalTransfers: boolean;
      readonly lawfulBasis: ProcessingActivityLawfulBasis;
      readonly location: string | null | undefined;
      readonly name: string;
      readonly personalDataCategory: string | null | undefined;
      readonly purpose: string | null | undefined;
      readonly recipients: string | null | undefined;
      readonly retentionPeriod: string | null | undefined;
      readonly securityMeasures: string | null | undefined;
      readonly specialOrCriminalData: ProcessingActivitySpecialOrCriminalData;
      readonly transferImpactAssessment: ProcessingActivityTransferImpactAssessment;
      readonly transferSafeguards: ProcessingActivityTransferSafeguards | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type ProcessingActivityGraphUpdateMutation = {
  response: ProcessingActivityGraphUpdateMutation$data;
  variables: ProcessingActivityGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateProcessingActivityPayload",
    "kind": "LinkedField",
    "name": "updateProcessingActivity",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ProcessingActivity",
        "kind": "LinkedField",
        "name": "processingActivity",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "purpose",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dataSubjectCategory",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "personalDataCategory",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "specialOrCriminalData",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "consentEvidenceLink",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "lawfulBasis",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "recipients",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "location",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "internationalTransfers",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "transferSafeguards",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "retentionPeriod",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "securityMeasures",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "dataProtectionImpactAssessment",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "transferImpactAssessment",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ProcessingActivityGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9ede6451e111196c25a7a4cee9ad86bf",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateMutation(\n  $input: UpdateProcessingActivityInput!\n) {\n  updateProcessingActivity(input: $input) {\n    processingActivity {\n      id\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessment\n      transferImpactAssessment\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ee0c2f6027edae2b80f603503f5f3815";

export default node;
