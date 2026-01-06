/**
 * @generated SignedSource<<7f9521bd0d19cc1410fcb66ddc75fbf2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ResetPasswordInput = {
  password: string;
  token: string;
};
export type ResetPasswordPageMutation$variables = {
  input: ResetPasswordInput;
};
export type ResetPasswordPageMutation$data = {
  readonly resetPassword: {
    readonly success: boolean;
  } | null | undefined;
};
export type ResetPasswordPageMutation = {
  response: ResetPasswordPageMutation$data;
  variables: ResetPasswordPageMutation$variables;
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
    "concreteType": "ResetPasswordPayload",
    "kind": "LinkedField",
    "name": "resetPassword",
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
    "name": "ResetPasswordPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ResetPasswordPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e83c05ebe7c286a362ab11d60182ab61",
    "id": null,
    "metadata": {},
    "name": "ResetPasswordPageMutation",
    "operationKind": "mutation",
    "text": "mutation ResetPasswordPageMutation(\n  $input: ResetPasswordInput!\n) {\n  resetPassword(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "dc17c5c4103a29d3f163d08877bfbb34";

export default node;
