/**
 * @generated SignedSource<<99134677b485c7ba80891aad922e354a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
export type DatumGraphNodeQuery$variables = {
  dataId: string;
};
export type DatumGraphNodeQuery$data = {
  readonly node: {
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly createdAt?: string;
    readonly dataClassification?: DataClassification;
    readonly id?: string;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
    };
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly snapshotId?: string | null | undefined;
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
export type DatumGraphNodeQuery = {
  response: DatumGraphNodeQuery$data;
  variables: DatumGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "dataId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "dataId"
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
  "name": "dataClassification",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "owner",
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
v7 = {
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
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/)
  ],
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v11 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:datum:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:datum:update\")"
},
v12 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:datum:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:datum:delete\")"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DatumGraphNodeQuery",
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
              (v12/*: any*/)
            ],
            "type": "Datum",
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
    "name": "DatumGraphNodeQuery",
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
              (v12/*: any*/)
            ],
            "type": "Datum",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "036588bb9b3daffb20c9d01a708976ac",
    "id": null,
    "metadata": {},
    "name": "DatumGraphNodeQuery",
    "operationKind": "query",
    "text": "query DatumGraphNodeQuery(\n  $dataId: ID!\n) {\n  node(id: $dataId) {\n    __typename\n    ... on Datum {\n      id\n      snapshotId\n      name\n      dataClassification\n      owner {\n        id\n        fullName\n      }\n      vendors(first: 50) {\n        edges {\n          node {\n            id\n            name\n            websiteUrl\n            category\n          }\n        }\n      }\n      organization {\n        id\n      }\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:datum:update\")\n      canDelete: permission(action: \"core:datum:delete\")\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "2b2e220e0bdd98bbee7ea72a00db7a61";

export default node;
