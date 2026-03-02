/**
 * @generated SignedSource<<4cd86d107988eb099fff3b496ceb4b1f>>
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
export type CreateProcessingActivityInput = {
  consentEvidenceLink?: string | null | undefined;
  dataProtectionImpactAssessmentNeeded: ProcessingActivityDataProtectionImpactAssessment;
  dataProtectionOfficerId?: string | null | undefined;
  dataSubjectCategory?: string | null | undefined;
  internationalTransfers: boolean;
  lastReviewDate?: string | null | undefined;
  lawfulBasis: ProcessingActivityLawfulBasis;
  location?: string | null | undefined;
  name: string;
  nextReviewDate?: string | null | undefined;
  organizationId: string;
  personalDataCategory?: string | null | undefined;
  purpose?: string | null | undefined;
  recipients?: string | null | undefined;
  retentionPeriod?: string | null | undefined;
  role: ProcessingActivityRole;
  securityMeasures?: string | null | undefined;
  specialOrCriminalData: ProcessingActivitySpecialOrCriminalDatum;
  transferImpactAssessmentNeeded: ProcessingActivityTransferImpactAssessment;
  transferSafeguards?: ProcessingActivityTransferSafeguard | null | undefined;
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
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly consentEvidenceLink: string | null | undefined;
        readonly createdAt: string;
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
            (v3/*: any*/),
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
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:processing-activity:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:processing-activity:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:processing-activity:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:processing-activity:delete\")"
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
    "cacheID": "534bd7c442063267f6556101350e0d74",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphCreateMutation(\n  $input: CreateProcessingActivityInput!\n) {\n  createProcessingActivity(input: $input) {\n    processingActivityEdge {\n      node {\n        id\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        specialOrCriminalData\n        consentEvidenceLink\n        lawfulBasis\n        recipients\n        location\n        internationalTransfers\n        transferSafeguards\n        retentionPeriod\n        securityMeasures\n        dataProtectionImpactAssessmentNeeded\n        transferImpactAssessmentNeeded\n        lastReviewDate\n        nextReviewDate\n        role\n        dataProtectionOfficer {\n          id\n          fullName\n        }\n        vendors(first: 50) {\n          edges {\n            node {\n              id\n              name\n              websiteUrl\n            }\n          }\n        }\n        createdAt\n        canUpdate: permission(action: \"core:processing-activity:update\")\n        canDelete: permission(action: \"core:processing-activity:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "69c155bfe5b8a028d1d72c3b57d8de98";

export default node;
