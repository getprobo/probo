/**
 * @generated SignedSource<<8369cd1e8f928481ea43a4e09f11a3f4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AcceptInvitationInput = {
  invitationId: string;
};
export type InvitationCardMutation$variables = {
  input: AcceptInvitationInput;
  pendingInvitationConnections: ReadonlyArray<string>;
};
export type InvitationCardMutation$data = {
  readonly acceptInvitation: {
    readonly invitation: {
      readonly id: string;
    };
    readonly membership: {
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
    };
  } | null | undefined;
};
export type InvitationCardMutation = {
  response: InvitationCardMutation$data;
  variables: InvitationCardMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "pendingInvitationConnections"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "InvitationCardMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "AcceptInvitationPayload",
        "kind": "LinkedField",
        "name": "acceptInvitation",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Invitation",
            "kind": "LinkedField",
            "name": "invitation",
            "plural": false,
            "selections": [
              (v2/*: any*/)
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Membership",
            "kind": "LinkedField",
            "name": "membership",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MembershipCardFragment"
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "InvitationCardMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "AcceptInvitationPayload",
        "kind": "LinkedField",
        "name": "acceptInvitation",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Invitation",
            "kind": "LinkedField",
            "name": "invitation",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "alias": null,
                "args": null,
                "filters": null,
                "handle": "deleteEdge",
                "key": "",
                "kind": "ScalarHandle",
                "name": "id",
                "handleArgs": [
                  {
                    "kind": "Variable",
                    "name": "connections",
                    "variableName": "pendingInvitationConnections"
                  }
                ]
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Membership",
            "kind": "LinkedField",
            "name": "membership",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Session",
                "kind": "LinkedField",
                "name": "lastSession",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "expiresAt",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "bebfc1241803845587648eaad4988fcb",
    "id": null,
    "metadata": {},
    "name": "InvitationCardMutation",
    "operationKind": "mutation",
    "text": "mutation InvitationCardMutation(\n  $input: AcceptInvitationInput!\n) {\n  acceptInvitation(input: $input) {\n    invitation {\n      id\n    }\n    membership {\n      id\n      ...MembershipCardFragment\n    }\n  }\n}\n\nfragment MembershipCardFragment on Membership {\n  lastSession {\n    id\n    expiresAt\n  }\n}\n"
  }
};
})();

(node as any).hash = "32d7fb0d26660c95e87012d3b31068e8";

export default node;
