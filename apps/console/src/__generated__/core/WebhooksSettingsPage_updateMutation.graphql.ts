/**
 * @generated SignedSource<<864603efd6135475b827a1234b128b7c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhookEventType = "MEETING_CREATED" | "MEETING_DELETED" | "MEETING_UPDATED" | "VENDOR_CREATED" | "VENDOR_DELETED" | "VENDOR_UPDATED";
export type UpdateWebhookSubscriptionInput = {
  endpointUrl?: string | null | undefined;
  id: string;
  selectedEvents?: ReadonlyArray<WebhookEventType> | null | undefined;
};
export type WebhooksSettingsPage_updateMutation$variables = {
  input: UpdateWebhookSubscriptionInput;
};
export type WebhooksSettingsPage_updateMutation$data = {
  readonly updateWebhookSubscription: {
    readonly webhookSubscription: {
      readonly endpointUrl: string;
      readonly id: string;
      readonly selectedEvents: ReadonlyArray<WebhookEventType>;
      readonly updatedAt: string;
    };
  };
};
export type WebhooksSettingsPage_updateMutation = {
  response: WebhooksSettingsPage_updateMutation$data;
  variables: WebhooksSettingsPage_updateMutation$variables;
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
    "concreteType": "UpdateWebhookSubscriptionPayload",
    "kind": "LinkedField",
    "name": "updateWebhookSubscription",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "WebhookSubscription",
        "kind": "LinkedField",
        "name": "webhookSubscription",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "endpointUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "selectedEvents",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          }
        ],
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
    "name": "WebhooksSettingsPage_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "WebhooksSettingsPage_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e12790e2de3f89c6606b50a597acb244",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_updateMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_updateMutation(\n  $input: UpdateWebhookSubscriptionInput!\n) {\n  updateWebhookSubscription(input: $input) {\n    webhookSubscription {\n      id\n      endpointUrl\n      selectedEvents\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f4c0f7667d77b64c41e436a7b5df6df5";

export default node;
