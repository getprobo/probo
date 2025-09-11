/**
 * @generated SignedSource<<84d0e43eced78862d1c82603e9c332e8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationsPageQuery$variables = Record<PropertyKey, never>;
export type OrganizationsPageQuery$data = {
  readonly viewer: {
    readonly organizations: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly logoUrl: string | null | undefined;
          readonly name: string;
        };
      }>;
    };
  };
};
export type OrganizationsPageQuery = {
  response: OrganizationsPageQuery$data;
  variables: OrganizationsPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "ASC",
    "field": "NAME"
  }
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "OrganizationEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Organization",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
            "name": "logoUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          }
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
v3 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  (v0/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationsPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": "organizations",
            "args": [
              (v0/*: any*/)
            ],
            "concreteType": "OrganizationConnection",
            "kind": "LinkedField",
            "name": "__OrganizationsPage_organizations_connection",
            "plural": false,
            "selections": (v2/*: any*/),
            "storageKey": "__OrganizationsPage_organizations_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
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
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationsPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": (v3/*: any*/),
            "concreteType": "OrganizationConnection",
            "kind": "LinkedField",
            "name": "organizations",
            "plural": false,
            "selections": (v2/*: any*/),
            "storageKey": "organizations(first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
          },
          {
            "alias": null,
            "args": (v3/*: any*/),
            "filters": [
              "orderBy"
            ],
            "handle": "connection",
            "key": "OrganizationsPage_organizations",
            "kind": "LinkedHandle",
            "name": "organizations"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1735764e6816660969c5f96922320ac5",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "viewer",
            "organizations"
          ]
        }
      ]
    },
    "name": "OrganizationsPageQuery",
    "operationKind": "query",
    "text": "query OrganizationsPageQuery {\n  viewer {\n    organizations(first: 1000, orderBy: {field: NAME, direction: ASC}) {\n      edges {\n        node {\n          id\n          name\n          logoUrl\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "6fde39384e4678f88f17c34bbd30e684";

export default node;
