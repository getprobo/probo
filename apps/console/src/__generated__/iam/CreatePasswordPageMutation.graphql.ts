/**
 * @generated SignedSource<<cde2278cb4ed7c350f1c8727933cd92d>>
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
export type CreatePasswordPageMutation$variables = {
  input: ResetPasswordInput;
};
export type CreatePasswordPageMutation$data = {
  readonly resetPassword: {
    readonly success: boolean;
  } | null | undefined;
};
export type CreatePasswordPageMutation = {
  response: CreatePasswordPageMutation$data;
  variables: CreatePasswordPageMutation$variables;
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
    "name": "CreatePasswordPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CreatePasswordPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "aedb0b874e2e55a106724665d51f840e",
    "id": null,
    "metadata": {},
    "name": "CreatePasswordPageMutation",
    "operationKind": "mutation",
    "text": "mutation CreatePasswordPageMutation(\n  $input: ResetPasswordInput!\n) {\n  resetPassword(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "d9bd91787098b9272f0046ad09554a1a";

export default node;
