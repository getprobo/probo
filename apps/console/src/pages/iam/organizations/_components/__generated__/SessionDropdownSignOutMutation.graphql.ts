/**
 * @generated SignedSource<<8cb399eb2f38eaf3059531ea4bc3fa30>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SessionDropdownSignOutMutation$variables = Record<PropertyKey, never>;
export type SessionDropdownSignOutMutation$data = {
  readonly signOut: {
    readonly success: boolean;
  } | null | undefined;
};
export type SessionDropdownSignOutMutation = {
  response: SessionDropdownSignOutMutation$data;
  variables: SessionDropdownSignOutMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "SignOutPayload",
    "kind": "LinkedField",
    "name": "signOut",
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
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "SessionDropdownSignOutMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "SessionDropdownSignOutMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "a677190a127d7f75d5c629fab236a7a3",
    "id": null,
    "metadata": {},
    "name": "SessionDropdownSignOutMutation",
    "operationKind": "mutation",
    "text": "mutation SessionDropdownSignOutMutation {\n  signOut {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "10c9bdf35ab3961c7488973e052cf978";

export default node;
