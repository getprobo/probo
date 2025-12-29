/**
 * @generated SignedSource<<84eef5bca22c2ad2b73bbe068767e35a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RevokePersonalAPIKeyInput = {
  tokenId: string;
};
export type PersonalAPIKeyListRevokeMutation$variables = {
  input: RevokePersonalAPIKeyInput;
};
export type PersonalAPIKeyListRevokeMutation$data = {
  readonly revokePersonalAPIKey: {
    readonly success: boolean;
  } | null | undefined;
};
export type PersonalAPIKeyListRevokeMutation = {
  response: PersonalAPIKeyListRevokeMutation$data;
  variables: PersonalAPIKeyListRevokeMutation$variables;
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
    "concreteType": "RevokePersonalAPIKeyPayload",
    "kind": "LinkedField",
    "name": "revokePersonalAPIKey",
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
    "name": "PersonalAPIKeyListRevokeMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PersonalAPIKeyListRevokeMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fd05bb0236b583ca54be466cffe8f45b",
    "id": null,
    "metadata": {},
    "name": "PersonalAPIKeyListRevokeMutation",
    "operationKind": "mutation",
    "text": "mutation PersonalAPIKeyListRevokeMutation(\n  $input: RevokePersonalAPIKeyInput!\n) {\n  revokePersonalAPIKey(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "34ef04d19251c479827ca5028346f65b";

export default node;
