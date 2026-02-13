/**
 * @generated SignedSource<<18c0948adc229a75b1d980e64730b423>>
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
          readonly endpointUrl: string;
          readonly events: {
            readonly totalCount: number;
          };
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
    "cacheID": "b5ca3cf4d05d2cb81c744e760433cb94",
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
    "text": "query WebhooksSettingsPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      webhookConfigurations(first: 50) {\n        edges {\n          node {\n            id\n            endpointUrl\n            selectedEvents\n            events(first: 0) {\n              totalCount\n            }\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "bac1812903d6f495d15bb0e70ce5ed48";

export default node;
