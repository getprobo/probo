/**
 * @generated SignedSource<<93ee9324d2ee86042f9f46eda06d0770>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ViewerDropdownSignOutMutation$variables = Record<PropertyKey, never>;
export type ViewerDropdownSignOutMutation$data = {
  readonly signOut: {
    readonly success: boolean;
  } | null | undefined;
};
export type ViewerDropdownSignOutMutation = {
  response: ViewerDropdownSignOutMutation$data;
  variables: ViewerDropdownSignOutMutation$variables;
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
    "name": "ViewerDropdownSignOutMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "ViewerDropdownSignOutMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "9252bcfdab3361ae1690a713e2765de3",
    "id": null,
    "metadata": {},
    "name": "ViewerDropdownSignOutMutation",
    "operationKind": "mutation",
    "text": "mutation ViewerDropdownSignOutMutation {\n  signOut {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "a08b0de590063055e1ccbacd1d443a0f";

export default node;
