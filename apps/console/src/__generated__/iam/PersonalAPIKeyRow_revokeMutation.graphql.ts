/**
 * @generated SignedSource<<64ad5ad6ff333dffc7b2d5d578028d7f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RevokePersonalAPIKeyInput = {
  personalAPIKeyId: string;
};
export type PersonalAPIKeyRow_revokeMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RevokePersonalAPIKeyInput;
};
export type PersonalAPIKeyRow_revokeMutation$data = {
  readonly revokePersonalAPIKey: {
    readonly personalAPIKeyId: string;
  } | null | undefined;
};
export type PersonalAPIKeyRow_revokeMutation = {
  response: PersonalAPIKeyRow_revokeMutation$data;
  variables: PersonalAPIKeyRow_revokeMutation$variables;
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
  "kind": "ScalarField",
  "name": "personalAPIKeyId",
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
    "name": "PersonalAPIKeyRow_revokeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RevokePersonalAPIKeyPayload",
        "kind": "LinkedField",
        "name": "revokePersonalAPIKey",
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
    "name": "PersonalAPIKeyRow_revokeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RevokePersonalAPIKeyPayload",
        "kind": "LinkedField",
        "name": "revokePersonalAPIKey",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "personalAPIKeyId",
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
    "cacheID": "29dfd9de3537585482c2863a1d547031",
    "id": null,
    "metadata": {},
    "name": "PersonalAPIKeyRow_revokeMutation",
    "operationKind": "mutation",
    "text": "mutation PersonalAPIKeyRow_revokeMutation(\n  $input: RevokePersonalAPIKeyInput!\n) {\n  revokePersonalAPIKey(input: $input) {\n    personalAPIKeyId\n  }\n}\n"
  }
};
})();

(node as any).hash = "258ca02064f55b6fd4a3630ba1f904a4";

export default node;
