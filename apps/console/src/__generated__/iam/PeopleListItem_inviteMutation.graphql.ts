/**
 * @generated SignedSource<<5c890fa68733de0a9a1b5475cd45b1a1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type InviteUserInput = {
  organizationId: string;
  profileId: string;
};
export type PeopleListItem_inviteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: InviteUserInput;
};
export type PeopleListItem_inviteMutation$data = {
  readonly inviteUser: {
    readonly invitationEdge: {
      readonly node: {
        readonly acceptedAt: string | null | undefined;
        readonly createdAt: string;
        readonly expiresAt: string;
        readonly id: string;
      };
    };
  } | null | undefined;
};
export type PeopleListItem_inviteMutation = {
  response: PeopleListItem_inviteMutation$data;
  variables: PeopleListItem_inviteMutation$variables;
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
  "concreteType": "InvitationEdge",
  "kind": "LinkedField",
  "name": "invitationEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Invitation",
      "kind": "LinkedField",
      "name": "node",
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
          "name": "expiresAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "acceptedAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "PeopleListItem_inviteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "InviteUserPayload",
        "kind": "LinkedField",
        "name": "inviteUser",
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
    "name": "PeopleListItem_inviteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "InviteUserPayload",
        "kind": "LinkedField",
        "name": "inviteUser",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "invitationEdge",
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
    "cacheID": "adf65c15aebb17087c5701c9b83e6711",
    "id": null,
    "metadata": {},
    "name": "PeopleListItem_inviteMutation",
    "operationKind": "mutation",
    "text": "mutation PeopleListItem_inviteMutation(\n  $input: InviteUserInput!\n) {\n  inviteUser(input: $input) {\n    invitationEdge {\n      node {\n        id\n        expiresAt\n        acceptedAt\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7ce62bafa9a5ec7a965c048fe7ef1763";

export default node;
