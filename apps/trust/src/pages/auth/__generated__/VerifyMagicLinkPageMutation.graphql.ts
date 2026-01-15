/**
 * @generated SignedSource<<5037d9722f3c6cc15c2a323ef954e3d6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type VerifyMagicLinkInput = {
  token: string;
};
export type VerifyMagicLinkPageMutation$variables = {
  input: VerifyMagicLinkInput;
};
export type VerifyMagicLinkPageMutation$data = {
  readonly verifyMagicLink: {
    readonly success: boolean;
  } | null | undefined;
};
export type VerifyMagicLinkPageMutation = {
  response: VerifyMagicLinkPageMutation$data;
  variables: VerifyMagicLinkPageMutation$variables;
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
    "concreteType": "VerifyMagicLinkPayload",
    "kind": "LinkedField",
    "name": "verifyMagicLink",
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
    "name": "VerifyMagicLinkPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "VerifyMagicLinkPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "07cf89de3f37725d847cda46557467f5",
    "id": null,
    "metadata": {},
    "name": "VerifyMagicLinkPageMutation",
    "operationKind": "mutation",
    "text": "mutation VerifyMagicLinkPageMutation(\n  $input: VerifyMagicLinkInput!\n) {\n  verifyMagicLink(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "074415601c4d50f50d177c06dfab64ef";

export default node;
