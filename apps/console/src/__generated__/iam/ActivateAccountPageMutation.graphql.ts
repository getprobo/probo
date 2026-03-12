/**
 * @generated SignedSource<<e2bdaf924ff82d11f3c2ef4a8353c8e8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ActivateAccountInput = {
  token: string;
};
export type ActivateAccountPageMutation$variables = {
  input: ActivateAccountInput;
};
export type ActivateAccountPageMutation$data = {
  readonly activateAccount: {
    readonly createPasswordToken: string | null | undefined;
    readonly ssoLoginUrl: string | null | undefined;
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
        "kind": "ScalarField",
        "name": "createPasswordToken",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "ssoLoginUrl",
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
    "cacheID": "ab62d7332fce4af79b46c8627e551370",
    "id": null,
    "metadata": {},
    "name": "ActivateAccountPageMutation",
    "operationKind": "mutation",
    "text": "mutation ActivateAccountPageMutation(\n  $input: ActivateAccountInput!\n) {\n  activateAccount(input: $input) {\n    createPasswordToken\n    ssoLoginUrl\n  }\n}\n"
  }
};
})();

(node as any).hash = "fd91da524bb46256a9d8e1ae31590dc2";

export default node;
