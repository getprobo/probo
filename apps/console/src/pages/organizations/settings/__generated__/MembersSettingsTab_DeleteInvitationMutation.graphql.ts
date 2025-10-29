/**
 * @generated SignedSource<<057a86325a80ac7377b18a50896f01e1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteInvitationInput = {
  invitationId: string;
};
export type MembersSettingsTab_DeleteInvitationMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteInvitationInput;
};
export type MembersSettingsTab_DeleteInvitationMutation$data = {
  readonly deleteInvitation: {
    readonly deletedInvitationId: string;
  };
};
export type MembersSettingsTab_DeleteInvitationMutation = {
  response: MembersSettingsTab_DeleteInvitationMutation$data;
  variables: MembersSettingsTab_DeleteInvitationMutation$variables;
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
  "name": "deletedInvitationId",
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
    "name": "MembersSettingsTab_DeleteInvitationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteInvitationPayload",
        "kind": "LinkedField",
        "name": "deleteInvitation",
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
    "name": "MembersSettingsTab_DeleteInvitationMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteInvitationPayload",
        "kind": "LinkedField",
        "name": "deleteInvitation",
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
            "name": "deletedInvitationId",
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
    "cacheID": "c995f13c967dab141f64f9ad6314f3a2",
    "id": null,
    "metadata": {},
    "name": "MembersSettingsTab_DeleteInvitationMutation",
    "operationKind": "mutation",
    "text": "mutation MembersSettingsTab_DeleteInvitationMutation(\n  $input: DeleteInvitationInput!\n) {\n  deleteInvitation(input: $input) {\n    deletedInvitationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "ad47509295919c7f0e6ff7895777231f";

export default node;
