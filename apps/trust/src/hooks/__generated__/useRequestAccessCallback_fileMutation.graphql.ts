/**
 * @generated SignedSource<<688bb06b7534c867b9427a6fcbb52d57>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type RequestTrustCenterFileAccessInput = {
  trustCenterFileId: string;
};
export type useRequestAccessCallback_fileMutation$variables = {
  input: RequestTrustCenterFileAccessInput;
};
export type useRequestAccessCallback_fileMutation$data = {
  readonly requestTrustCenterFileAccess: {
    readonly file: {
      readonly access: {
        readonly id: string;
        readonly status: DocumentAccessStatus;
      } | null | undefined;
    } | null | undefined;
  };
};
export type useRequestAccessCallback_fileMutation = {
  response: useRequestAccessCallback_fileMutation$data;
  variables: useRequestAccessCallback_fileMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "DocumentAccess",
  "kind": "LinkedField",
  "name": "access",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "useRequestAccessCallback_fileMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestFileAccessPayload",
        "kind": "LinkedField",
        "name": "requestTrustCenterFileAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFile",
            "kind": "LinkedField",
            "name": "file",
            "plural": false,
            "selections": [
              (v3/*: any*/)
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_fileMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestFileAccessPayload",
        "kind": "LinkedField",
        "name": "requestTrustCenterFileAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterFile",
            "kind": "LinkedField",
            "name": "file",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "0c66e592bdb92aa92fbaef21c8133e17",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_fileMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_fileMutation(\n  $input: RequestTrustCenterFileAccessInput!\n) {\n  requestTrustCenterFileAccess(input: $input) {\n    file {\n      access {\n        id\n        status\n      }\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0dbe835b29f81d086b94cfecb8fc832f";

export default node;
