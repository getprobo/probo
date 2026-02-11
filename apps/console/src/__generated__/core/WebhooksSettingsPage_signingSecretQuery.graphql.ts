/**
 * @generated SignedSource<<af555c91779a2fbceaae3666d5bae667>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhooksSettingsPage_signingSecretQuery$variables = {
  webhookConfigurationId: string;
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
    "name": "webhookConfigurationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "webhookConfigurationId"
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
  "type": "WebhookConfiguration",
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
    "cacheID": "266c1519da14eb09ed28e26cc6394f23",
    "id": null,
    "metadata": {},
    "name": "WebhooksSettingsPage_signingSecretQuery",
    "operationKind": "query",
    "text": "query WebhooksSettingsPage_signingSecretQuery(\n  $webhookConfigurationId: ID!\n) {\n  node(id: $webhookConfigurationId) {\n    __typename\n    ... on WebhookConfiguration {\n      signingSecret\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "9d981d1a353a82f234e7ce79edb6d5d9";

export default node;
