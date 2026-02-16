/**
 * @generated SignedSource<<dd59cf57f55eff8458d1481b25e38991>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type InviteUserInput = {
  organizationId: string;
  profileId: string;
};
export type InviteUserDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: InviteUserInput;
};
export type InviteUserDialogMutation$data = {
  readonly inviteUser: {
    readonly invitationEdge: {
      readonly node: {
        readonly acceptedAt: string | null | undefined;
        readonly canDelete: boolean;
        readonly createdAt: string;
        readonly email: string;
        readonly expiresAt: string;
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
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "iam:invitation:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"iam:invitation:delete\")"
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
    "name": "InviteUserDialogMutation",
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
    "cacheID": "685e0bc12871b33065d1fe03334110b6",
    "id": null,
    "metadata": {},
    "name": "InviteUserDialogMutation",
    "operationKind": "mutation",
    "text": "mutation InviteUserDialogMutation(\n  $input: InviteUserInput!\n) {\n  inviteUser(input: $input) {\n    invitationEdge {\n      node {\n        id\n        email\n        fullName\n        role\n        expiresAt\n        acceptedAt\n        createdAt\n        canDelete: permission(action: \"iam:invitation:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "70766e9b81d0ade0f128fae2e15aac62";

export default node;
