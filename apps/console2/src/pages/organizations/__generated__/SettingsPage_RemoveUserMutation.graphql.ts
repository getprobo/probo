/**
 * @generated SignedSource<<e5fd0bf99e25c4f410c6c53f4025ed15>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RemoveUserInput = {
  organizationId: string;
  userId: string;
};
export type SettingsPage_RemoveUserMutation$variables = {
  input: RemoveUserInput;
};
export type SettingsPage_RemoveUserMutation$data = {
  readonly removeUser: {
    readonly success: boolean;
  };
};
export type SettingsPage_RemoveUserMutation = {
  response: SettingsPage_RemoveUserMutation$data;
  variables: SettingsPage_RemoveUserMutation$variables;
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
    "concreteType": "RemoveUserPayload",
    "kind": "LinkedField",
    "name": "removeUser",
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
    "name": "SettingsPage_RemoveUserMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SettingsPage_RemoveUserMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a397335d917a93baf77ceba83c735326",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_RemoveUserMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_RemoveUserMutation(\n  $input: RemoveUserInput!\n) {\n  removeUser(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "ba5bf0e182d15f4805080f63d1604d91";

export default node;
