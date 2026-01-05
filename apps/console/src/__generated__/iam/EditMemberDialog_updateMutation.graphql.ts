/**
 * @generated SignedSource<<d29dfbdd0293d60af6d3f72e09dd5e9d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type UpdateMembershipInput = {
  membershipId: string;
  organizationId: string;
  role: MembershipRole;
};
export type EditMemberDialog_updateMutation$variables = {
  input: UpdateMembershipInput;
};
export type EditMemberDialog_updateMutation$data = {
  readonly updateMembership: {
    readonly membership: {
      readonly id: string;
      readonly role: MembershipRole;
    };
  };
};
export type EditMemberDialog_updateMutation = {
  response: EditMemberDialog_updateMutation$data;
  variables: EditMemberDialog_updateMutation$variables;
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
    "concreteType": "UpdateMembershipPayload",
    "kind": "LinkedField",
    "name": "updateMembership",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Membership",
        "kind": "LinkedField",
        "name": "membership",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "role",
            "storageKey": null
          }
        ],
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
    "name": "EditMemberDialog_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditMemberDialog_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "696ec9f74220b92171bf60630e5c8d7f",
    "id": null,
    "metadata": {},
    "name": "EditMemberDialog_updateMutation",
    "operationKind": "mutation",
    "text": "mutation EditMemberDialog_updateMutation(\n  $input: UpdateMembershipInput!\n) {\n  updateMembership(input: $input) {\n    membership {\n      id\n      role\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "fc971ff69092912586c14c7f6e2111ed";

export default node;
