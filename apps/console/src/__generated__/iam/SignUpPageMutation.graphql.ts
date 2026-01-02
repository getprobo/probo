/**
 * @generated SignedSource<<450f7808784d82c4ee82e21f72c57465>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SignUpInput = {
  email: string;
  fullName: string;
  password: string;
};
export type SignUpPageMutation$variables = {
  input: SignUpInput;
};
export type SignUpPageMutation$data = {
  readonly signUp: {
    readonly identity: {
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
};
export type SignUpPageMutation = {
  response: SignUpPageMutation$data;
  variables: SignUpPageMutation$variables;
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
    "concreteType": "SignUpPayload",
    "kind": "LinkedField",
    "name": "signUp",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "identity",
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
    "name": "SignUpPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SignUpPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "266f2f87a14a72bb166348ae9142a885",
    "id": null,
    "metadata": {},
    "name": "SignUpPageMutation",
    "operationKind": "mutation",
    "text": "mutation SignUpPageMutation(\n  $input: SignUpInput!\n) {\n  signUp(input: $input) {\n    identity {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b8ec213c1c715d5a7fbfa5896047785d";

export default node;
