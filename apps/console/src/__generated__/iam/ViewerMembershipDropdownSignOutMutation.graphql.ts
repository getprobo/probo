/**
 * @generated SignedSource<<5d6fc3e9d931dccebc57b68196c98bed>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ViewerMembershipDropdownSignOutMutation$variables = Record<PropertyKey, never>;
export type ViewerMembershipDropdownSignOutMutation$data = {
  readonly signOut: {
    readonly success: boolean;
  } | null | undefined;
};
export type ViewerMembershipDropdownSignOutMutation = {
  response: ViewerMembershipDropdownSignOutMutation$data;
  variables: ViewerMembershipDropdownSignOutMutation$variables;
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
    "name": "ViewerMembershipDropdownSignOutMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "ViewerMembershipDropdownSignOutMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "da5fc8d24f8fac6617c4e4de8d9a629d",
    "id": null,
    "metadata": {},
    "name": "ViewerMembershipDropdownSignOutMutation",
    "operationKind": "mutation",
    "text": "mutation ViewerMembershipDropdownSignOutMutation {\n  signOut {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "856ff5d4affd9ad27871ab5f4fef3b1e";

export default node;
