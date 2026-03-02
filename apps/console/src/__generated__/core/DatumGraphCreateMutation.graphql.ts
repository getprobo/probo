/**
 * @generated SignedSource<<ab09b8693f7f7e20e0a8f516a8b01017>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DataClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type CreateDatumInput = {
  dataClassification: DataClassification;
  name: string;
  organizationId: string;
  ownerId: string;
  vendorIds?: ReadonlyArray<string> | null | undefined;
};
export type DatumGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateDatumInput;
};
export type DatumGraphCreateMutation$data = {
  readonly createDatum: {
    readonly datumEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
        readonly dataClassification: DataClassification;
        readonly id: string;
        readonly name: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
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
export type DatumGraphCreateMutation = {
  response: DatumGraphCreateMutation$data;
  variables: DatumGraphCreateMutation$variables;
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
  "concreteType": "DatumEdge",
  "kind": "LinkedField",
  "name": "datumEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Datum",
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
          "name": "dataClassification",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "Profile",
          "kind": "LinkedField",
          "name": "owner",
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
              "value": "core:datum:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:datum:update\")"
        },
        {
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
    "name": "DatumGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDatumPayload",
        "kind": "LinkedField",
        "name": "createDatum",
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
    "name": "DatumGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateDatumPayload",
        "kind": "LinkedField",
        "name": "createDatum",
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
            "name": "datumEdge",
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
    "cacheID": "37d345acfe37d5edf653d021f8aad43c",
    "id": null,
    "metadata": {},
    "name": "DatumGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation DatumGraphCreateMutation(\n  $input: CreateDatumInput!\n) {\n  createDatum(input: $input) {\n    datumEdge {\n      node {\n        id\n        name\n        dataClassification\n        owner {\n          id\n          fullName\n        }\n        vendors(first: 50) {\n          edges {\n            node {\n              id\n              name\n              websiteUrl\n            }\n          }\n        }\n        createdAt\n        canUpdate: permission(action: \"core:datum:update\")\n        canDelete: permission(action: \"core:datum:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "db4781176e1b58224619f8237af92716";

export default node;
