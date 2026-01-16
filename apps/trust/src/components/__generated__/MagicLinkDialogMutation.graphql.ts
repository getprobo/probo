/**
 * @generated SignedSource<<1d6509148442dad28e2af20068216f49>>
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
export type MagicLinkDialogMutation$variables = {
  input: SendMagicLinkInput;
};
export type MagicLinkDialogMutation$data = {
  readonly sendMagicLink: {
    readonly success: boolean;
  } | null | undefined;
};
export type MagicLinkDialogMutation = {
  response: MagicLinkDialogMutation$data;
  variables: MagicLinkDialogMutation$variables;
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
    "name": "MagicLinkDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MagicLinkDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d0e02db0cec956d7a21f5ad5a5a69d61",
    "id": null,
    "metadata": {},
    "name": "MagicLinkDialogMutation",
    "operationKind": "mutation",
    "text": "mutation MagicLinkDialogMutation(\n  $input: SendMagicLinkInput!\n) {\n  sendMagicLink(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "c1ee1fe3ef7c232fcb43b5cf8e56c2de";

export default node;
