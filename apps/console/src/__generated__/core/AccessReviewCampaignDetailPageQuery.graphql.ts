/**
 * @generated SignedSource<<7f577672301845bf16478a2ee1294e39>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AccessReviewCampaignSourceFetchStatus = "FAILED" | "FETCHING" | "QUEUED" | "SUCCESS";
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
export type AccessReviewCampaignDetailPageQuery$variables = {
  campaignId: string;
};
export type AccessReviewCampaignDetailPageQuery$data = {
  readonly node: {
    readonly canCancel?: boolean;
    readonly canClose?: boolean;
    readonly canStart?: boolean;
    readonly completedAt?: string | null | undefined;
    readonly createdAt?: string;
    readonly frameworkControls?: ReadonlyArray<string> | null | undefined;
    readonly id?: string;
    readonly name?: string;
    readonly scopeSources?: ReadonlyArray<{
      readonly attemptCount: number;
      readonly fetchCompletedAt: string | null | undefined;
      readonly fetchStartedAt: string | null | undefined;
      readonly fetchStatus: AccessReviewCampaignSourceFetchStatus;
      readonly fetchedAccountsCount: number;
      readonly id: string;
      readonly lastError: string | null | undefined;
      readonly name: string;
    }>;
    readonly startedAt?: string | null | undefined;
    readonly status?: AccessReviewCampaignStatus;
    readonly updatedAt?: string;
    readonly " $fragmentSpreads": FragmentRefs<"AccessReviewCampaignDetailPageEntriesFragment">;
  };
};
export type AccessReviewCampaignDetailPageQuery = {
  response: AccessReviewCampaignDetailPageQuery$data;
  variables: AccessReviewCampaignDetailPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "campaignId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "campaignId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "startedAt",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "completedAt",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "frameworkControls",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v10 = {
  "alias": "canStart",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-review-campaign:start"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-review-campaign:start\")"
},
v11 = {
  "alias": "canClose",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-review-campaign:close"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-review-campaign:close\")"
},
v12 = {
  "alias": "canCancel",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-review-campaign:cancel"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-review-campaign:cancel\")"
},
v13 = {
  "alias": null,
  "args": null,
  "concreteType": "AccessReviewCampaignScopeSource",
  "kind": "LinkedField",
  "name": "scopeSources",
  "plural": true,
  "selections": [
    (v2/*: any*/),
    (v3/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fetchStatus",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fetchedAccountsCount",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "attemptCount",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "lastError",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fetchStartedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fetchCompletedAt",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v15 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "AccessReviewCampaignDetailPageQuery",
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
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "AccessReviewCampaignDetailPageEntriesFragment"
              }
            ],
            "type": "AccessReviewCampaign",
            "abstractKey": null
          }
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
    "name": "AccessReviewCampaignDetailPageQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v14/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "alias": null,
                "args": (v15/*: any*/),
                "concreteType": "AccessEntryConnection",
                "kind": "LinkedField",
                "name": "entries",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AccessEntryEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "AccessEntry",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "email",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "fullName",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "role",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "flag",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "decision",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "decisionNote",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "incrementalTag",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "mfaStatus",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "authMethod",
                            "storageKey": null
                          },
                          {
                            "alias": "canDecide",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:access-entry:decide"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:access-entry:decide\")"
                          },
                          (v14/*: any*/)
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
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "hasPreviousPage",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "startCursor",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "kind": "ClientExtension",
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "__id",
                        "storageKey": null
                      }
                    ]
                  }
                ],
                "storageKey": "entries(first:50,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v15/*: any*/),
                "filters": [
                  "orderBy",
                  "accessSourceId"
                ],
                "handle": "connection",
                "key": "AccessReviewCampaignDetailPage_entries",
                "kind": "LinkedHandle",
                "name": "entries"
              }
            ],
            "type": "AccessReviewCampaign",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "e3cd6351adce7f07472dbc86ffdfa46c",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageQuery",
    "operationKind": "query",
    "text": "query AccessReviewCampaignDetailPageQuery(\n  $campaignId: ID!\n) {\n  node(id: $campaignId) {\n    __typename\n    ... on AccessReviewCampaign {\n      id\n      name\n      status\n      startedAt\n      completedAt\n      frameworkControls\n      createdAt\n      updatedAt\n      canStart: permission(action: \"core:access-review-campaign:start\")\n      canClose: permission(action: \"core:access-review-campaign:close\")\n      canCancel: permission(action: \"core:access-review-campaign:cancel\")\n      scopeSources {\n        id\n        name\n        fetchStatus\n        fetchedAccountsCount\n        attemptCount\n        lastError\n        fetchStartedAt\n        fetchCompletedAt\n      }\n      ...AccessReviewCampaignDetailPageEntriesFragment\n    }\n    id\n  }\n}\n\nfragment AccessEntryRowFragment on AccessEntry {\n  id\n  email\n  fullName\n  role\n  flag\n  decision\n  decisionNote\n  incrementalTag\n  mfaStatus\n  authMethod\n  canDecide: permission(action: \"core:access-entry:decide\")\n}\n\nfragment AccessReviewCampaignDetailPageEntriesFragment on AccessReviewCampaign {\n  entries(first: 50, orderBy: {direction: DESC, field: CREATED_AT}) {\n    edges {\n      node {\n        id\n        ...AccessEntryRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "2bab94e0c2258c6e622cb62778e5b73b";

export default node;
