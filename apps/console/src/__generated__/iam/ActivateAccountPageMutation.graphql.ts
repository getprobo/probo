/**
 * @generated SignedSource<<852ecee082c322aba273f04575353fbd>>
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
    "cacheID": "388c90522ecf3a3365e171a623fb2d2e",
    "id": null,
    "metadata": {},
    "name": "ActivateAccountPageMutation",
    "operationKind": "mutation",
    "text": "mutation ActivateAccountPageMutation(\n  $input: ActivateAccountInput!\n) {\n  activateAccount(input: $input) {\n    createPasswordToken\n  }\n}\n"
  }
};
})();

(node as any).hash = "c27b723383934423bf92926820426653";

export default node;
