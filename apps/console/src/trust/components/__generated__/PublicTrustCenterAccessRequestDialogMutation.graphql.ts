/**
 * @generated SignedSource<<4ce5109725ad53ad77aedf4d39bf463b>>
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
export type PublicTrustCenterAccessRequestDialogMutation$variables = {
  input: CreateTrustCenterAccessInput;
};
export type PublicTrustCenterAccessRequestDialogMutation$data = {
  readonly createTrustCenterAccess: {
    readonly trustCenterAccess: {
      readonly email: string;
      readonly id: string;
      readonly name: string;
    };
  };
};
export type PublicTrustCenterAccessRequestDialogMutation = {
  response: PublicTrustCenterAccessRequestDialogMutation$data;
  variables: PublicTrustCenterAccessRequestDialogMutation$variables;
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
    "name": "PublicTrustCenterAccessRequestDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PublicTrustCenterAccessRequestDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "bdc03aeaa40a243db4af3886abffb1a0",
    "id": null,
    "metadata": {},
    "name": "PublicTrustCenterAccessRequestDialogMutation",
    "operationKind": "mutation",
    "text": "mutation PublicTrustCenterAccessRequestDialogMutation(\n  $input: CreateTrustCenterAccessInput!\n) {\n  createTrustCenterAccess(input: $input) {\n    trustCenterAccess {\n      id\n      email\n      name\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "d895df22aae3dcc8438bb794910cbfcc";

export default node;
