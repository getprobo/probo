/**
 * @generated SignedSource<<56bfc2d1746edae72b0da44526fab12f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhookEventType = "MEETING_CREATED" | "MEETING_DELETED" | "MEETING_UPDATED" | "VENDOR_CREATED" | "VENDOR_DELETED" | "VENDOR_UPDATED";
export type CreateWebhookConfigurationInput = {
  endpointUrl: string;
  organizationId: string;
  selectedEvents: ReadonlyArray<WebhookEventType>;
};
export type WebhooksSettingsPage_createMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateWebhookConfigurationInput;
};
export type WebhooksSettingsPage_createMutation$data = {
  readonly createWebhookConfiguration: {
    readonly webhookConfigurationEdge: {
      readonly node: {
        readonly endpointUrl: string;
        readonly id: string;
        readonly selectedEvents: ReadonlyArray<WebhookEventType>;
      };
    };
  };
};
export type WebhooksSettingsPage_createMutation = {
  response: WebhooksSettingsPage_createMutation$data;
  variables: WebhooksSettingsPage_createMutation$variables;
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
  "concreteType": "WebhookConfigurationEdge",
  "kind": "LinkedField",
  "name": "webhookConfigurationEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "WebhookConfiguration",
      "kind": "LinkedField",
      "name": "node",
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
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "WebhooksSettingsPage_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateWebhookConfigurationPayload",
        "kind": "LinkedField",
        "name": "createWebhookConfiguration",
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
    "name": "WebhooksSettingsPage_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateWebhookConfigurationPayload",
        "kind": "LinkedField",
        "name": "createWebhookConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "webhookConfigurationEdge",
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
    "cacheID": "0e88932f635b8f16e41f730abf9e05ff",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_createMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_createMutation(\n  $input: CreateWebhookConfigurationInput!\n) {\n  createWebhookConfiguration(input: $input) {\n    webhookConfigurationEdge {\n      node {\n        id\n        endpointUrl\n        selectedEvents\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "4e1a7692c83c4dda0c2fc557f15d669d";

export default node;
