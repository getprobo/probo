/**
 * @generated SignedSource<<13d2dd64b5963f27b35077c80106abfb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AssetOrderField = "AMOUNT" | "CREATED_AT";
export type OrderDirection = "ASC" | "DESC";
export type AssetOrder = {
  direction: OrderDirection;
  field: AssetOrderField;
};
export type AssetsListQuery$variables = {
  after?: string | null | undefined;
  before?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  last?: number | null | undefined;
  orderBy?: AssetOrder | null | undefined;
  snapshotId?: string | null | undefined;
};
export type AssetsListQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"AssetsPageFragment">;
  };
};
export type AssetsListQuery = {
  response: AssetsListQuery$data;
  variables: AssetsListQuery$variables;
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
  "defaultValue": 10,
  "kind": "LocalArgument",
  "name": "first"
},
v3 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v4 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "last"
},
v5 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "orderBy"
},
v6 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "snapshotId"
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
  "name": "first",
  "variableName": "first"
},
v11 = {
  "kind": "Variable",
  "name": "last",
  "variableName": "last"
},
v12 = {
  "kind": "Variable",
  "name": "orderBy",
  "variableName": "orderBy"
},
v13 = {
  "kind": "Variable",
  "name": "snapshotId",
  "variableName": "snapshotId"
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v16 = [
  (v8/*: any*/),
  (v9/*: any*/),
  {
    "fields": [
      (v13/*: any*/)
    ],
    "kind": "ObjectValue",
    "name": "filter"
  },
  (v10/*: any*/),
  (v11/*: any*/),
  (v12/*: any*/)
],
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
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
    "name": "AssetsListQuery",
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
              (v13/*: any*/)
            ],
            "kind": "FragmentSpread",
            "name": "AssetsPageFragment"
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
      (v4/*: any*/),
      (v5/*: any*/),
      (v6/*: any*/),
      (v3/*: any*/)
    ],
    "kind": "Operation",
    "name": "AssetsListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v7/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v14/*: any*/),
          (v15/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v16/*: any*/),
                "concreteType": "AssetConnection",
                "kind": "LinkedField",
                "name": "assets",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AssetEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Asset",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v15/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "snapshotId",
                            "storageKey": null
                          },
                          (v17/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "amount",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "assetType",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "dataTypesStored",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Profile",
                            "kind": "LinkedField",
                            "name": "owner",
                            "plural": false,
                            "selections": [
                              (v15/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "fullName",
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "first",
                                "value": 50
                              }
                            ],
                            "concreteType": "VendorConnection",
                            "kind": "LinkedField",
                            "name": "vendors",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "VendorEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "Vendor",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v15/*: any*/),
                                      (v17/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "websiteUrl",
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "vendors(first:50)"
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "createdAt",
                            "storageKey": null
                          },
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:asset:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:asset:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:asset:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:asset:delete\")"
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
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v16/*: any*/),
                "filters": [
                  "filter"
                ],
                "handle": "connection",
                "key": "AssetsPage_assets",
                "kind": "LinkedHandle",
                "name": "assets"
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
    "cacheID": "350938b31ac1c7f7a28849123995da3a",
    "id": null,
    "metadata": {},
    "name": "AssetsListQuery",
    "operationKind": "query",
    "text": "query AssetsListQuery(\n  $after: CursorKey = null\n  $before: CursorKey = null\n  $first: Int = 10\n  $last: Int = null\n  $orderBy: AssetOrder = null\n  $snapshotId: ID = null\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...AssetsPageFragment_38J9tm\n    id\n  }\n}\n\nfragment AssetsPageFragment_38J9tm on Organization {\n  assets(first: $first, after: $after, last: $last, before: $before, orderBy: $orderBy, filter: {snapshotId: $snapshotId}) {\n    edges {\n      node {\n        id\n        snapshotId\n        name\n        amount\n        assetType\n        dataTypesStored\n        owner {\n          id\n          fullName\n        }\n        vendors(first: 50) {\n          edges {\n            node {\n              id\n              name\n              websiteUrl\n            }\n          }\n        }\n        createdAt\n        canUpdate: permission(action: \"core:asset:update\")\n        canDelete: permission(action: \"core:asset:delete\")\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n"
  }
};
})();

(node as any).hash = "4e870b93c801e9f0b6e26ab6069863fe";

export default node;
