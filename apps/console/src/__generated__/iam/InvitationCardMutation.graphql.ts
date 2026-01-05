/**
 * @generated SignedSource<<8c8ece60966bd2f39668d4ee4459ae86>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AcceptInvitationInput = {
  invitationId: string;
};
export type InvitationCardMutation$variables = {
  input: AcceptInvitationInput;
};
export type InvitationCardMutation$data = {
  readonly acceptInvitation: {
    readonly membershipEdge: {
      readonly node: {
        readonly id: string;
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
    "concreteType": "AcceptInvitationPayload",
    "kind": "LinkedField",
    "name": "acceptInvitation",
    "plural": false,
    "selections": [
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
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "id",
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "InvitationCardMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "InvitationCardMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "592fad5ce23af0f457d7e0ccf5f7123f",
    "id": null,
    "metadata": {},
    "name": "InvitationCardMutation",
    "operationKind": "mutation",
    "text": "mutation InvitationCardMutation(\n  $input: AcceptInvitationInput!\n) {\n  acceptInvitation(input: $input) {\n    membershipEdge {\n      node {\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "00a20b1db5dd8e98838c8c9659560bb9";

export default node;
