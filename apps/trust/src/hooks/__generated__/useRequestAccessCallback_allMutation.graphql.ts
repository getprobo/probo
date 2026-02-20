/**
 * @generated SignedSource<<71c76287347e6e819c625f5072b4fae4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type useRequestAccessCallback_allMutation$variables = Record<PropertyKey, never>;
export type useRequestAccessCallback_allMutation$data = {
  readonly requestAllAccesses: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type useRequestAccessCallback_allMutation = {
  response: useRequestAccessCallback_allMutation$data;
  variables: useRequestAccessCallback_allMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "RequestAccessesPayload",
    "kind": "LinkedField",
    "name": "requestAllAccesses",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
        "plural": false,
        "selections": [
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
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "useRequestAccessCallback_allMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "useRequestAccessCallback_allMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "dd26f85da037142a278f934f06c8c0a0",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_allMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_allMutation {\n  requestAllAccesses {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7f32c58dba23b76a39950d08c25eaaa7";

export default node;
