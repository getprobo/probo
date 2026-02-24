/**
 * @generated SignedSource<<3bf6a924487de062f5debcc4fbaf03fb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AccessEntryOrderField = "CREATED_AT";
export type OrderDirection = "ASC" | "DESC";
export type AccessEntryOrder = {
  direction: OrderDirection;
  field: AccessEntryOrderField;
};
export type AccessReviewCampaignDetailPageEntriesPaginationQuery$variables = {
  accessSourceId?: string | null | undefined;
  after?: string | null | undefined;
  before?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  last?: number | null | undefined;
  order?: AccessEntryOrder | null | undefined;
};
export type AccessReviewCampaignDetailPageEntriesPaginationQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"AccessReviewCampaignDetailPageEntriesFragment">;
  };
};
export type AccessReviewCampaignDetailPageEntriesPaginationQuery = {
  response: AccessReviewCampaignDetailPageEntriesPaginationQuery$data;
  variables: AccessReviewCampaignDetailPageEntriesPaginationQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "accessSourceId"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "after"
},
v2 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "before"
},
v3 = {
  "defaultValue": 50,
  "kind": "LocalArgument",
  "name": "first"
},
v4 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v5 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "last"
},
v6 = {
  "defaultValue": {
    "direction": "DESC",
    "field": "CREATED_AT"
  },
  "kind": "LocalArgument",
  "name": "order"
},
v7 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v8 = {
  "kind": "Variable",
  "name": "accessSourceId",
  "variableName": "accessSourceId"
},
v9 = {
  "kind": "Variable",
  "name": "after",
  "variableName": "after"
},
v10 = {
  "kind": "Variable",
  "name": "before",
  "variableName": "before"
},
v11 = {
  "kind": "Variable",
  "name": "first",
  "variableName": "first"
},
v12 = {
  "kind": "Variable",
  "name": "last",
  "variableName": "last"
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v15 = [
  (v8/*: any*/),
  (v9/*: any*/),
  (v10/*: any*/),
  (v11/*: any*/),
  (v12/*: any*/),
  {
    "kind": "Variable",
    "name": "orderBy",
    "variableName": "order"
  }
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/),
      (v3/*: any*/),
      (v4/*: any*/),
      (v5/*: any*/),
      (v6/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "AccessReviewCampaignDetailPageEntriesPaginationQuery",
    "selections": [
      {
        "alias": null,
        "args": (v7/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": [
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              {
                "kind": "Variable",
                "name": "order",
                "variableName": "order"
              }
            ],
            "kind": "FragmentSpread",
            "name": "AccessReviewCampaignDetailPageEntriesFragment"
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
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/),
      (v3/*: any*/),
      (v5/*: any*/),
      (v6/*: any*/),
      (v4/*: any*/)
    ],
    "kind": "Operation",
    "name": "AccessReviewCampaignDetailPageEntriesPaginationQuery",
    "selections": [
      {
        "alias": null,
        "args": (v7/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v13/*: any*/),
          (v14/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
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
                          (v14/*: any*/),
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
                          (v13/*: any*/)
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
                "storageKey": null
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
    "cacheID": "4b3243051b7c4d4ddfa8ac0c3c18cd39",
    "id": null,
    "metadata": {},
    "name": "AccessReviewCampaignDetailPageEntriesPaginationQuery",
    "operationKind": "query",
    "text": "query AccessReviewCampaignDetailPageEntriesPaginationQuery(\n  $accessSourceId: ID = null\n  $after: CursorKey = null\n  $before: CursorKey = null\n  $first: Int = 50\n  $last: Int = null\n  $order: AccessEntryOrder = {direction: DESC, field: CREATED_AT}\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...AccessReviewCampaignDetailPageEntriesFragment_4wbOLD\n    id\n  }\n}\n\nfragment AccessEntryRowFragment on AccessEntry {\n  id\n  email\n  fullName\n  role\n  flag\n  decision\n  decisionNote\n  incrementalTag\n  mfaStatus\n  authMethod\n  canDecide: permission(action: \"core:access-entry:decide\")\n}\n\nfragment AccessReviewCampaignDetailPageEntriesFragment_4wbOLD on AccessReviewCampaign {\n  entries(first: $first, after: $after, last: $last, before: $before, orderBy: $order, accessSourceId: $accessSourceId) {\n    edges {\n      node {\n        id\n        ...AccessEntryRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "084eb3177c6ed8fe67eb8d88caed0a6f";

export default node;
