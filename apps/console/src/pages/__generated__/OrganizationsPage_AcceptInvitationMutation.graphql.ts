/**
 * @generated SignedSource<<b883cd0523e827b2c71b89e1d249d6dd>>
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
export type OrganizationsPage_AcceptInvitationMutation$variables = {
  input: AcceptInvitationInput;
};
export type OrganizationsPage_AcceptInvitationMutation$data = {
  readonly acceptInvitation: {
    readonly invitation: {
      readonly id: string;
    };
  };
};
export type OrganizationsPage_AcceptInvitationMutation = {
  response: OrganizationsPage_AcceptInvitationMutation$data;
  variables: OrganizationsPage_AcceptInvitationMutation$variables;
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
        "concreteType": "Invitation",
        "kind": "LinkedField",
        "name": "invitation",
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationsPage_AcceptInvitationMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "OrganizationsPage_AcceptInvitationMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "cc4a442037edaa624948b5be0c009823",
    "id": null,
    "metadata": {},
    "name": "OrganizationsPage_AcceptInvitationMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationsPage_AcceptInvitationMutation(\n  $input: AcceptInvitationInput!\n) {\n  acceptInvitation(input: $input) {\n    invitation {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "190213ab6fdc068343a270b4fa94e160";

export default node;
