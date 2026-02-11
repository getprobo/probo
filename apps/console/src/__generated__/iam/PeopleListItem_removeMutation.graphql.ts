/**
 * @generated SignedSource<<8c0cc8f8d9d0418ea1d817412793f2dc>>
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
export type PeopleListItem_removeMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RemoveMemberInput;
};
export type PeopleListItem_removeMutation$data = {
  readonly removeMember: {
    readonly deletedMembershipId: string;
  } | null | undefined;
};
export type PeopleListItem_removeMutation = {
  response: PeopleListItem_removeMutation$data;
  variables: PeopleListItem_removeMutation$variables;
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
    "name": "PeopleListItem_removeMutation",
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
    "name": "PeopleListItem_removeMutation",
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
    "cacheID": "d5a0e195ee72906380f566bcecb45a3f",
    "id": null,
    "metadata": {},
    "name": "PeopleListItem_removeMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleListItem_removeMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMembershipId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c1bceb22b556aff77543ec1956445187";

export default node;
