/**
 * @generated SignedSource<<bbbb8700958f1bee82226fdbf53189b3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteWebhookSubscriptionInput = {
  webhookSubscriptionId: string;
};
export type WebhooksSettingsPage_deleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteWebhookSubscriptionInput;
};
export type WebhooksSettingsPage_deleteMutation$data = {
  readonly deleteWebhookSubscription: {
    readonly deletedWebhookSubscriptionId: string;
  };
};
export type WebhooksSettingsPage_deleteMutation = {
  response: WebhooksSettingsPage_deleteMutation$data;
  variables: WebhooksSettingsPage_deleteMutation$variables;
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
  "name": "deletedWebhookSubscriptionId",
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
    "name": "WebhooksSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteWebhookSubscriptionPayload",
        "kind": "LinkedField",
        "name": "deleteWebhookSubscription",
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
    "name": "WebhooksSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteWebhookSubscriptionPayload",
        "kind": "LinkedField",
        "name": "deleteWebhookSubscription",
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
            "name": "deletedWebhookSubscriptionId",
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
    "cacheID": "c1bab418b14488e982f2c444b8693919",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_deleteMutation(\n  $input: DeleteWebhookSubscriptionInput!\n) {\n  deleteWebhookSubscription(input: $input) {\n    deletedWebhookSubscriptionId\n  }\n}\n"
  }
};
})();

(node as any).hash = "794d73fc91d4332f5e65a070ea11d27d";

export default node;
