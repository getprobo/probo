/**
 * @generated SignedSource<<ec7d0ef29b8366117ddf75e75669f5fa>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SendSigningNotificationsInput = {
  organizationId: string;
};
export type PolicyListViewSendSigningNotificationsMutation$variables = {
  input: SendSigningNotificationsInput;
};
export type PolicyListViewSendSigningNotificationsMutation$data = {
  readonly sendSigningNotifications: {
    readonly success: boolean;
  };
};
export type PolicyListViewSendSigningNotificationsMutation = {
  response: PolicyListViewSendSigningNotificationsMutation$data;
  variables: PolicyListViewSendSigningNotificationsMutation$variables;
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
    "concreteType": "SendSigningNotificationsPayload",
    "kind": "LinkedField",
    "name": "sendSigningNotifications",
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
    "name": "PolicyListViewSendSigningNotificationsMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PolicyListViewSendSigningNotificationsMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a6b25b9e4abf4243131b90b47f4e2ef4",
    "id": null,
    "metadata": {},
    "name": "PolicyListViewSendSigningNotificationsMutation",
    "operationKind": "mutation",
    "text": "mutation PolicyListViewSendSigningNotificationsMutation(\n  $input: SendSigningNotificationsInput!\n) {\n  sendSigningNotifications(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "71b0a1e7d209ed71a9bb6552eaf95cee";

export default node;
