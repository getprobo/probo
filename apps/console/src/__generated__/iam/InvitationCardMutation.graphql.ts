/**
 * @generated SignedSource<<2d970c5558233522bbd695a075469cad>>
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
  membershipConnections: ReadonlyArray<string>;
  pendingInvitationConnections: ReadonlyArray<string>;
};
export type InvitationCardMutation$data = {
  readonly acceptInvitation: {
    readonly invitation: {
      readonly id: string;
    };
    readonly membershipEdge: {
      readonly node: {
        readonly id: string;
        readonly organization: {
          readonly name: string;
        } | null | undefined;
        readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
      };
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
    "name": "membershipConnections"
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
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
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
            "concreteType": "MembershipEdge",
            "kind": "LinkedField",
            "name": "membershipEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "MembershipCardFragment"
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Organization",
                    "kind": "LinkedField",
                    "name": "organization",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/)
                    ],
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
            "concreteType": "MembershipEdge",
            "kind": "LinkedField",
            "name": "membershipEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "node",
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
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Organization",
                    "kind": "LinkedField",
                    "name": "organization",
                    "plural": false,
                    "selections": [
                      (v2/*: any*/),
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "logoUrl",
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
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "membershipEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "membershipConnections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "15ed0089d35b0163252ef542a4b1b667",
    "id": null,
    "metadata": {},
    "name": "InvitationCardMutation",
    "operationKind": "mutation",
    "text": "mutation InvitationCardMutation(\n  $input: AcceptInvitationInput!\n) {\n  acceptInvitation(input: $input) {\n    invitation {\n      id\n    }\n    membershipEdge {\n      node {\n        id\n        ...MembershipCardFragment\n        organization {\n          name\n          id\n        }\n      }\n    }\n  }\n}\n\nfragment MembershipCardFragment on Membership {\n  lastSession {\n    id\n    expiresAt\n  }\n  organization {\n    id\n    name\n    logoUrl\n  }\n}\n"
  }
};
})();

(node as any).hash = "587f3fc115f88830e5ee49759f3ab7b5";

export default node;
