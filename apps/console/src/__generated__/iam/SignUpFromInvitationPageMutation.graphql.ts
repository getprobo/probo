/**
 * @generated SignedSource<<59df18257cc2bae52aefd576c20f03cf>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SignUpFromInvitationInput = {
  fullName: string;
  password: string;
  token: string;
};
export type SignUpFromInvitationPageMutation$variables = {
  input: SignUpFromInvitationInput;
};
export type SignUpFromInvitationPageMutation$data = {
  readonly signUpFromInvitation: {
    readonly identity: {
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
};
export type SignUpFromInvitationPageMutation = {
  response: SignUpFromInvitationPageMutation$data;
  variables: SignUpFromInvitationPageMutation$variables;
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
    "concreteType": "SignUpFromInvitationPayload",
    "kind": "LinkedField",
    "name": "signUpFromInvitation",
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
    "name": "SignUpFromInvitationPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SignUpFromInvitationPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "9bdca3bd56cb86a2e9c0627379f7f58f",
    "id": null,
    "metadata": {},
    "name": "SignUpFromInvitationPageMutation",
    "operationKind": "mutation",
    "text": "mutation SignUpFromInvitationPageMutation(\n  $input: SignUpFromInvitationInput!\n) {\n  signUpFromInvitation(input: $input) {\n    identity {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "418fe53fbdc92e4ce5d2256062f08322";

export default node;
