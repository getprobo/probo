/**
 * @generated SignedSource<<5dea7ccaf98a1223d9e4cc1ecdfbb0bc>>
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
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
export type ProcessingActivityGraphNodeQuery$variables = {
  processingActivityId: string;
};
export type ProcessingActivityGraphNodeQuery$data = {
  readonly node: {
    readonly consentEvidenceLink?: string | null | undefined;
    readonly createdAt?: any;
    readonly dataProtectionImpactAssessment?: ProcessingActivityDataProtectionImpactAssessment;
    readonly dataSubjectCategory?: string | null | undefined;
    readonly id?: string;
    readonly internationalTransfers?: boolean;
    readonly lawfulBasis?: ProcessingActivityLawfulBasis;
    readonly location?: string | null | undefined;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly personalDataCategory?: string | null | undefined;
    readonly purpose?: string | null | undefined;
    readonly recipients?: string | null | undefined;
    readonly retentionPeriod?: string | null | undefined;
    readonly securityMeasures?: string | null | undefined;
    readonly snapshotId?: string | null | undefined;
    readonly specialOrCriminalData?: ProcessingActivitySpecialOrCriminalData;
    readonly transferImpactAssessment?: ProcessingActivityTransferImpactAssessment;
    readonly transferSafeguards?: ProcessingActivityTransferSafeguards | null | undefined;
    readonly updatedAt?: any;
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
  "name": "dataProtectionImpactAssessment",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "transferImpactAssessment",
  "storageKey": null
},
v19 = {
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
v20 = {
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
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v22 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
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
              (v22/*: any*/)
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
              (v22/*: any*/)
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
    "cacheID": "9e9d13d95bf99c04488b8bdd313e8f33",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphNodeQuery",
    "operationKind": "query",
    "text": "query ProcessingActivityGraphNodeQuery(\n  $processingActivityId: ID!\n) {\n  node(id: $processingActivityId) {\n    __typename\n    ... on ProcessingActivity {\n      id\n      snapshotId\n      name\n      purpose\n      dataSubjectCategory\n      personalDataCategory\n      specialOrCriminalData\n      consentEvidenceLink\n      lawfulBasis\n      recipients\n      location\n      internationalTransfers\n      transferSafeguards\n      retentionPeriod\n      securityMeasures\n      dataProtectionImpactAssessment\n      transferImpactAssessment\n      vendors(first: 50) {\n        edges {\n          node {\n            id\n            name\n            websiteUrl\n            category\n          }\n        }\n      }\n      organization {\n        id\n        name\n      }\n      createdAt\n      updatedAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "0112adb4f323533e0ce7a0922c5866d2";

export default node;
