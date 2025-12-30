/**
 * @generated SignedSource<<84f3f3da89ad2bdfecd79b479bf74814>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type InviteMemberInput = {
  createPeople: boolean;
  email: any;
  fullName: string;
  organizationId: string;
  role: MembershipRole;
};
export type InviteUserDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: InviteMemberInput;
};
export type InviteUserDialogMutation$data = {
  readonly inviteMember: {
    readonly invitationEdge: {
      readonly node: {
        readonly acceptedAt: any | null | undefined;
        readonly createdAt: any;
        readonly email: any;
        readonly expiresAt: any;
        readonly fullName: string;
        readonly id: string;
        readonly role: MembershipRole;
      };
    };
  } | null | undefined;
};
export type InviteUserDialogMutation = {
  response: InviteUserDialogMutation$data;
  variables: InviteUserDialogMutation$variables;
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
          "name": "email",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "role",
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
    "name": "InviteUserDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "InviteMemberPayload",
        "kind": "LinkedField",
        "name": "inviteMember",
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
    "name": "InviteUserDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "InviteMemberPayload",
        "kind": "LinkedField",
        "name": "inviteMember",
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
    "cacheID": "43ea730d742b7d259b50b50d3684147c",
    "id": null,
    "metadata": {},
    "name": "InviteUserDialogMutation",
    "operationKind": "mutation",
    "text": "mutation InviteUserDialogMutation(\n  $input: InviteMemberInput!\n) {\n  inviteMember(input: $input) {\n    invitationEdge {\n      node {\n        id\n        email\n        fullName\n        role\n        expiresAt\n        acceptedAt\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b45737075b750b79b3a90748cb2565f3";

export default node;
