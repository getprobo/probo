/**
 * @generated SignedSource<<f058193b755f9fed3efe4910aae9a7e3>>
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
export type SettingsPage_DeleteInvitationMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteInvitationInput;
};
export type SettingsPage_DeleteInvitationMutation$data = {
  readonly deleteInvitation: {
    readonly deletedInvitationId: string;
  };
};
export type SettingsPage_DeleteInvitationMutation = {
  response: SettingsPage_DeleteInvitationMutation$data;
  variables: SettingsPage_DeleteInvitationMutation$variables;
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
    "name": "SettingsPage_DeleteInvitationMutation",
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
    "name": "SettingsPage_DeleteInvitationMutation",
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
    "cacheID": "1c362c5db7a985d7548166b2b1eb42c9",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_DeleteInvitationMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_DeleteInvitationMutation(\n  $input: DeleteInvitationInput!\n) {\n  deleteInvitation(input: $input) {\n    deletedInvitationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "3c484508ba04b5a75eca62fa6afeb16d";

export default node;
