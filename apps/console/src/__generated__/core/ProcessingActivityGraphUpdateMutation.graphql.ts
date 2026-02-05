/**
 * @generated SignedSource<<39a1315d42c190fd43eeec5e06747aca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ProcessingActivityDataProtectionImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityLawfulBasis = "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
export type ProcessingActivityRole = "CONTROLLER" | "PROCESSOR";
export type ProcessingActivitySpecialOrCriminalDatum = "NO" | "POSSIBLE" | "YES";
export type ProcessingActivityTransferImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityTransferSafeguard = "ADEQUACY_DECISION" | "BINDING_CORPORATE_RULES" | "CERTIFICATION_MECHANISMS" | "CODES_OF_CONDUCT" | "DEROGATIONS" | "STANDARD_CONTRACTUAL_CLAUSES";
export type UpdateProcessingActivityInput = {
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessmentNeeded?: ProcessingActivityDataProtectionImpactAssessment | null | undefined;
  dataProtectionOfficerId?: string | null | undefined;
  dataSubjectCategory?: string | null | undefined;
  id: string;
  internationalTransfers?: boolean | null | undefined;
  lastReviewDate?: string | null | undefined;
  lawfulBasis?: ProcessingActivityLawfulBasis | null | undefined;
  location?: string | null | undefined;
  name?: string | null | undefined;
  nextReviewDate?: string | null | undefined;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  role?: ProcessingActivityRole | null | undefined;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData?: ProcessingActivitySpecialOrCriminalDatum | null | undefined;
  transferImpactAssessmentNeeded?: ProcessingActivityTransferImpactAssessment | null | undefined;
  transferSafeguards?: ProcessingActivityTransferSafeguard | null | undefined;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type ProcessingActivityGraphUpdateMutation$variables = {
  input: UpdateProcessingActivityInput;
};
export type ProcessingActivityGraphUpdateMutation$data = {
  readonly updateProcessingActivity: {
    readonly processingActivity: {
      readonly consentEvidenceLink: string | null | undefined;
      readonly dataProtectionImpactAssessmentNeeded: ProcessingActivityDataProtectionImpactAssessment;
      readonly dataProtectionOfficer: {
        readonly fullName: string;
        readonly id: string;
      } | null | undefined;
      readonly dataSubjectCategory: string | null | undefined;
      readonly id: string;
      readonly internationalTransfers: boolean;
      readonly lastReviewDate: string | null | undefined;
      readonly lawfulBasis: ProcessingActivityLawfulBasis;
      readonly location: string | null | undefined;
      readonly name: string;
      readonly nextReviewDate: string | null | undefined;
      readonly personalDataCategory: string | null | undefined;
      readonly purpose: string | null | undefined;
      readonly recipients: string | null | undefined;
      readonly retentionPeriod: string | null | undefined;
      readonly role: ProcessingActivityRole;
      readonly securityMeasures: string | null | undefined;
      readonly specialOrCriminalData: ProcessingActivitySpecialOrCriminalDatum;
      readonly transferImpactAssessmentNeeded: ProcessingActivityTransferImpactAssessment;
      readonly transferSafeguards: ProcessingActivityTransferSafeguard | null | undefined;
      readonly updatedAt: string;
      readonly vendors: {
        readonly edges: ReadonlyArray<{
          readonly node: {
            readonly id: string;
            readonly name: string;
            readonly websiteUrl: string | null | undefined;
          };
        }>;
      };
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
            "name": "dataProtectionImpactAssessmentNeeded",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "transferImpactAssessmentNeeded",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "lastReviewDate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "nextReviewDate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "role",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Profile",
            "kind": "LinkedField",
            "name": "dataProtectionOfficer",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": [
              {
                "kind": "Literal",
                "name": "first",
                "value": 50
              }
            ],
            "concreteType": "VendorConnection",
            "kind": "LinkedField",
            "name": "vendors",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Vendor",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v1/*: any*/),
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "websiteUrl",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "vendors(first:50)"
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
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphUpdateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "9df89e1a1b2eb8ee2b031a21c42dbea5",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphUpdateMutation(\n  $input: UpdateProcessingActivityInput!\n) {\n  updateProcessingActivity(input: $input) {\n    processingActivity {\n      id\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessmentNeeded\n      transferImpactAssessmentNeeded\n      lastReviewDate\n      nextReviewDate\n      role\n      dataProtectionOfficer {\n        id\n        fullName\n      }\n      vendors(first: 50) {\n        edges {\n          node {\n            id\n            name\n            websiteUrl\n          }\n        }\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b290e0aeeb96b3ca94e0002b50830611";

export default node;
