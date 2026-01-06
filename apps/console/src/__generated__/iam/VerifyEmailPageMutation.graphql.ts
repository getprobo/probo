/**
 * @generated SignedSource<<793a4f7b17b13f925362dab5dfb590e2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type VerifyEmailInput = {
  token: string;
};
export type VerifyEmailPageMutation$variables = {
  input: VerifyEmailInput;
};
export type VerifyEmailPageMutation$data = {
  readonly verifyEmail: {
    readonly success: boolean;
  } | null | undefined;
};
export type VerifyEmailPageMutation = {
  response: VerifyEmailPageMutation$data;
  variables: VerifyEmailPageMutation$variables;
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
    "concreteType": "VerifyEmailPayload",
    "kind": "LinkedField",
    "name": "verifyEmail",
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
    "name": "VerifyEmailPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "VerifyEmailPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c1ea88cfa9a90fed624ad996bbbec2ef",
    "id": null,
    "metadata": {},
    "name": "VerifyEmailPageMutation",
    "operationKind": "mutation",
    "text": "mutation VerifyEmailPageMutation(\n  $input: VerifyEmailInput!\n) {\n  verifyEmail(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "3e4812eb6e13aa35849bd929b26268b0";

export default node;
