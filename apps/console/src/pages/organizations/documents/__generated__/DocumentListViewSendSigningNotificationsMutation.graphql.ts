/**
 * @generated SignedSource<<25fe683751dd38b0c417c7df9a7f773f>>
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
export type DocumentListViewSendSigningNotificationsMutation$variables = {
  input: SendSigningNotificationsInput;
};
export type DocumentListViewSendSigningNotificationsMutation$data = {
  readonly sendSigningNotifications: {
    readonly success: boolean;
  };
};
export type DocumentListViewSendSigningNotificationsMutation = {
  response: DocumentListViewSendSigningNotificationsMutation$data;
  variables: DocumentListViewSendSigningNotificationsMutation$variables;
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
    "name": "DocumentListViewSendSigningNotificationsMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentListViewSendSigningNotificationsMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "0eb5a217a3b01f80c0ec446df15b0c40",
    "id": null,
    "metadata": {},
    "name": "DocumentListViewSendSigningNotificationsMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentListViewSendSigningNotificationsMutation(\n  $input: SendSigningNotificationsInput!\n) {\n  sendSigningNotifications(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "8a959080f5aa7cc752ce039bd7383dfa";

export default node;
