/**
 * @generated SignedSource<<b5471d97fe00df7ef4c16669ebc89916>>
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
export type MembersSettingsTab_RemoveMemberMutation$variables = {
  connections: ReadonlyArray<string>;
  input: RemoveMemberInput;
};
export type MembersSettingsTab_RemoveMemberMutation$data = {
  readonly removeMember: {
    readonly deletedMemberId: string;
  };
};
export type MembersSettingsTab_RemoveMemberMutation = {
  response: MembersSettingsTab_RemoveMemberMutation$data;
  variables: MembersSettingsTab_RemoveMemberMutation$variables;
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
    "name": "MembersSettingsTab_RemoveMemberMutation",
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
    "name": "MembersSettingsTab_RemoveMemberMutation",
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
    "cacheID": "6ccac45c6bedfbfe98b6c6344ea5df28",
    "id": null,
    "metadata": {},
    "name": "MembersSettingsTab_RemoveMemberMutation",
    "operationKind": "mutation",
    "text": "mutation MembersSettingsTab_RemoveMemberMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    deletedMemberId\n  }\n}\n"
  }
};
})();

(node as any).hash = "97f72349476066a0de4e580d2e4e1b0e";

export default node;
