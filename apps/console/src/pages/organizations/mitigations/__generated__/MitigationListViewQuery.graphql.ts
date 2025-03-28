/**
 * @generated SignedSource<<c919ae742bdab52e5e68fb5c7fe92812>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MitigationImportance = "ADVANCED" | "MANDATORY" | "PREFERRED";
export type MitigationState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type MitigationListViewQuery$variables = {
  first?: number | null | undefined;
  organizationId: string;
};
export type MitigationListViewQuery$data = {
  readonly organization: {
    readonly id: string;
    readonly mitigations?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: string;
          readonly createdAt: string;
          readonly description: string;
          readonly id: string;
          readonly importance: MitigationImportance;
          readonly name: string;
          readonly state: MitigationState;
          readonly updatedAt: string;
        };
      }>;
    };
  };
};
export type MitigationListViewQuery = {
  response: MitigationListViewQuery$data;
  variables: MitigationListViewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "first"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "organizationId"
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "ASC",
    "field": "CREATED_AT"
  }
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v6 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "MitigationEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Mitigation",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/),
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
            "name": "category",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "state",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "importance",
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
          (v5/*: any*/)
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
  }
],
v7 = [
  {
    "kind": "Variable",
    "name": "first",
    "variableName": "first"
  },
  (v4/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "MitigationListViewQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": "mitigations",
                "args": [
                  (v4/*: any*/)
                ],
                "concreteType": "MitigationConnection",
                "kind": "LinkedField",
                "name": "__MitigationListView_mitigations_connection",
                "plural": false,
                "selections": (v6/*: any*/),
                "storageKey": "__MitigationListView_mitigations_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "MitigationListViewQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v7/*: any*/),
                "concreteType": "MitigationConnection",
                "kind": "LinkedField",
                "name": "mitigations",
                "plural": false,
                "selections": (v6/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "MitigationListView_mitigations",
                "kind": "LinkedHandle",
                "name": "mitigations"
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
    "cacheID": "cb49e5e7740ec71ce0b8e1abbdbcf24d",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": "first",
          "cursor": null,
          "direction": "forward",
          "path": [
            "organization",
            "mitigations"
          ]
        }
      ]
    },
    "name": "MitigationListViewQuery",
    "operationKind": "query",
    "text": "query MitigationListViewQuery(\n  $organizationId: ID!\n  $first: Int\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    id\n    ... on Organization {\n      mitigations(first: $first, orderBy: {direction: ASC, field: CREATED_AT}) {\n        edges {\n          node {\n            id\n            name\n            description\n            category\n            state\n            importance\n            createdAt\n            updatedAt\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ee7fdf984d5cb376068ba53d0a56c59b";

export default node;
