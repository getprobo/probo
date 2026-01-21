/**
 * @generated SignedSource<<be08976a189938be59196a67454e0a60>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrderDirection = "ASC" | "DESC";
export type StateOfApplicabilityOrderField = "CREATED_AT" | "NAME";
export type StateOfApplicabilityFilter = {
  snapshotId?: string | null | undefined;
};
export type StateOfApplicabilityOrder = {
  direction: OrderDirection;
  field: StateOfApplicabilityOrderField;
};
export type StatesOfApplicabilityPagePaginationQuery$variables = {
  after?: string | null | undefined;
  before?: string | null | undefined;
  filter?: StateOfApplicabilityFilter | null | undefined;
  first?: number | null | undefined;
  id: string;
  last?: number | null | undefined;
  order?: StateOfApplicabilityOrder | null | undefined;
};
export type StatesOfApplicabilityPagePaginationQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"StatesOfApplicabilityPageFragment">;
  };
};
export type StatesOfApplicabilityPagePaginationQuery = {
  response: StatesOfApplicabilityPagePaginationQuery$data;
  variables: StatesOfApplicabilityPagePaginationQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "after"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "before"
},
v2 = {
  "defaultValue": {
    "snapshotId": null
  },
  "kind": "LocalArgument",
  "name": "filter"
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
  "name": "after",
  "variableName": "after"
},
v9 = {
  "kind": "Variable",
  "name": "before",
  "variableName": "before"
},
v10 = {
  "kind": "Variable",
  "name": "filter",
  "variableName": "filter"
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
    "name": "StatesOfApplicabilityPagePaginationQuery",
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
            "name": "StatesOfApplicabilityPageFragment"
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
    "name": "StatesOfApplicabilityPagePaginationQuery",
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
                "concreteType": "StateOfApplicabilityConnection",
                "kind": "LinkedField",
                "name": "statesOfApplicability",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "StateOfApplicabilityEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "StateOfApplicability",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v14/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "name",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "createdAt",
                            "storageKey": null
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:state-of-applicability:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:state-of-applicability:delete\")"
                          },
                          {
                            "alias": "statementsInfo",
                            "args": null,
                            "concreteType": "ApplicabilityStatementConnection",
                            "kind": "LinkedField",
                            "name": "applicabilityStatements",
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
                            "storageKey": null
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
                  "filter"
                ],
                "handle": "connection",
                "key": "StatesOfApplicabilityPage_statesOfApplicability",
                "kind": "LinkedHandle",
                "name": "statesOfApplicability"
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
    "cacheID": "812dcba28ad3fb441e49e281a199fd62",
    "id": null,
    "metadata": {},
    "name": "StatesOfApplicabilityPagePaginationQuery",
    "operationKind": "query",
    "text": "query StatesOfApplicabilityPagePaginationQuery(\n  $after: CursorKey = null\n  $before: CursorKey = null\n  $filter: StateOfApplicabilityFilter = {snapshotId: null}\n  $first: Int = 50\n  $last: Int = null\n  $order: StateOfApplicabilityOrder = {direction: DESC, field: CREATED_AT}\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...StatesOfApplicabilityPageFragment_4cFWzS\n    id\n  }\n}\n\nfragment StateOfApplicabilityRowFragment on StateOfApplicability {\n  id\n  name\n  createdAt\n  canDelete: permission(action: \"core:state-of-applicability:delete\")\n  statementsInfo: applicabilityStatements {\n    totalCount\n  }\n}\n\nfragment StatesOfApplicabilityPageFragment_4cFWzS on Organization {\n  statesOfApplicability(first: $first, after: $after, last: $last, before: $before, orderBy: $order, filter: $filter) {\n    edges {\n      node {\n        id\n        ...StateOfApplicabilityRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "30ea5e415eca53b251933d1ce086a9b7";

export default node;
