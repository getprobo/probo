/**
 * @generated SignedSource<<1eebc63e903c6dd5ab9f0178d2aaaa90>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RemoveMemberInput = {
  membershipId: string;
  organizationId: string;
};
export type MemberListItem_removeMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RemoveMemberInput;
};
export type MemberListItem_removeMutation$data = {
  readonly removeMember: {
    readonly deletedMembershipId: string;
  } | null | undefined;
};
export type MemberListItem_removeMutation = {
  response: MemberListItem_removeMutation$data;
  variables: MemberListItem_removeMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedMembershipId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "MemberListItem_removeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveMemberPayload",
        "kind": "LinkedField",
        "name": "removeMember",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "MemberListItem_removeMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveMemberPayload",
        "kind": "LinkedField",
        "name": "removeMember",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedMembershipId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "6bf37e12e069385eabd2bba63e0e980f",
    "id": null,
    "metadata": {},
    "name": "MemberListItem_removeMutation",
    "operationKind": "mutation",
    "text": "mutation MemberListItem_removeMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMembershipId\n  }\n}\n"
  }
};
})();

(node as any).hash = "9732b5ce9ca9de5c31e71a690e1763ed";

export default node;
