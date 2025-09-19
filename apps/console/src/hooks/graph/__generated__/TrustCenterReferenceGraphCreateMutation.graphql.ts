/**
 * @generated SignedSource<<216f3e7b4510c3d1127c3910bebc8b95>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateTrustCenterReferenceInput = {
  description: string;
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
      readonly cursor: any;
      readonly node: {
        readonly createdAt: any;
        readonly description: string;
        readonly id: string;
        readonly logoUrl: string;
        readonly name: string;
        readonly updatedAt: any;
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
            "handle": "prependEdge",
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
    "cacheID": "17a206ba4702d641ffd7631cb8e3e03b",
    "id": null,
    "metadata": {},
    "name": "TrustCenterReferenceGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterReferenceGraphCreateMutation(\n  $input: CreateTrustCenterReferenceInput!\n) {\n  createTrustCenterReference(input: $input) {\n    trustCenterReferenceEdge {\n      cursor\n      node {\n        id\n        name\n        description\n        websiteUrl\n        logoUrl\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e0a76c3f5582bf1ed5def08db36d9e25";

export default node;
