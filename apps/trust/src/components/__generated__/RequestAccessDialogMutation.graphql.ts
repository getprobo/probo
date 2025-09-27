/**
 * @generated SignedSource<<843902dd4fb6e2a5ee4d2d175ddcab03>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateTrustCenterAccessInput = {
  email: string;
  name: string;
  trustCenterId: string;
};
export type RequestAccessDialogMutation$variables = {
  input: CreateTrustCenterAccessInput;
};
export type RequestAccessDialogMutation$data = {
  readonly createTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type RequestAccessDialogMutation = {
  response: RequestAccessDialogMutation$data;
  variables: RequestAccessDialogMutation$variables;
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
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "CreateTrustCenterAccessPayload",
    "kind": "LinkedField",
    "name": "createTrustCenterAccess",
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "RequestAccessDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RequestAccessDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "13fedcfe7c72292417b76b3624c5434f",
    "id": null,
    "metadata": {},
    "name": "RequestAccessDialogMutation",
    "operationKind": "mutation",
    "text": "mutation RequestAccessDialogMutation(\n  $input: CreateTrustCenterAccessInput!\n) {\n  createTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "18075ac4298dd3ca05bbcf8fb1717a9d";

export default node;
