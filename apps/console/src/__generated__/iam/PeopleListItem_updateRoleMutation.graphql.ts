/**
 * @generated SignedSource<<021e2d930ac4fe38c64cede962db93cf>>
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
export type PeopleListItem_updateRoleMutation$variables = {
  input: UpdateMembershipInput;
};
export type PeopleListItem_updateRoleMutation$data = {
  readonly updateMembership: {
    readonly membership: {
      readonly id: string;
      readonly role: MembershipRole;
    };
  };
};
export type PeopleListItem_updateRoleMutation = {
  response: PeopleListItem_updateRoleMutation$data;
  variables: PeopleListItem_updateRoleMutation$variables;
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
    "name": "PeopleListItem_updateRoleMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PeopleListItem_updateRoleMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "d15bdc1ff79729c3878c0cef9095a241",
    "id": null,
    "metadata": {},
    "name": "PeopleListItem_updateRoleMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleListItem_updateRoleMutation(\n  $input: UpdateMembershipInput!\n) {\n  updateMembership(input: $input) {\n    membership {\n      id\n      role\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "05245fe86928cb4afa5a958835ad167e";

export default node;
