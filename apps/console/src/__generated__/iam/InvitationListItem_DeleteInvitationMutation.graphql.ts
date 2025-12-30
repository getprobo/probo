/**
 * @generated SignedSource<<a46996c5fe127b005124bc79ea88fea4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteInvitationInput = {
  invitationId: string;
  organizationId: string;
};
export type InvitationListItem_DeleteInvitationMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteInvitationInput;
};
export type InvitationListItem_DeleteInvitationMutation$data = {
  readonly deleteInvitation: {
    readonly deletedInvitationId: string;
  } | null | undefined;
};
export type InvitationListItem_DeleteInvitationMutation = {
  response: InvitationListItem_DeleteInvitationMutation$data;
  variables: InvitationListItem_DeleteInvitationMutation$variables;
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
    "name": "InvitationListItem_DeleteInvitationMutation",
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
    "name": "InvitationListItem_DeleteInvitationMutation",
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
    "cacheID": "35736b4156c15bb9eeb863b4d40fc4e6",
    "id": null,
    "metadata": {},
    "name": "InvitationListItem_DeleteInvitationMutation",
    "operationKind": "mutation",
    "text": "mutation InvitationListItem_DeleteInvitationMutation(\n  $input: DeleteInvitationInput!\n) {\n  deleteInvitation(input: $input) {\n    deletedInvitationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "2191c08f6c6bf5bace0d70a6481124bb";

export default node;
