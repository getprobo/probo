/**
 * @generated SignedSource<<5b81c905d9cd12c78b5f77efdcc654ea>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SignInInput = {
  email: any;
  password: string;
};
export type SignInPageMutation$variables = {
  input: SignInInput;
};
export type SignInPageMutation$data = {
  readonly signIn: {
    readonly session: {
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
};
export type SignInPageMutation = {
  response: SignInPageMutation$data;
  variables: SignInPageMutation$variables;
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
    "name": "SignInPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SignInPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8e13b6ebdfd4ca6bde158dff53f596db",
    "id": null,
    "metadata": {},
    "name": "SignInPageMutation",
    "operationKind": "mutation",
    "text": "mutation SignInPageMutation(\n  $input: SignInInput!\n) {\n  signIn(input: $input) {\n    session {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6424104df106b64a6237d92299239a63";

export default node;
