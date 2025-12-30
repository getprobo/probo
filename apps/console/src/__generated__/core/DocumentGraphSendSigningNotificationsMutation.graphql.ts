/**
 * @generated SignedSource<<242a8628b9203b02dccea5e311141ec4>>
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
export type DocumentGraphSendSigningNotificationsMutation$variables = {
  input: SendSigningNotificationsInput;
};
export type DocumentGraphSendSigningNotificationsMutation$data = {
  readonly sendSigningNotifications: {
    readonly success: boolean;
  };
};
export type DocumentGraphSendSigningNotificationsMutation = {
  response: DocumentGraphSendSigningNotificationsMutation$data;
  variables: DocumentGraphSendSigningNotificationsMutation$variables;
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
    "name": "DocumentGraphSendSigningNotificationsMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentGraphSendSigningNotificationsMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "111167f828fa6eaad3beb44c43e43f06",
    "id": null,
    "metadata": {},
    "name": "DocumentGraphSendSigningNotificationsMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentGraphSendSigningNotificationsMutation(\n  $input: SendSigningNotificationsInput!\n) {\n  sendSigningNotifications(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "eaf2fba7629e202537cfc7bd9b5c9586";

export default node;
