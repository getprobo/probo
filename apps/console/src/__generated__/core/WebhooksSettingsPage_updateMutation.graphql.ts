/**
 * @generated SignedSource<<3331839e2cc81d14ac1b2fcfe3fd84f1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhookEventType = "MEETING_CREATED" | "MEETING_DELETED" | "MEETING_UPDATED" | "VENDOR_CREATED" | "VENDOR_DELETED" | "VENDOR_UPDATED";
export type UpdateWebhookConfigurationInput = {
  endpointUrl?: string | null | undefined;
  id: string;
  selectedEvents?: ReadonlyArray<WebhookEventType> | null | undefined;
};
export type WebhooksSettingsPage_updateMutation$variables = {
  input: UpdateWebhookConfigurationInput;
};
export type WebhooksSettingsPage_updateMutation$data = {
  readonly updateWebhookConfiguration: {
    readonly webhookConfiguration: {
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
    "concreteType": "UpdateWebhookConfigurationPayload",
    "kind": "LinkedField",
    "name": "updateWebhookConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "WebhookConfiguration",
        "kind": "LinkedField",
        "name": "webhookConfiguration",
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
    "cacheID": "399a35c005f10dd321784a7533fb4f80",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_updateMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_updateMutation(\n  $input: UpdateWebhookConfigurationInput!\n) {\n  updateWebhookConfiguration(input: $input) {\n    webhookConfiguration {\n      id\n      endpointUrl\n      selectedEvents\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f974aae43d0ad8b5fe3cab60bd07a3ff";

export default node;
