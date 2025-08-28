/**
 * @generated SignedSource<<86ea716b0545bec9ee5eabbc2bf35b74>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityRegistryDataProtectionImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityRegistryLawfulBasis = "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
export type ProcessingActivityRegistrySpecialOrCriminalData = "NO" | "POSSIBLE" | "YES";
export type ProcessingActivityRegistryTransferImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityRegistryTransferSafeguards = "ADEQUACY_DECISION" | "BINDING_CORPORATE_RULES" | "CERTIFICATION_MECHANISMS" | "CODES_OF_CONDUCT" | "DEROGATIONS" | "STANDARD_CONTRACTUAL_CLAUSES";
export type UpdateProcessingActivityRegistryInput = {
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessment?: ProcessingActivityRegistryDataProtectionImpactAssessment | null | undefined;
  dataSubjectCategory?: string | null | undefined;
  id: string;
  internationalTransfers?: boolean | null | undefined;
  lawfulBasis?: ProcessingActivityRegistryLawfulBasis | null | undefined;
  location?: string | null | undefined;
  name?: string | null | undefined;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData?: ProcessingActivityRegistrySpecialOrCriminalData | null | undefined;
  transferImpactAssessment?: ProcessingActivityRegistryTransferImpactAssessment | null | undefined;
  transferSafeguards?: ProcessingActivityRegistryTransferSafeguards | null | undefined;
};
export type ProcessingActivityRegistryGraphUpdateMutation$variables = {
  input: UpdateProcessingActivityRegistryInput;
};
export type ProcessingActivityRegistryGraphUpdateMutation$data = {
  readonly updateProcessingActivityRegistry: {
    readonly processingActivityRegistry: {
      readonly consentEvidenceLink: string | null | undefined;
      readonly dataProtectionImpactAssessment: ProcessingActivityRegistryDataProtectionImpactAssessment;
      readonly dataSubjectCategory: string | null | undefined;
      readonly id: string;
      readonly internationalTransfers: boolean;
      readonly lawfulBasis: ProcessingActivityRegistryLawfulBasis;
      readonly location: string | null | undefined;
      readonly name: string;
      readonly personalDataCategory: string | null | undefined;
      readonly purpose: string | null | undefined;
      readonly recipients: string | null | undefined;
      readonly retentionPeriod: string | null | undefined;
      readonly securityMeasures: string | null | undefined;
      readonly specialOrCriminalData: ProcessingActivityRegistrySpecialOrCriminalData;
      readonly transferImpactAssessment: ProcessingActivityRegistryTransferImpactAssessment;
      readonly transferSafeguards: ProcessingActivityRegistryTransferSafeguards | null | undefined;
      readonly updatedAt: any;
    };
  };
};
export type ProcessingActivityRegistryGraphUpdateMutation = {
  response: ProcessingActivityRegistryGraphUpdateMutation$data;
  variables: ProcessingActivityRegistryGraphUpdateMutation$variables;
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
    "concreteType": "UpdateProcessingActivityRegistryPayload",
    "kind": "LinkedField",
    "name": "updateProcessingActivityRegistry",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "ProcessingActivityRegistry",
        "kind": "LinkedField",
        "name": "processingActivityRegistry",
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
    "name": "ProcessingActivityRegistryGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityRegistryGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d4baf575ac4c35aab308228550cc6474",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityRegistryGraphUpdateMutation(\n  $input: UpdateProcessingActivityRegistryInput!\n) {\n  updateProcessingActivityRegistry(input: $input) {\n    processingActivityRegistry {\n      id\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessment\n      transferImpactAssessment\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2a6f14684cad524d383d66421354d6c7";

export default node;
