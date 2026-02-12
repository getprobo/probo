/**
 * @generated SignedSource<<cc57c5db0d6374e4941bb0d4f1fb73c2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type WebhookEventType = "MEETING_CREATED" | "MEETING_DELETED" | "MEETING_UPDATED" | "VENDOR_CREATED" | "VENDOR_DELETED" | "VENDOR_UPDATED";
export type WebhooksSettingsPageQuery$variables = {
  organizationId: string;
};
export type WebhooksSettingsPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly id: string;
    readonly webhookConfigurations: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly calls: {
            readonly totalCount: number;
          };
          readonly endpointUrl: string;
          readonly id: string;
          readonly selectedEvents: ReadonlyArray<WebhookEventType>;
        };
      }>;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type WebhooksSettingsPageQuery = {
  response: WebhooksSettingsPageQuery$data;
  variables: WebhooksSettingsPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "WebhookConfigurationEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "WebhookConfiguration",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/),
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
            "concreteType": "WebhookCallConnection",
            "kind": "LinkedField",
            "name": "calls",
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
            "storageKey": "calls(first:0)"
          },
          (v2/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "cursor",
        "storageKey": null
      }
    ],
    "storageKey": null
  },
  {
    "alias": null,
    "args": null,
    "concreteType": "PageInfo",
    "kind": "LinkedField",
    "name": "pageInfo",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "endCursor",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "hasNextPage",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
],
v5 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "WebhooksSettingsPageQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": "organization",
          "args": (v1/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                (v3/*: any*/),
                {
                  "alias": "webhookConfigurations",
                  "args": null,
                  "concreteType": "WebhookConfigurationConnection",
                  "kind": "LinkedField",
                  "name": "__WebhooksSettingsPage_webhookConfigurations_connection",
                  "plural": false,
                  "selections": (v4/*: any*/),
                  "storageKey": null
                }
              ],
              "type": "Organization",
              "abstractKey": null
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "WebhooksSettingsPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v5/*: any*/),
                "concreteType": "WebhookConfigurationConnection",
                "kind": "LinkedField",
                "name": "webhookConfigurations",
                "plural": false,
                "selections": (v4/*: any*/),
                "storageKey": "webhookConfigurations(first:50)"
              },
              {
                "alias": null,
                "args": (v5/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "WebhooksSettingsPage_webhookConfigurations",
                "kind": "LinkedHandle",
                "name": "webhookConfigurations"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "97b7c454586fe958e456fa7929b46686",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "organization",
            "webhookConfigurations"
          ]
        }
      ]
    },
    "name": "WebhooksSettingsPageQuery",
    "operationKind": "query",
    "text": "query WebhooksSettingsPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      webhookConfigurations(first: 50) {\n        edges {\n          node {\n            id\n            endpointUrl\n            selectedEvents\n            calls(first: 0) {\n              totalCount\n            }\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "58012f55644dae31fe5e18c2c1e193a7";

export default node;
