/**
 * @generated SignedSource<<53aa8e26ebabc319fa53dbec78bda306>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ComplianceBadgeOrderField = "CREATED_AT" | "NAME" | "RANK" | "UPDATED_AT";
export type OrderDirection = "ASC" | "DESC";
export type ComplianceBadgeOrder = {
  direction: OrderDirection;
  field: ComplianceBadgeOrderField;
};
export type CompliancePageBadgeListQuery$variables = {
  after?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  order?: ComplianceBadgeOrder | null | undefined;
};
export type CompliancePageBadgeListQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageBadgeListFragment">;
  };
};
export type CompliancePageBadgeListQuery = {
  response: CompliancePageBadgeListQuery$data;
  variables: CompliancePageBadgeListQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "after"
},
v1 = {
  "defaultValue": 100,
  "kind": "LocalArgument",
  "name": "first"
},
v2 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v3 = {
  "defaultValue": {
    "direction": "ASC",
    "field": "RANK"
  },
  "kind": "LocalArgument",
  "name": "order"
},
v4 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v5 = {
  "kind": "Variable",
  "name": "after",
  "variableName": "after"
},
v6 = {
  "kind": "Variable",
  "name": "first",
  "variableName": "first"
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v9 = [
  (v5/*: any*/),
  (v6/*: any*/),
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
      (v3/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageBadgeListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": [
              (v5/*: any*/),
              (v6/*: any*/),
              {
                "kind": "Variable",
                "name": "order",
                "variableName": "order"
              }
            ],
            "kind": "FragmentSpread",
            "name": "CompliancePageBadgeListFragment"
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
      (v3/*: any*/),
      (v2/*: any*/)
    ],
    "kind": "Operation",
    "name": "CompliancePageBadgeListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v7/*: any*/),
          (v8/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "ComplianceBadgeConnection",
                "kind": "LinkedField",
                "name": "complianceBadges",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ComplianceBadgeEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceBadge",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v8/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "rank",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "iconUrl",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "name",
                            "storageKey": null
                          },
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:compliance-badge:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:compliance-badge:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:compliance-badge:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:compliance-badge:delete\")"
                          },
                          (v7/*: any*/)
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
                "args": (v9/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "CompliancePageBadgeList_complianceBadges",
                "kind": "LinkedHandle",
                "name": "complianceBadges"
              }
            ],
            "type": "TrustCenter",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1c6845d2e5cbfacfde28de205885fda8",
    "id": null,
    "metadata": {},
    "name": "CompliancePageBadgeListQuery",
    "operationKind": "query",
    "text": "query CompliancePageBadgeListQuery(\n  $after: CursorKey = null\n  $first: Int = 100\n  $order: ComplianceBadgeOrder = {field: RANK, direction: ASC}\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...CompliancePageBadgeListFragment_gOVF5\n    id\n  }\n}\n\nfragment CompliancePageBadgeListFragment_gOVF5 on TrustCenter {\n  complianceBadges(first: $first, after: $after, orderBy: $order) {\n    edges {\n      node {\n        id\n        rank\n        ...CompliancePageBadgeListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment CompliancePageBadgeListItemFragment on ComplianceBadge {\n  id\n  iconUrl\n  name\n  rank\n  canUpdate: permission(action: \"core:compliance-badge:update\")\n  canDelete: permission(action: \"core:compliance-badge:delete\")\n}\n"
  }
};
})();

(node as any).hash = "2e75072eb105ce998cdd56ce051c9ea2";

export default node;
