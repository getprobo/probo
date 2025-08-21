/**
 * @generated SignedSource<<8ce81f99207fec5774de2469a63953e1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ConfirmEmailInput = {
  authToken: string;
};
export type ConfirmEmailPageMutation$variables = {
  input: ConfirmEmailInput;
};
export type ConfirmEmailPageMutation$data = {
  readonly confirmEmail: {
    readonly success: boolean;
  };
};
export type ConfirmEmailPageMutation = {
  response: ConfirmEmailPageMutation$data;
  variables: ConfirmEmailPageMutation$variables;
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
    "concreteType": "ConfirmEmailPayload",
    "kind": "LinkedField",
    "name": "confirmEmail",
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
    "name": "ConfirmEmailPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ConfirmEmailPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "b6bde3a559a4ecb70a519ffb7fa83330",
    "id": null,
    "metadata": {},
    "name": "ConfirmEmailPageMutation",
    "operationKind": "mutation",
    "text": "mutation ConfirmEmailPageMutation(\n  $input: ConfirmEmailInput!\n) {\n  confirmEmail(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "e7f442b5acc6d912bf272ca2f5dc1ef1";

export default node;
