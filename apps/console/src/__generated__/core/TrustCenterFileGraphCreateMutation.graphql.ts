/**
 * @generated SignedSource<<e70731a64db78c30ad662330d2ec2099>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type CreateTrustCenterFileInput = {
  category: string;
  file: any;
  name: string;
  organizationId: string;
  trustCenterVisibility: TrustCenterVisibility;
};
export type TrustCenterFileGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterFileInput;
};
export type TrustCenterFileGraphCreateMutation$data = {
  readonly createTrustCenterFile: {
    readonly trustCenterFileEdge: {
      readonly node: {
        readonly category: string;
        readonly createdAt: any;
        readonly fileUrl: string;
        readonly id: string;
        readonly name: string;
        readonly trustCenterVisibility: TrustCenterVisibility;
        readonly updatedAt: any;
      };
    };
  };
};
export type TrustCenterFileGraphCreateMutation = {
  response: TrustCenterFileGraphCreateMutation$data;
  variables: TrustCenterFileGraphCreateMutation$variables;
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
  "concreteType": "TrustCenterFileEdge",
  "kind": "LinkedField",
  "name": "trustCenterFileEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "TrustCenterFile",
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
          "name": "category",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fileUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "trustCenterVisibility",
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
    "name": "TrustCenterFileGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterFile",
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
    "name": "TrustCenterFileGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterFilePayload",
        "kind": "LinkedField",
        "name": "createTrustCenterFile",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "trustCenterFileEdge",
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
    "cacheID": "57c42327b83556d36745cf8218eb01b1",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileGraphCreateMutation(\n  $input: CreateTrustCenterFileInput!\n) {\n  createTrustCenterFile(input: $input) {\n    trustCenterFileEdge {\n      node {\n        id\n        name\n        category\n        fileUrl\n        trustCenterVisibility\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "fa3851bf4f5e38bc26d8a063d6afab4e";

export default node;
