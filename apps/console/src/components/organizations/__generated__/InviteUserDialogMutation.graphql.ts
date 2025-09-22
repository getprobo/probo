/**
 * @generated SignedSource<<9ed3f33dde5f6a00c742b8db245ba03b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type InviteUserInput = {
  createPeople: boolean;
  email: string;
  fullName: string;
  organizationId: string;
};
export type InviteUserDialogMutation$variables = {
  input: InviteUserInput;
};
export type InviteUserDialogMutation$data = {
  readonly inviteUser: {
    readonly success: boolean;
  };
};
export type InviteUserDialogMutation = {
  response: InviteUserDialogMutation$data;
  variables: InviteUserDialogMutation$variables;
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
    "concreteType": "InviteUserPayload",
    "kind": "LinkedField",
    "name": "inviteUser",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
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
    "name": "InviteUserDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "InviteUserDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "aa667927b5e2cd0019a8457edc286181",
    "id": null,
    "metadata": {},
    "name": "InviteUserDialogMutation",
    "operationKind": "mutation",
    "text": "mutation InviteUserDialogMutation(\n  $input: InviteUserInput!\n) {\n  inviteUser(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "de4c5b5208a2ff0953e9d15b844df842";

export default node;
