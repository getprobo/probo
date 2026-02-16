/**
 * @generated SignedSource<<e920b2bbd380e12014a100bfb963c8f7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ActivateAccountInput = {
  password: string;
  token: string;
};
export type ActivateAccountPageMutation$variables = {
  input: ActivateAccountInput;
};
export type ActivateAccountPageMutation$data = {
  readonly activateAccount: {
    readonly profile: {
      readonly id: string;
    } | null | undefined;
  } | null | undefined;
};
export type ActivateAccountPageMutation = {
  response: ActivateAccountPageMutation$data;
  variables: ActivateAccountPageMutation$variables;
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
    "concreteType": "ActivateAccountPayload",
    "kind": "LinkedField",
    "name": "activateAccount",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Profile",
        "kind": "LinkedField",
        "name": "profile",
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
    "name": "ActivateAccountPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ActivateAccountPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "57d92d0ae2a6220af84fbd9fe27f7aa3",
    "id": null,
    "metadata": {},
    "name": "ActivateAccountPageMutation",
    "operationKind": "mutation",
    "text": "mutation ActivateAccountPageMutation(\n  $input: ActivateAccountInput!\n) {\n  activateAccount(input: $input) {\n    profile {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0e073ce00eb7c435a875b5797f3e6db0";

export default node;
