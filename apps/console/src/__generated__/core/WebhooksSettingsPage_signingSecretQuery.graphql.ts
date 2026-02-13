/**
 * @generated SignedSource<<d341340b571530212b3cdc8722d694f1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhooksSettingsPage_signingSecretQuery$variables = {
  webhookSubscriptionId: string;
};
export type WebhooksSettingsPage_signingSecretQuery$data = {
  readonly node: {
    readonly signingSecret?: string;
  };
};
export type WebhooksSettingsPage_signingSecretQuery = {
  response: WebhooksSettingsPage_signingSecretQuery$data;
  variables: WebhooksSettingsPage_signingSecretQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "webhookSubscriptionId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "webhookSubscriptionId"
  }
],
v2 = {
  "kind": "InlineFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "signingSecret",
      "storageKey": null
    }
  ],
  "type": "WebhookSubscription",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "WebhooksSettingsPage_signingSecretQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "WebhooksSettingsPage_signingSecretQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "df35742c680c055e5b820401cfb05271",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_signingSecretQuery",
    "operationKind": "query",
    "text": "query WebhooksSettingsPage_signingSecretQuery(\n  $webhookSubscriptionId: ID!\n) {\n  node(id: $webhookSubscriptionId) {\n    __typename\n    ... on WebhookSubscription {\n      signingSecret\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "2d895bf4be96e0ef2c767af8149c8533";

export default node;
