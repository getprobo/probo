/**
 * @generated SignedSource<<272734160b60bf94a6f031e685751b64>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DataSensitivity = "CRITICAL" | "HIGH" | "LOW" | "MEDIUM" | "NONE";
export type DatumViewQuery$variables = {
  datumId: string;
  organizationId: string;
};
export type DatumViewQuery$data = {
  readonly node: {
    readonly createdAt?: string;
    readonly dataSensitivity?: DataSensitivity;
    readonly id?: string;
    readonly name?: string;
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly updatedAt?: string;
    readonly vendors?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly name: string;
        };
      }>;
    };
  };
  readonly organization: {
    readonly id?: string;
    readonly vendors?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly name: string;
        };
      }>;
    };
    readonly " $fragmentSpreads": FragmentRefs<"PeopleSelector_organization">;
  };
};
export type DatumViewQuery = {
  response: DatumViewQuery$data;
  variables: DatumViewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "datumId"
  },
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
    "variableName": "datumId"
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
  "name": "dataSensitivity",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
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
            (v2/*: any*/),
            (v3/*: any*/)
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v6/*: any*/)
  ],
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
v10 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v11 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "ASC",
    "field": "NAME"
  }
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v14 = {
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
v15 = [
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
          (v2/*: any*/),
          (v3/*: any*/),
          (v12/*: any*/)
        ],
        "storageKey": null
      },
      (v13/*: any*/)
    ],
    "storageKey": null
  },
  (v14/*: any*/)
],
v16 = {
  "kind": "Literal",
  "name": "first",
  "value": 100
},
v17 = [
  (v16/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "FULL_NAME"
    }
  }
],
v18 = [
  "orderBy"
],
v19 = [
  (v16/*: any*/),
  (v11/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DatumViewQuery",
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
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/)
            ],
            "type": "Datum",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v10/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "PeopleSelector_organization"
              },
              {
                "alias": "vendors",
                "args": [
                  (v11/*: any*/)
                ],
                "concreteType": "VendorConnection",
                "kind": "LinkedField",
                "name": "__DatumView_vendors_connection",
                "plural": false,
                "selections": (v15/*: any*/),
                "storageKey": "__DatumView_vendors_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
              }
            ],
            "type": "Organization",
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
    "name": "DatumViewQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v12/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/)
            ],
            "type": "Datum",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v10/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v12/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v17/*: any*/),
                "concreteType": "PeopleConnection",
                "kind": "LinkedField",
                "name": "peoples",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "PeopleEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "People",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "primaryEmailAddress",
                            "storageKey": null
                          },
                          (v12/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v13/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v14/*: any*/)
                ],
                "storageKey": "peoples(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "PeopleSelector_organization_peoples",
                "kind": "LinkedHandle",
                "name": "peoples"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "concreteType": "VendorConnection",
                "kind": "LinkedField",
                "name": "vendors",
                "plural": false,
                "selections": (v15/*: any*/),
                "storageKey": "vendors(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "DatumView_vendors",
                "kind": "LinkedHandle",
                "name": "vendors"
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
    "cacheID": "c3fd8609637a69d90a532344150b4d6a",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "organization",
            "vendors"
          ]
        }
      ]
    },
    "name": "DatumViewQuery",
    "operationKind": "query",
    "text": "query DatumViewQuery(\n  $datumId: ID!\n  $organizationId: ID!\n) {\n  node(id: $datumId) {\n    __typename\n    ... on Datum {\n      id\n      name\n      dataSensitivity\n      vendors {\n        edges {\n          node {\n            id\n            name\n          }\n        }\n      }\n      owner {\n        id\n        fullName\n      }\n      createdAt\n      updatedAt\n    }\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      ...PeopleSelector_organization\n      vendors(first: 100, orderBy: {direction: ASC, field: NAME}) {\n        edges {\n          node {\n            id\n            name\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment PeopleSelector_organization on Organization {\n  id\n  peoples(first: 100, orderBy: {direction: ASC, field: FULL_NAME}) {\n    edges {\n      node {\n        id\n        fullName\n        primaryEmailAddress\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2f759d45493db37e45b3a7166dd36c2e";

export default node;
