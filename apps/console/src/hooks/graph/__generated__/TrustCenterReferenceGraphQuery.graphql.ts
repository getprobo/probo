/**
 * @generated SignedSource<<85556d15b776b3e61e312057a994d482>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterReferenceGraphQuery$variables = {
  trustCenterId: string;
};
export type TrustCenterReferenceGraphQuery$data = {
  readonly node: {
    readonly id?: string;
    readonly references?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly cursor: any;
        readonly node: {
          readonly createdAt: any;
          readonly description: string;
          readonly id: string;
          readonly logoUrl: string;
          readonly name: string;
          readonly rank: number;
          readonly updatedAt: any;
          readonly websiteUrl: string;
        };
      }>;
      readonly pageInfo: {
        readonly endCursor: any | null | undefined;
        readonly hasNextPage: boolean;
        readonly hasPreviousPage: boolean;
        readonly startCursor: any | null | undefined;
      };
    };
  };
};
export type TrustCenterReferenceGraphQuery = {
  response: TrustCenterReferenceGraphQuery$data;
  variables: TrustCenterReferenceGraphQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "trustCenterId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "trustCenterId"
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
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "ASC",
    "field": "RANK"
  }
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v5 = [
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
        "name": "hasPreviousPage",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "startCursor",
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
    "alias": null,
    "args": null,
    "concreteType": "TrustCenterReferenceEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "cursor",
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterReference",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
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
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "websiteUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
            "storageKey": null
          },
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
          (v4/*: any*/)
        ],
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
v6 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  (v3/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterReferenceGraphQuery",
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
              {
                "alias": "references",
                "args": [
                  (v3/*: any*/)
                ],
                "concreteType": "TrustCenterReferenceConnection",
                "kind": "LinkedField",
                "name": "__TrustCenterReferencesSection_references_connection",
                "plural": false,
                "selections": (v5/*: any*/),
                "storageKey": "__TrustCenterReferencesSection_references_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
              }
            ],
            "type": "TrustCenter",
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
    "name": "TrustCenterReferenceGraphQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "TrustCenterReferenceConnection",
                "kind": "LinkedField",
                "name": "references",
                "plural": false,
                "selections": (v5/*: any*/),
                "storageKey": "references(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "TrustCenterReferencesSection_references",
                "kind": "LinkedHandle",
                "name": "references"
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
    "cacheID": "8fdd97abcaa0d9eb889eb454c85f2e4c",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "references"
          ]
        }
      ]
    },
    "name": "TrustCenterReferenceGraphQuery",
    "operationKind": "query",
    "text": "query TrustCenterReferenceGraphQuery(\n  $trustCenterId: ID!\n) {\n  node(id: $trustCenterId) {\n    __typename\n    ... on TrustCenter {\n      id\n      references(first: 100, orderBy: {field: RANK, direction: ASC}) {\n        pageInfo {\n          hasNextPage\n          hasPreviousPage\n          startCursor\n          endCursor\n        }\n        edges {\n          cursor\n          node {\n            id\n            name\n            description\n            websiteUrl\n            logoUrl\n            rank\n            createdAt\n            updatedAt\n            __typename\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "cb4b9b8a68249ae8066a3ff2bfc6acd8";

export default node;
