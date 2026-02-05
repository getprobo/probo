/**
 * @generated SignedSource<<95c0d606637cea849328f351b4d8b72b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataProtectionImpactAssessmentResidualRisk = "HIGH" | "LOW" | "MEDIUM";
export type ProcessingActivityDataProtectionImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityLawfulBasis = "CONSENT" | "CONTRACTUAL_NECESSITY" | "LEGAL_OBLIGATION" | "LEGITIMATE_INTEREST" | "PUBLIC_TASK" | "VITAL_INTERESTS";
export type ProcessingActivityRole = "CONTROLLER" | "PROCESSOR";
export type ProcessingActivitySpecialOrCriminalDatum = "NO" | "POSSIBLE" | "YES";
export type ProcessingActivityTransferImpactAssessment = "NEEDED" | "NOT_NEEDED";
export type ProcessingActivityTransferSafeguard = "ADEQUACY_DECISION" | "BINDING_CORPORATE_RULES" | "CERTIFICATION_MECHANISMS" | "CODES_OF_CONDUCT" | "DEROGATIONS" | "STANDARD_CONTRACTUAL_CLAUSES";
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
export type ProcessingActivityGraphNodeQuery$variables = {
  processingActivityId: string;
};
export type ProcessingActivityGraphNodeQuery$data = {
  readonly node: {
    readonly canCreateDPIA?: boolean;
    readonly canCreateTIA?: boolean;
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly consentEvidenceLink?: string | null | undefined;
    readonly createdAt?: string;
    readonly dataProtectionImpactAssessment?: {
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly description: string | null | undefined;
      readonly id: string;
      readonly mitigations: string | null | undefined;
      readonly necessityAndProportionality: string | null | undefined;
      readonly potentialRisk: string | null | undefined;
      readonly residualRisk: DataProtectionImpactAssessmentResidualRisk | null | undefined;
      readonly updatedAt: string;
    } | null | undefined;
    readonly dataProtectionImpactAssessmentNeeded?: ProcessingActivityDataProtectionImpactAssessment;
    readonly dataProtectionOfficer?: {
      readonly fullName: string;
      readonly id: string;
    } | null | undefined;
    readonly dataSubjectCategory?: string | null | undefined;
    readonly id?: string;
    readonly internationalTransfers?: boolean;
    readonly lastReviewDate?: string | null | undefined;
    readonly lawfulBasis?: ProcessingActivityLawfulBasis;
    readonly location?: string | null | undefined;
    readonly name?: string;
    readonly nextReviewDate?: string | null | undefined;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly personalDataCategory?: string | null | undefined;
    readonly purpose?: string | null | undefined;
    readonly recipients?: string | null | undefined;
    readonly retentionPeriod?: string | null | undefined;
    readonly role?: ProcessingActivityRole;
    readonly securityMeasures?: string | null | undefined;
    readonly snapshotId?: string | null | undefined;
    readonly specialOrCriminalData?: ProcessingActivitySpecialOrCriminalDatum;
    readonly transferImpactAssessment?: {
      readonly canDelete: boolean;
      readonly canUpdate: boolean;
      readonly createdAt: string;
      readonly dataSubjects: string | null | undefined;
      readonly id: string;
      readonly legalMechanism: string | null | undefined;
      readonly localLawRisk: string | null | undefined;
      readonly supplementaryMeasures: string | null | undefined;
      readonly transfer: string | null | undefined;
      readonly updatedAt: string;
    } | null | undefined;
    readonly transferImpactAssessmentNeeded?: ProcessingActivityTransferImpactAssessment;
    readonly transferSafeguards?: ProcessingActivityTransferSafeguard | null | undefined;
    readonly updatedAt?: string;
    readonly vendors?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: VendorCategory;
          readonly id: string;
          readonly name: string;
          readonly websiteUrl: string | null | undefined;
        };
      }>;
    };
  };
};
export type ProcessingActivityGraphNodeQuery = {
  response: ProcessingActivityGraphNodeQuery$data;
  variables: ProcessingActivityGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "processingActivityId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "processingActivityId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
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
  "kind": "ScalarField",
  "name": "purpose",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataSubjectCategory",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "personalDataCategory",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "specialOrCriminalData",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "consentEvidenceLink",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "lawfulBasis",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "recipients",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "location",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "internationalTransfers",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferSafeguards",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "retentionPeriod",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "securityMeasures",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataProtectionImpactAssessmentNeeded",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferImpactAssessmentNeeded",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "lastReviewDate",
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "nextReviewDate",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "role",
  "storageKey": null
},
v22 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "dataProtectionOfficer",
  "plural": false,
  "selections": [
    (v2/*: any*/),
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
v23 = {
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
            (v2/*: any*/),
            (v4/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "websiteUrl",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "category",
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
v24 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v25 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v26 = {
  "alias": null,
  "args": null,
  "concreteType": "DataProtectionImpactAssessment",
  "kind": "LinkedField",
  "name": "dataProtectionImpactAssessment",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "description",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "necessityAndProportionality",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "potentialRisk",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "mitigations",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "residualRisk",
      "storageKey": null
    },
    (v24/*: any*/),
    (v25/*: any*/),
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:data-protection-impact-assessment:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:data-protection-impact-assessment:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:data-protection-impact-assessment:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:data-protection-impact-assessment:delete\")"
    }
  ],
  "storageKey": null
},
v27 = {
  "alias": null,
  "args": null,
  "concreteType": "TransferImpactAssessment",
  "kind": "LinkedField",
  "name": "transferImpactAssessment",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "dataSubjects",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "legalMechanism",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "transfer",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "localLawRisk",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "supplementaryMeasures",
      "storageKey": null
    },
    (v24/*: any*/),
    (v25/*: any*/),
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:transfer-impact-assessment:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:transfer-impact-assessment:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:transfer-impact-assessment:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:transfer-impact-assessment:delete\")"
    }
  ],
  "storageKey": null
},
v28 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v4/*: any*/)
  ],
  "storageKey": null
},
v29 = {
  "alias": "canCreateDPIA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:data-protection-impact-assessment:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:data-protection-impact-assessment:create\")"
},
v30 = {
  "alias": "canCreateTIA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:transfer-impact-assessment:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:transfer-impact-assessment:create\")"
},
v31 = {
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
v32 = {
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
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ProcessingActivityGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/),
              (v23/*: any*/),
              (v26/*: any*/),
              (v27/*: any*/),
              (v28/*: any*/),
              (v24/*: any*/),
              (v25/*: any*/),
              (v29/*: any*/),
              (v30/*: any*/),
              (v31/*: any*/),
              (v32/*: any*/)
            ],
            "type": "ProcessingActivity",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/),
              (v23/*: any*/),
              (v26/*: any*/),
              (v27/*: any*/),
              (v28/*: any*/),
              (v24/*: any*/),
              (v25/*: any*/),
              (v29/*: any*/),
              (v30/*: any*/),
              (v31/*: any*/),
              (v32/*: any*/)
            ],
            "type": "ProcessingActivity",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "29bddd12b9399a78fb24420f11b6d56d",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphNodeQuery",
    "operationKind": "query",
    "text": "query ProcessingActivityGraphNodeQuery(\n  $processingActivityId: ID!\n) {\n  node(id: $processingActivityId) {\n    __typename\n    ... on ProcessingActivity {\n      id\n      snapshotId\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessmentNeeded\n      transferImpactAssessmentNeeded\n      lastReviewDate\n      nextReviewDate\n      role\n      dataProtectionOfficer {\n        id\n        fullName\n      }\n      vendors(first: 50) {\n        edges {\n          node {\n            id\n            name\n            websiteUrl\n            category\n          }\n        }\n      }\n      dataProtectionImpactAssessment {\n        id\n        description\n        necessityAndProportionality\n        potentialRisk\n        mitigations\n        residualRisk\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:data-protection-impact-assessment:update\")\n        canDelete: permission(action: \"core:data-protection-impact-assessment:delete\")\n      }\n      transferImpactAssessment {\n        id\n        dataSubjects\n        legalMechanism\n        transfer\n        localLawRisk\n        supplementaryMeasures\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:transfer-impact-assessment:update\")\n        canDelete: permission(action: \"core:transfer-impact-assessment:delete\")\n      }\n      organization {\n        id\n        name\n      }\n      createdAt\n      updatedAt\n      canCreateDPIA: permission(action: \"core:data-protection-impact-assessment:create\")\n      canCreateTIA: permission(action: \"core:transfer-impact-assessment:create\")\n      canUpdate: permission(action: \"core:processing-activity:update\")\n      canDelete: permission(action: \"core:processing-activity:delete\")\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d670e7ec4d612b19532332247c8c1f8f";

export default node;
