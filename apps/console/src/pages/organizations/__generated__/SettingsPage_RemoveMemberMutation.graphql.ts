/**
 * @generated SignedSource<<d210161405099a17aa25d06ca4563634>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RemoveMemberInput = {
  memberId: string;
  organizationId: string;
};
export type SettingsPage_RemoveMemberMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RemoveMemberInput;
};
export type SettingsPage_RemoveMemberMutation$data = {
  readonly removeMember: {
    readonly deletedMemberId: string;
  };
};
export type SettingsPage_RemoveMemberMutation = {
  response: SettingsPage_RemoveMemberMutation$data;
  variables: SettingsPage_RemoveMemberMutation$variables;
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
  "name": "deletedMemberId",
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
    "name": "SettingsPage_RemoveMemberMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveMemberPayload",
        "kind": "LinkedField",
        "name": "removeMember",
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
    "name": "SettingsPage_RemoveMemberMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "RemoveMemberPayload",
        "kind": "LinkedField",
        "name": "removeMember",
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
            "name": "deletedMemberId",
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
    "cacheID": "e2dd0f4d7327ce3bc97754c85d3f700d",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_RemoveMemberMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_RemoveMemberMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMemberId\n  }\n}\n"
  }
};
})();

(node as any).hash = "9909a8b95f8d8621ffdf02da34ec8da2";

export default node;
