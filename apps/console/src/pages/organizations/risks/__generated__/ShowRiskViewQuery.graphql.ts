/**
 * @generated SignedSource<<b49e4f8dc994a66c6e6c9450126028f5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED";
export type ShowRiskViewQuery$variables = {
  riskId: string;
};
export type ShowRiskViewQuery$data = {
  readonly node: {
    readonly controls?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: string;
          readonly description: string;
          readonly id: string;
          readonly name: string;
          readonly sectionTitle: string;
        };
      }>;
    };
    readonly createdAt?: string;
    readonly description?: string;
    readonly documents?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: string;
          readonly id: string;
          readonly title: string;
        };
      }>;
    };
    readonly id: string;
    readonly inherentImpact?: number;
    readonly inherentLikelihood?: number;
    readonly measures?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: string;
          readonly createdAt: string;
          readonly description: string;
          readonly id: string;
          readonly name: string;
          readonly state: MeasureState;
        };
      }>;
    };
    readonly name?: string;
    readonly note?: string;
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    } | null | undefined;
    readonly residualImpact?: number;
    readonly residualLikelihood?: number;
    readonly treatment?: RiskTreatment;
    readonly updatedAt?: string;
  };
};
export type ShowRiskViewQuery = {
  response: ShowRiskViewQuery$data;
  variables: ShowRiskViewQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "riskId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "riskId"
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
  "name": "description",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "treatment",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v2/*: any*/),
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
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "inherentLikelihood",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "inherentImpact",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "residualLikelihood",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "residualImpact",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "note",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
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
  "name": "cursor",
  "storageKey": null
},
v16 = {
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
v17 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "MeasureEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Measure",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "category",
            "storageKey": null
          },
          (v12/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "state",
            "storageKey": null
          },
          (v14/*: any*/)
        ],
        "storageKey": null
      },
      (v15/*: any*/)
    ],
    "storageKey": null
  },
  (v16/*: any*/)
],
v18 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "DocumentEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Document",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "title",
            "storageKey": null
          },
          (v12/*: any*/),
          (v14/*: any*/)
        ],
        "storageKey": null
      },
      (v15/*: any*/)
    ],
    "storageKey": null
  },
  (v16/*: any*/)
],
v19 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "ControlEdge",
    "kind": "LinkedField",
    "name": "edges",
    "plural": true,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Control",
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "sectionTitle",
            "storageKey": null
          },
          (v3/*: any*/),
          (v4/*: any*/),
          (v12/*: any*/),
          (v14/*: any*/)
        ],
        "storageKey": null
      },
      (v15/*: any*/)
    ],
    "storageKey": null
  },
  (v16/*: any*/)
],
v20 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ShowRiskViewQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "alias": "measures",
                "args": null,
                "concreteType": "MeasureConnection",
                "kind": "LinkedField",
                "name": "__Risk__measures_connection",
                "plural": false,
                "selections": (v17/*: any*/),
                "storageKey": null
              },
              {
                "alias": "documents",
                "args": null,
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "__Risk__documents_connection",
                "plural": false,
                "selections": (v18/*: any*/),
                "storageKey": null
              },
              {
                "alias": "controls",
                "args": null,
                "concreteType": "ControlConnection",
                "kind": "LinkedField",
                "name": "__Risk__controls_connection",
                "plural": false,
                "selections": (v19/*: any*/),
                "storageKey": null
              }
            ],
            "type": "Risk",
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
    "name": "ShowRiskViewQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v14/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "alias": null,
                "args": (v20/*: any*/),
                "concreteType": "MeasureConnection",
                "kind": "LinkedField",
                "name": "measures",
                "plural": false,
                "selections": (v17/*: any*/),
                "storageKey": "measures(first:100)"
              },
              {
                "alias": null,
                "args": (v20/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__measures",
                "kind": "LinkedHandle",
                "name": "measures"
              },
              {
                "alias": null,
                "args": (v20/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": (v18/*: any*/),
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v20/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__documents",
                "kind": "LinkedHandle",
                "name": "documents"
              },
              {
                "alias": null,
                "args": (v20/*: any*/),
                "concreteType": "ControlConnection",
                "kind": "LinkedField",
                "name": "controls",
                "plural": false,
                "selections": (v19/*: any*/),
                "storageKey": "controls(first:100)"
              },
              {
                "alias": null,
                "args": (v20/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__controls",
                "kind": "LinkedHandle",
                "name": "controls"
              }
            ],
            "type": "Risk",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "fe20c6f4cd1e70c14bb099af7ecfffc2",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "measures"
          ]
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "documents"
          ]
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "controls"
          ]
        }
      ]
    },
    "name": "ShowRiskViewQuery",
    "operationKind": "query",
    "text": "query ShowRiskViewQuery(\n  $riskId: ID!\n) {\n  node(id: $riskId) {\n    __typename\n    id\n    ... on Risk {\n      name\n      description\n      treatment\n      owner {\n        id\n        fullName\n      }\n      inherentLikelihood\n      inherentImpact\n      residualLikelihood\n      residualImpact\n      note\n      createdAt\n      updatedAt\n      measures(first: 100) {\n        edges {\n          node {\n            id\n            name\n            description\n            category\n            createdAt\n            state\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      documents(first: 100) {\n        edges {\n          node {\n            id\n            title\n            createdAt\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      controls(first: 100) {\n        edges {\n          node {\n            id\n            sectionTitle\n            name\n            description\n            createdAt\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f1695786e4eb823466d6ebee4555cd9b";

export default node;
