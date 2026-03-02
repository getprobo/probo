/**
 * @generated SignedSource<<480c5420502cccfbc55c32a61e1662c3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AssumePageQuery$variables = Record<PropertyKey, never>;
export type AssumePageQuery$data = {
  readonly viewer: {
    readonly __typename: "Identity";
    readonly ssoLoginURL: string | null | undefined;
  };
};
export type AssumePageQuery = {
  response: AssumePageQuery$data;
  variables: AssumePageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "ssoLoginURL",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "AssumePageQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "Identity",
          "kind": "LinkedField",
          "name": "viewer",
          "plural": false,
          "selections": [
            (v0/*: any*/),
            (v1/*: any*/)
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "AssumePageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f922b5514212bff8d2da85b12f00c51e",
    "id": null,
    "metadata": {},
    "name": "AssumePageQuery",
    "operationKind": "query",
    "text": "query AssumePageQuery {\n  viewer {\n    __typename\n    ssoLoginURL\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "266ce6c5836b9e826718b07e03fcc36e";

export default node;
