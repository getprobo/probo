/**
 * @generated SignedSource<<b5ef7f2a5ceb3688743c82362f625a55>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentDetailPageUserEmailQuery$variables = Record<PropertyKey, never>;
export type DocumentDetailPageUserEmailQuery$data = {
  readonly viewer: {
    readonly user: {
      readonly email: any;
    };
  };
};
export type DocumentDetailPageUserEmailQuery = {
  response: DocumentDetailPageUserEmailQuery$data;
  variables: DocumentDetailPageUserEmailQuery$variables;
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
    "name": "DocumentDetailPageUserEmailQuery",
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
    "name": "DocumentDetailPageUserEmailQuery",
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
    "cacheID": "9321ba3f2fb159e06a07bfa9b2d91922",
    "id": null,
    "metadata": {},
    "name": "DocumentDetailPageUserEmailQuery",
    "operationKind": "query",
    "text": "query DocumentDetailPageUserEmailQuery {\n  viewer {\n    user {\n      email\n      id\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "b4ffc243e94463feab8fd6555a674c58";

export default node;
