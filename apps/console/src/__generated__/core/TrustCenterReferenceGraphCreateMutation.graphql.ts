/**
 * @generated SignedSource<<9d8cd7c4e9e9c2217dce89750390f2d7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateTrustCenterReferenceInput = {
  description?: string | null | undefined;
  logoFile: any;
  name: string;
  trustCenterId: string;
  websiteUrl: string;
};
export type TrustCenterReferenceGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterReferenceInput;
};
export type TrustCenterReferenceGraphCreateMutation$data = {
  readonly createTrustCenterReference: {
    readonly trustCenterReferenceEdge: {
      readonly cursor: string;
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly logoUrl: string;
        readonly name: string;
        readonly rank: number;
        readonly updatedAt: string;
        readonly websiteUrl: string;
      };
    };
  };
};
export type TrustCenterReferenceGraphCreateMutation = {
  response: TrustCenterReferenceGraphCreateMutation$data;
  variables: TrustCenterReferenceGraphCreateMutation$variables;
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
  "concreteType": "TrustCenterReferenceEdge",
  "kind": "LinkedField",
  "name": "trustCenterReferenceEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "cursor",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "TrustCenterReference",
      "kind": "LinkedField",
      "name": "node",
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
          "name": "description",
          "storageKey": null
        },
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
          "name": "logoUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "rank",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:trust-center-reference:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:trust-center-reference:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:trust-center-reference:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:trust-center-reference:delete\")"
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
    "name": "TrustCenterReferenceGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterReferencePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterReference",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "TrustCenterReferenceGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterReferencePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterReference",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "trustCenterReferenceEdge",
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
    "cacheID": "4cffbce769a3af5a0b95188c5fdb1ab1",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphCreateMutation(\n  $input: CreateTrustCenterReferenceInput!\n) {\n  createTrustCenterReference(input: $input) {\n    trustCenterReferenceEdge {\n      cursor\n      node {\n        id\n        name\n        description\n        websiteUrl\n        logoUrl\n        rank\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:trust-center-reference:update\")\n        canDelete: permission(action: \"core:trust-center-reference:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2235fdf1f63c55b4a6670eaa0f32fe20";

export default node;
