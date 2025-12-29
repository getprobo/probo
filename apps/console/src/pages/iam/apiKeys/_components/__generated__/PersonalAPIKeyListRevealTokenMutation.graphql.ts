/**
 * @generated SignedSource<<7a6fb678e9147fd1810e02bd09d5cef1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RevealPersonalAPIKeyTokenInput = {
  tokenId: string;
};
export type PersonalAPIKeyListRevealTokenMutation$variables = {
  input: RevealPersonalAPIKeyTokenInput;
};
export type PersonalAPIKeyListRevealTokenMutation$data = {
  readonly revealPersonalAPIKeyToken: {
    readonly token: string;
  } | null | undefined;
};
export type PersonalAPIKeyListRevealTokenMutation = {
  response: PersonalAPIKeyListRevealTokenMutation$data;
  variables: PersonalAPIKeyListRevealTokenMutation$variables;
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
    "concreteType": "RevealPersonalAPIKeyTokenPayload",
    "kind": "LinkedField",
    "name": "revealPersonalAPIKeyToken",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "token",
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
    "name": "PersonalAPIKeyListRevealTokenMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PersonalAPIKeyListRevealTokenMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "25ffba7c8a69e22dcc46b56a6c8633ed",
    "id": null,
    "metadata": {},
    "name": "PersonalAPIKeyListRevealTokenMutation",
    "operationKind": "mutation",
    "text": "mutation PersonalAPIKeyListRevealTokenMutation(\n  $input: RevealPersonalAPIKeyTokenInput!\n) {\n  revealPersonalAPIKeyToken(input: $input) {\n    token\n  }\n}\n"
  }
};
})();

(node as any).hash = "7dad7ebd6e8086a01b2bac8aa80c532e";

export default node;
