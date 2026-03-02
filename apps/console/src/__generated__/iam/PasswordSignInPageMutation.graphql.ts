/**
 * @generated SignedSource<<3bef621b9a5eb43afa619cd76b2bb6b6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SignInInput = {
  email: string;
  organizationId?: string | null | undefined;
  password: string;
};
export type PasswordSignInPageMutation$variables = {
  input: SignInInput;
};
export type PasswordSignInPageMutation$data = {
  readonly signIn: {
    readonly session: {
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
};
export type PasswordSignInPageMutation = {
  response: PasswordSignInPageMutation$data;
  variables: PasswordSignInPageMutation$variables;
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
    "concreteType": "SignInPayload",
    "kind": "LinkedField",
    "name": "signIn",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Session",
        "kind": "LinkedField",
        "name": "session",
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
    "name": "PasswordSignInPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PasswordSignInPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f60c1dcde17bef06ff8170d14c077dae",
    "id": null,
    "metadata": {},
    "name": "PasswordSignInPageMutation",
    "operationKind": "mutation",
    "text": "mutation PasswordSignInPageMutation(\n  $input: SignInInput!\n) {\n  signIn(input: $input) {\n    session {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9b9b3de59d38ae7c288348c6bfc01afa";

export default node;
