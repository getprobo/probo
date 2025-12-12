/**
 * @generated SignedSource<<959a59c16a4d27856820501f80379429>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentsPageUserEmailQuery$variables = Record<PropertyKey, never>;
export type DocumentsPageUserEmailQuery$data = {
  readonly viewer: {
    readonly user: {
      readonly email: any;
    };
  };
};
export type DocumentsPageUserEmailQuery = {
  response: DocumentsPageUserEmailQuery$data;
  variables: DocumentsPageUserEmailQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentsPageUserEmailQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "User",
            "kind": "LinkedField",
            "name": "user",
            "plural": false,
            "selections": [
              (v0/*: any*/)
            ],
            "storageKey": null
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
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "DocumentsPageUserEmailQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "User",
            "kind": "LinkedField",
            "name": "user",
            "plural": false,
            "selections": [
              (v0/*: any*/),
              (v1/*: any*/)
            ],
            "storageKey": null
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "d2d6cd91f55d59dc35d1c919f7760bac",
    "id": null,
    "metadata": {},
    "name": "DocumentsPageUserEmailQuery",
    "operationKind": "query",
    "text": "query DocumentsPageUserEmailQuery {\n  viewer {\n    user {\n      email\n      id\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "e8af56c98fa637c3e7487374fb2eaa0e";

export default node;
