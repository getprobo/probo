/**
 * @generated SignedSource<<f441167b6ed9402b92fcdf6dbaa77170>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MainLayoutQuery$variables = {
  organizationId: string;
};
export type MainLayoutQuery$data = {
  readonly organization: {
    readonly id?: string;
    readonly logoUrl?: string | null | undefined;
    readonly name?: string;
  };
  readonly viewer: {
    readonly id: string;
  };
};
export type MainLayoutQuery = {
  response: MainLayoutQuery$data;
  variables: MainLayoutQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "concreteType": "Viewer",
  "kind": "LinkedField",
  "name": "viewer",
  "plural": false,
  "selections": [
    (v1/*: any*/)
  ],
  "storageKey": null
},
v3 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
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
  "name": "logoUrl",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MainLayoutQuery",
    "selections": [
      (v2/*: any*/),
      {
        "alias": "organization",
        "args": (v3/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v1/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/)
            ],
            "type": "Organization",
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
    "name": "MainLayoutQuery",
    "selections": [
      (v2/*: any*/),
      {
        "alias": "organization",
        "args": (v3/*: any*/),
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
          (v1/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/)
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "2ec49ff40720bacec29e7b6b1bf1408b",
    "id": null,
    "metadata": {},
    "name": "MainLayoutQuery",
    "operationKind": "query",
    "text": "query MainLayoutQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      name\n      logoUrl\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d958a2acbd9d13698b9c2b350b05d5a0";

export default node;
