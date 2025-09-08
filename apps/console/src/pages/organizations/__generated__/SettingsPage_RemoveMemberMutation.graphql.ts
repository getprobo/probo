/**
 * @generated SignedSource<<994eb713978ecef3329fd9873e4c744c>>
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
  input: RemoveMemberInput;
};
export type SettingsPage_RemoveMemberMutation$data = {
  readonly removeMember: {
    readonly success: boolean;
  };
};
export type SettingsPage_RemoveMemberMutation = {
  response: SettingsPage_RemoveMemberMutation$data;
  variables: SettingsPage_RemoveMemberMutation$variables;
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
    "concreteType": "RemoveMemberPayload",
    "kind": "LinkedField",
    "name": "removeMember",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
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
    "name": "SettingsPage_RemoveMemberMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SettingsPage_RemoveMemberMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "97e29046871ce8aab01abf98a62236fc",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_RemoveMemberMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_RemoveMemberMutation(\n  $input: RemoveMemberInput!\n) {\n  removeMember(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "f61071a0fb6f6554e56e79b6b04bc135";

export default node;
