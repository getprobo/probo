/**
 * @generated SignedSource<<a234c36ae13e75cd1b0d893c4e232fc4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RevokeTrustCenterAccessInput = {
  accessId: string;
};
export type TrustCenterAccessGraphRevokeMutation$variables = {
  input: RevokeTrustCenterAccessInput;
};
export type TrustCenterAccessGraphRevokeMutation$data = {
  readonly revokeTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly active: boolean;
      readonly createdAt: any;
      readonly email: string;
      readonly id: string;
      readonly name: string;
    };
  };
};
export type TrustCenterAccessGraphRevokeMutation = {
  response: TrustCenterAccessGraphRevokeMutation$data;
  variables: TrustCenterAccessGraphRevokeMutation$variables;
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
    "concreteType": "RevokeTrustCenterAccessPayload",
    "kind": "LinkedField",
    "name": "revokeTrustCenterAccess",
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
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "email",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "active",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
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
    "name": "TrustCenterAccessGraphRevokeMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAccessGraphRevokeMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b26927a2b8d02eb2e3754850b0a44d8b",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphRevokeMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphRevokeMutation(\n  $input: RevokeTrustCenterAccessInput!\n) {\n  revokeTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n      active\n      createdAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "269cce2fe60fac04d807f1004ef0777f";

export default node;
