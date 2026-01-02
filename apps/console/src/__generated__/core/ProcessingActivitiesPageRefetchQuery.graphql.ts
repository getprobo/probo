/**
 * @generated SignedSource<<504b0d0e997eb43eb8d370a21cb3e0fb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ProcessingActivitiesPageRefetchQuery$variables = {
  after?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  snapshotId?: string | null | undefined;
};
export type ProcessingActivitiesPageRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"ProcessingActivitiesPageFragment">;
  };
};
export type ProcessingActivitiesPageRefetchQuery = {
  response: ProcessingActivitiesPageRefetchQuery$data;
  variables: ProcessingActivitiesPageRefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "after"
},
v1 = {
  "defaultValue": 10,
  "kind": "LocalArgument",
  "name": "first"
},
v2 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v3 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "snapshotId"
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
  "kind": "Variable",
  "name": "snapshotId",
  "variableName": "snapshotId"
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v10 = [
  (v5/*: any*/),
  {
    "fields": [
      (v7/*: any*/)
    ],
    "kind": "ObjectValue",
    "name": "filter"
  },
  (v6/*: any*/)
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
    "name": "ProcessingActivitiesPageRefetchQuery",
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
              (v7/*: any*/)
            ],
            "kind": "FragmentSpread",
            "name": "ProcessingActivitiesPageFragment"
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
    "name": "ProcessingActivitiesPageRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v8/*: any*/),
          (v9/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v10/*: any*/),
                "concreteType": "ProcessingActivityConnection",
                "kind": "LinkedField",
                "name": "processingActivities",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "totalCount",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ProcessingActivityEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ProcessingActivity",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v9/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "snapshotId",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "sourceId",
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
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "purpose",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "dataSubjectCategory",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "personalDataCategory",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "lawfulBasis",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "location",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "internationalTransfers",
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
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "updatedAt",
                            "storageKey": null
                          },
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:processing-activity:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:processing-activity:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:processing-activity:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:processing-activity:delete\")"
                          },
                          (v8/*: any*/)
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
                        "name": "hasNextPage",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "endCursor",
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
                "args": (v10/*: any*/),
                "filters": [
                  "filter"
                ],
                "handle": "connection",
                "key": "ProcessingActivitiesPage_processingActivities",
                "kind": "LinkedHandle",
                "name": "processingActivities"
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
    "cacheID": "71c255d719fe0ffafda13169e925b4b8",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivitiesPageRefetchQuery",
    "operationKind": "query",
    "text": "query ProcessingActivitiesPageRefetchQuery(\n  $after: CursorKey\n  $first: Int = 10\n  $snapshotId: ID = null\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...ProcessingActivitiesPageFragment_35e0S5\n    id\n  }\n}\n\nfragment ProcessingActivitiesPageFragment_35e0S5 on Organization {\n  id\n  processingActivities(first: $first, after: $after, filter: {snapshotId: $snapshotId}) {\n    totalCount\n    edges {\n      node {\n        id\n        snapshotId\n        sourceId\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        lawfulBasis\n        location\n        internationalTransfers\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:processing-activity:update\")\n        canDelete: permission(action: \"core:processing-activity:delete\")\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "df348ac63bc56da06da97e484fb2c0e5";

export default node;
