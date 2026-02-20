/**
 * @generated SignedSource<<5570f368d6cc3c3be80a32390e12c8f9>>
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
    readonly continue: string | null | undefined;
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
        "name": "continue",
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
    "cacheID": "05d0e504b6f11ad7dd11ed84059a96ac",
    "id": null,
    "metadata": {},
    "name": "VerifyMagicLinkPageMutation",
    "operationKind": "mutation",
    "text": "mutation VerifyMagicLinkPageMutation(\n  $input: VerifyMagicLinkInput!\n) {\n  verifyMagicLink(input: $input) {\n    continue\n  }\n}\n"
  }
};
})();

(node as any).hash = "cc9e7f7886d9d61b95f13c5ba66c41ec";

export default node;
