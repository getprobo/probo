/**
 * @generated SignedSource<<a089ede7865d3d5f828ac7df7d3a2286>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ForgotPasswordInput = {
  email: string;
};
export type ForgotPasswordPageMutation$variables = {
  input: ForgotPasswordInput;
};
export type ForgotPasswordPageMutation$data = {
  readonly forgotPassword: {
    readonly success: boolean;
  } | null | undefined;
};
export type ForgotPasswordPageMutation = {
  response: ForgotPasswordPageMutation$data;
  variables: ForgotPasswordPageMutation$variables;
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
    "concreteType": "ForgotPasswordPayload",
    "kind": "LinkedField",
    "name": "forgotPassword",
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
    "name": "ForgotPasswordPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ForgotPasswordPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "bb493ef26d51e7d560857008669c2ec5",
    "id": null,
    "metadata": {},
    "name": "ForgotPasswordPageMutation",
    "operationKind": "mutation",
    "text": "mutation ForgotPasswordPageMutation(\n  $input: ForgotPasswordInput!\n) {\n  forgotPassword(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "de03243068e655c7f48b7e96e145a617";

export default node;
