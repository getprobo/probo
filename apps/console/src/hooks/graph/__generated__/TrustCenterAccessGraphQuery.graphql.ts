/**
 * @generated SignedSource<<2333a6a7d5415f1a08a5612dcaceee8f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterAccessGraphQuery$variables = {
  trustCenterId: string;
};
export type TrustCenterAccessGraphQuery$data = {
  readonly node: {
    readonly accesses?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly cursor: any;
        readonly node: {
          readonly active: boolean;
          readonly createdAt: any;
          readonly email: string;
          readonly id: string;
          readonly name: string;
        };
      }>;
      readonly pageInfo: {
        readonly endCursor: any | null | undefined;
        readonly hasNextPage: boolean;
        readonly hasPreviousPage: boolean;
        readonly startCursor: any | null | undefined;
      };
    };
    readonly id?: string;
  };
};
export type TrustCenterAccessGraphQuery = {
  response: TrustCenterAccessGraphQuery$data;
  variables: TrustCenterAccessGraphQuery$variables;
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
    "direction": "DESC",
    "field": "CREATED_AT"
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
    "concreteType": "TrustCenterAccessEdge",
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
        "concreteType": "TrustCenterAccess",
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
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "active",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "createdAt",
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
    "name": "TrustCenterAccessGraphQuery",
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
                "alias": "accesses",
                "args": [
                  (v3/*: any*/)
                ],
                "concreteType": "TrustCenterAccessConnection",
                "kind": "LinkedField",
                "name": "__TrustCenterAccessTab_accesses_connection",
                "plural": false,
                "selections": (v5/*: any*/),
                "storageKey": "__TrustCenterAccessTab_accesses_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
    "name": "TrustCenterAccessGraphQuery",
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
                "concreteType": "TrustCenterAccessConnection",
                "kind": "LinkedField",
                "name": "accesses",
                "plural": false,
                "selections": (v5/*: any*/),
                "storageKey": "accesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "TrustCenterAccessTab_accesses",
                "kind": "LinkedHandle",
                "name": "accesses"
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
    "cacheID": "5b83cd4ae2434ce2e00de7230d264432",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "accesses"
          ]
        }
      ]
    },
    "name": "TrustCenterAccessGraphQuery",
    "operationKind": "query",
    "text": "query TrustCenterAccessGraphQuery(\n  $trustCenterId: ID!\n) {\n  node(id: $trustCenterId) {\n    __typename\n    ... on TrustCenter {\n      id\n      accesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        pageInfo {\n          hasNextPage\n          hasPreviousPage\n          startCursor\n          endCursor\n        }\n        edges {\n          cursor\n          node {\n            id\n            email\n            name\n            active\n            createdAt\n            __typename\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "af598fd2af198e63ed84fd618857a985";

export default node;
