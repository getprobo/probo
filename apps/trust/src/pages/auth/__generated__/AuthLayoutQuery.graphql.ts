/**
 * @generated SignedSource<<c3e41c1d38cd6923e718ea03bba5bb44>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AuthLayoutQuery$variables = Record<PropertyKey, never>;
export type AuthLayoutQuery$data = {
  readonly currentTrustCenter: {
    readonly darkLogoFileUrl: string | null | undefined;
    readonly logoFileUrl: string | null | undefined;
  };
};
export type AuthLayoutQuery = {
  response: AuthLayoutQuery$data;
  variables: AuthLayoutQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "logoFileUrl",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "darkLogoFileUrl",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "AuthLayoutQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenter",
          "kind": "LinkedField",
          "name": "currentTrustCenter",
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
    "name": "AuthLayoutQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "currentTrustCenter",
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
    "cacheID": "92bbf193e99aa98673d1926bd1fdc890",
    "id": null,
    "metadata": {},
    "name": "AuthLayoutQuery",
    "operationKind": "query",
    "text": "query AuthLayoutQuery {\n  currentTrustCenter {\n    logoFileUrl\n    darkLogoFileUrl\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d76b10f9b919f4224c4c8b478e66b4f1";

export default node;
