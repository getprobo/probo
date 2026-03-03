/**
 * @generated SignedSource<<19d9adae53e45d4399358c9307928265>>
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
export type TrustCenterFileRow_requestAccessMutation$variables = {
  input: RequestTrustCenterFileAccessInput;
};
export type TrustCenterFileRow_requestAccessMutation$data = {
  readonly requestTrustCenterFileAccess: {
    readonly file: {
      readonly access: {
        readonly id: string;
        readonly status: DocumentAccessStatus;
      } | null | undefined;
    } | null | undefined;
  };
};
export type TrustCenterFileRow_requestAccessMutation = {
  response: TrustCenterFileRow_requestAccessMutation$data;
  variables: TrustCenterFileRow_requestAccessMutation$variables;
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
    "name": "TrustCenterFileRow_requestAccessMutation",
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
    "name": "TrustCenterFileRow_requestAccessMutation",
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
    "cacheID": "f96fd0464a34ad565cc7b6bb2e13308f",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileRow_requestAccessMutation(\n  $input: RequestTrustCenterFileAccessInput!\n) {\n  requestTrustCenterFileAccess(input: $input) {\n    file {\n      access {\n        id\n        status\n      }\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1a9f7ac6860d4bca88d04d73f9c6020d";

export default node;
