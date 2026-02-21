/**
 * @generated SignedSource<<240851e53a2871e13a6e19a5b327684c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhookEventType = "MEETING_CREATED" | "MEETING_DELETED" | "MEETING_UPDATED" | "VENDOR_CREATED" | "VENDOR_DELETED" | "VENDOR_UPDATED";
export type CreateWebhookSubscriptionInput = {
  endpointUrl: string;
  organizationId: string;
  selectedEvents: ReadonlyArray<WebhookEventType>;
};
export type WebhooksSettingsPage_createMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateWebhookSubscriptionInput;
};
export type WebhooksSettingsPage_createMutation$data = {
  readonly createWebhookSubscription: {
    readonly webhookSubscriptionEdge: {
      readonly node: {
        readonly endpointUrl: string;
        readonly events: {
          readonly totalCount: number;
        };
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
  "concreteType": "WebhookSubscriptionEdge",
  "kind": "LinkedField",
  "name": "webhookSubscriptionEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "WebhookSubscription",
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
        },
        {
          "alias": null,
          "args": [
            {
              "kind": "Literal",
              "name": "first",
              "value": 0
            }
          ],
          "concreteType": "WebhookEventConnection",
          "kind": "LinkedField",
          "name": "events",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "totalCount",
              "storageKey": null
            }
          ],
          "storageKey": "events(first:0)"
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
        "concreteType": "CreateWebhookSubscriptionPayload",
        "kind": "LinkedField",
        "name": "createWebhookSubscription",
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
        "concreteType": "CreateWebhookSubscriptionPayload",
        "kind": "LinkedField",
        "name": "createWebhookSubscription",
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
            "name": "webhookSubscriptionEdge",
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
    "cacheID": "680c8a1d89c3f45d06e203072c68e593",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_createMutation",
    "operationKind": "mutation",
    "text": "mutation WebhooksSettingsPage_createMutation(\n  $input: CreateWebhookSubscriptionInput!\n) {\n  createWebhookSubscription(input: $input) {\n    webhookSubscriptionEdge {\n      node {\n        id\n        endpointUrl\n        selectedEvents\n        events(first: 0) {\n          totalCount\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ba13f2e1f417d449aa7d959c6411f858";

export default node;
