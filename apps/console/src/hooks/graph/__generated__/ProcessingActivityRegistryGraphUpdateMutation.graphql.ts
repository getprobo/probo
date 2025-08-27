/**
 * @generated SignedSource<<945fa4ed1c5bab25f39047b601307e03>>
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
  auditId?: string | null | undefined;
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
      readonly audit: {
        readonly framework: {
          readonly id: string;
          readonly name: string;
        };
        readonly id: string;
        readonly name: string | null | undefined;
      };
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
      readonly transferSafeguards: ProcessingActivityRegistryTransferSafeguards;
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
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = [
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
          (v1/*: any*/),
          (v2/*: any*/),
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
            "concreteType": "Audit",
            "kind": "LinkedField",
            "name": "audit",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v2/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  (v2/*: any*/)
                ],
                "storageKey": null
              }
            ],
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
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityRegistryGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "40f01e772e5377add784e6c455b173a9",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityRegistryGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityRegistryGraphUpdateMutation(\n  $input: UpdateProcessingActivityRegistryInput!\n) {\n  updateProcessingActivityRegistry(input: $input) {\n    processingActivityRegistry {\n      id\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessment\n      transferImpactAssessment\n      audit {\n        id\n        name\n        framework {\n          id\n          name\n        }\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0833d096c755275a5255d98276922c04";

export default node;
