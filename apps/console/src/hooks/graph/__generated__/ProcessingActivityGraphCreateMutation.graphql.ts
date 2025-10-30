/**
 * @generated SignedSource<<d01ca9f47a4ce6e732cffd5deb9150c6>>
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
export type CreateProcessingActivityInput = {
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessment: ProcessingActivityDataProtectionImpactAssessment;
  dataSubjectCategory?: string | null | undefined;
  internationalTransfers: boolean;
  lawfulBasis: ProcessingActivityLawfulBasis;
  location?: string | null | undefined;
  name: string;
  organizationId: string;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData: ProcessingActivitySpecialOrCriminalData;
  transferImpactAssessment: ProcessingActivityTransferImpactAssessment;
  transferSafeguards?: ProcessingActivityTransferSafeguards | null | undefined;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type ProcessingActivityGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateProcessingActivityInput;
};
export type ProcessingActivityGraphCreateMutation$data = {
  readonly createProcessingActivity: {
    readonly processingActivityEdge: {
      readonly node: {
        readonly consentEvidenceLink: string | null | undefined;
        readonly createdAt: any;
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
};
export type ProcessingActivityGraphCreateMutation = {
  response: ProcessingActivityGraphCreateMutation$data;
  variables: ProcessingActivityGraphCreateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "ProcessingActivityEdge",
  "kind": "LinkedField",
  "name": "processingActivityEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ProcessingActivity",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        (v4/*: any*/),
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
                    (v3/*: any*/),
                    (v4/*: any*/),
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
          "name": "createdAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "ProcessingActivityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateProcessingActivityPayload",
        "kind": "LinkedField",
        "name": "createProcessingActivity",
        "plural": false,
        "selections": [
          (v5/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "ProcessingActivityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateProcessingActivityPayload",
        "kind": "LinkedField",
        "name": "createProcessingActivity",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "processingActivityEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "4992c543938b28bc08d2316602d7bd1e",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateMutation(\n  $input: CreateProcessingActivityInput!\n) {\n  createProcessingActivity(input: $input) {\n    processingActivityEdge {\n      node {\n        id\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        specialOrCriminalData\n        consentEvidenceLink\n        lawfulBasis\n        recipients\n        location\n        internationalTransfers\n        transferSafeguards\n        retentionPeriod\n        securityMeasures\n        dataProtectionImpactAssessment\n        transferImpactAssessment\n        vendors(first: 50) {\n          edges {\n            node {\n              id\n              name\n              websiteUrl\n            }\n          }\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2150e01b5dfb4ebf1553817b7cb39f08";

export default node;
