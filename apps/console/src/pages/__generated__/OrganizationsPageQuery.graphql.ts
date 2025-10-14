/**
 * @generated SignedSource<<fd6d397e513dc681f055a502fa93a6b4>>
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
    readonly invitations: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly acceptedAt: any | null | undefined;
          readonly createdAt: any;
          readonly email: string;
          readonly expiresAt: any;
          readonly fullName: string;
          readonly id: string;
          readonly organization: {
            readonly id: string;
            readonly name: string;
          };
          readonly role: string;
        };
      }>;
    };
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
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v5 = {
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
v6 = {
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
},
v7 = [
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
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
            "storageKey": null
          },
          (v3/*: any*/)
        ],
        "storageKey": null
      },
      (v4/*: any*/)
    ],
    "storageKey": null
  },
  (v5/*: any*/),
  (v6/*: any*/)
],
v8 = {
  "kind": "Literal",
  "name": "filter",
  "value": {
    "status": "PENDING"
  }
},
v9 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "DESC",
    "field": "CREATED_AT"
  }
},
v10 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "InvitationEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Invitation",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
            "name": "expiresAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "acceptedAt",
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
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
          },
          (v3/*: any*/)
        ],
        "storageKey": null
      },
      (v4/*: any*/)
    ],
    "storageKey": null
  },
  (v5/*: any*/),
  (v6/*: any*/)
],
v11 = {
  "kind": "Literal",
  "name": "first",
  "value": 1000
},
v12 = [
  (v11/*: any*/),
  (v0/*: any*/)
],
v13 = [
  (v8/*: any*/),
  (v11/*: any*/),
  (v9/*: any*/)
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
            "selections": (v7/*: any*/),
            "storageKey": "__OrganizationsPage_organizations_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
          },
          {
            "alias": "invitations",
            "args": [
              (v8/*: any*/),
              (v9/*: any*/)
            ],
            "concreteType": "InvitationConnection",
            "kind": "LinkedField",
            "name": "__OrganizationsPage_invitations_connection",
            "plural": false,
            "selections": (v10/*: any*/),
            "storageKey": "__OrganizationsPage_invitations_connection(filter:{\"status\":\"PENDING\"},orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
            "args": (v12/*: any*/),
            "concreteType": "OrganizationConnection",
            "kind": "LinkedField",
            "name": "organizations",
            "plural": false,
            "selections": (v7/*: any*/),
            "storageKey": "organizations(first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"NAME\"})"
          },
          {
            "alias": null,
            "args": (v12/*: any*/),
            "filters": [
              "orderBy"
            ],
            "handle": "connection",
            "key": "OrganizationsPage_organizations",
            "kind": "LinkedHandle",
            "name": "organizations"
          },
          {
            "alias": null,
            "args": (v13/*: any*/),
            "concreteType": "InvitationConnection",
            "kind": "LinkedField",
            "name": "invitations",
            "plural": false,
            "selections": (v10/*: any*/),
            "storageKey": "invitations(filter:{\"status\":\"PENDING\"},first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          {
            "alias": null,
            "args": (v13/*: any*/),
            "filters": [
              "orderBy",
              "filter"
            ],
            "handle": "connection",
            "key": "OrganizationsPage_invitations",
            "kind": "LinkedHandle",
            "name": "invitations"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f8cd83b9b7ce0e1b43adc7aec05198bb",
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
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "viewer",
            "invitations"
          ]
        }
      ]
    },
    "name": "OrganizationsPageQuery",
    "operationKind": "query",
    "text": "query OrganizationsPageQuery {\n  viewer {\n    organizations(first: 1000, orderBy: {field: NAME, direction: ASC}) {\n      edges {\n        node {\n          id\n          name\n          logoUrl\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    invitations(first: 1000, orderBy: {field: CREATED_AT, direction: DESC}, filter: {status: PENDING}) {\n      edges {\n        node {\n          id\n          email\n          fullName\n          role\n          expiresAt\n          acceptedAt\n          createdAt\n          organization {\n            id\n            name\n          }\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d1957c93062d07020dc978b0f62bdf41";

export default node;
