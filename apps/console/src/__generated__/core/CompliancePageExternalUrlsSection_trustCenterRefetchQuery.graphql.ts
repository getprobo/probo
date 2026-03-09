/**
 * @generated SignedSource<<1fdf3d464ce9b10a9dd2ea55a9dbfc26>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ComplianceExternalURLOrderField = "CREATED_AT" | "RANK";
export type OrderDirection = "ASC" | "DESC";
export type ComplianceExternalURLOrder = {
  direction: OrderDirection;
  field: ComplianceExternalURLOrderField;
};
export type CompliancePageExternalUrlsSection_trustCenterRefetchQuery$variables = {
  after?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  order?: ComplianceExternalURLOrder | null | undefined;
};
export type CompliancePageExternalUrlsSection_trustCenterRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageExternalUrlsSection_trustCenterFragment">;
  };
};
export type CompliancePageExternalUrlsSection_trustCenterRefetchQuery = {
  response: CompliancePageExternalUrlsSection_trustCenterRefetchQuery$data;
  variables: CompliancePageExternalUrlsSection_trustCenterRefetchQuery$variables;
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
    "name": "CompliancePageExternalUrlsSection_trustCenterRefetchQuery",
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
            "name": "CompliancePageExternalUrlsSection_trustCenterFragment"
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
    "name": "CompliancePageExternalUrlsSection_trustCenterRefetchQuery",
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
                "alias": "canUpdate",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:trust-center:update"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:trust-center:update\")"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "ComplianceExternalURLConnection",
                "kind": "LinkedField",
                "name": "externalUrls",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ComplianceExternalURLEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceExternalURL",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v8/*: any*/),
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
                            "name": "url",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "rank",
                            "storageKey": null
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
                "key": "CompliancePageExternalUrlsSection_externalUrls",
                "kind": "LinkedHandle",
                "name": "externalUrls"
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
    "cacheID": "0732bd06586ed31ed89b4dfad14d1fbd",
    "id": null,
    "metadata": {},
    "name": "CompliancePageExternalUrlsSection_trustCenterRefetchQuery",
    "operationKind": "query",
    "text": "query CompliancePageExternalUrlsSection_trustCenterRefetchQuery(\n  $after: CursorKey = null\n  $first: Int = 100\n  $order: ComplianceExternalURLOrder = {field: RANK, direction: ASC}\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...CompliancePageExternalUrlsSection_trustCenterFragment_gOVF5\n    id\n  }\n}\n\nfragment CompliancePageExternalUrlsSection_trustCenterFragment_gOVF5 on TrustCenter {\n  id\n  canUpdate: permission(action: \"core:trust-center:update\")\n  externalUrls(first: $first, after: $after, orderBy: $order) {\n    edges {\n      node {\n        id\n        name\n        url\n        rank\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f2614c45eaf65a9d6c6f857abc7c66f6";

export default node;
