/**
 * @generated SignedSource<<e40f8c235af2bfed3a33bdf0e5a3c24b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type UpdateMembershipInput = {
  memberId: string;
  organizationId: string;
  role: MembershipRole;
};
export type MembersSettingsTab_UpdateMembershipMutation$variables = {
  input: UpdateMembershipInput;
};
export type MembersSettingsTab_UpdateMembershipMutation$data = {
  readonly updateMembership: {
    readonly membership: {
      readonly id: string;
      readonly role: MembershipRole;
    };
  };
};
export type MembersSettingsTab_UpdateMembershipMutation = {
  response: MembersSettingsTab_UpdateMembershipMutation$data;
  variables: MembersSettingsTab_UpdateMembershipMutation$variables;
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
    "name": "MembersSettingsTab_UpdateMembershipMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MembersSettingsTab_UpdateMembershipMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "675437becfdd9cc5ea20d2b40b9ace37",
    "id": null,
    "metadata": {},
    "name": "MembersSettingsTab_UpdateMembershipMutation",
    "operationKind": "mutation",
    "text": "mutation MembersSettingsTab_UpdateMembershipMutation(\n  $input: UpdateMembershipInput!\n) {\n  updateMembership(input: $input) {\n    membership {\n      id\n      role\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "29ead2b06842cc8ed98bbea1cf6c1bed";

export default node;
