/**
 * @generated SignedSource<<4b128196e63e9c3b04543e8ddd87139d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateComplianceBadgeInput = {
  iconFile: any;
  name: string;
  trustCenterId: string;
};
export type ComplianceBadgeGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateComplianceBadgeInput;
};
export type ComplianceBadgeGraphCreateMutation$data = {
  readonly createComplianceBadge: {
    readonly complianceBadgeEdge: {
      readonly cursor: string;
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
        readonly iconUrl: string;
        readonly id: string;
        readonly name: string;
        readonly rank: number;
        readonly updatedAt: string;
      };
    };
  };
};
export type ComplianceBadgeGraphCreateMutation = {
  response: ComplianceBadgeGraphCreateMutation$data;
  variables: ComplianceBadgeGraphCreateMutation$variables;
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
  "concreteType": "ComplianceBadgeEdge",
  "kind": "LinkedField",
  "name": "complianceBadgeEdge",
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
      "concreteType": "ComplianceBadge",
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
          "name": "iconUrl",
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
              "value": "core:compliance-badge:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:compliance-badge:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:compliance-badge:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:compliance-badge:delete\")"
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
    "name": "ComplianceBadgeGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceBadgePayload",
        "kind": "LinkedField",
        "name": "createComplianceBadge",
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
    "name": "ComplianceBadgeGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceBadgePayload",
        "kind": "LinkedField",
        "name": "createComplianceBadge",
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
            "name": "complianceBadgeEdge",
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
    "cacheID": "d25925f6cf0e228a066abf8f080fc661",
    "id": null,
    "metadata": {},
    "name": "ComplianceBadgeGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceBadgeGraphCreateMutation(\n  $input: CreateComplianceBadgeInput!\n) {\n  createComplianceBadge(input: $input) {\n    complianceBadgeEdge {\n      cursor\n      node {\n        id\n        name\n        iconUrl\n        rank\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:compliance-badge:update\")\n        canDelete: permission(action: \"core:compliance-badge:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ed1080ef56f1ba00329eb23c0ecc99b7";

export default node;
