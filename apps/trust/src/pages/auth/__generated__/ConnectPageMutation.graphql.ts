/**
 * @generated SignedSource<<bcbab3713eaca3141e6dfb8f54a54e7d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SendMagicLinkInput = {
  email: any;
};
export type ConnectPageMutation$variables = {
  input: SendMagicLinkInput;
};
export type ConnectPageMutation$data = {
  readonly sendMagicLink: {
    readonly success: boolean;
  } | null | undefined;
};
export type ConnectPageMutation = {
  response: ConnectPageMutation$data;
  variables: ConnectPageMutation$variables;
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
    "concreteType": "SendMagicLinkPayload",
    "kind": "LinkedField",
    "name": "sendMagicLink",
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
    "name": "ConnectPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ConnectPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "95987f897611118fa771a2e54cb43b66",
    "id": null,
    "metadata": {},
    "name": "ConnectPageMutation",
    "operationKind": "mutation",
    "text": "mutation ConnectPageMutation(\n  $input: SendMagicLinkInput!\n) {\n  sendMagicLink(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "4680fea978582bb020b4613737c14d9a";

export default node;
