/**
 * @generated SignedSource<<494ff8512e4a50c2bbc2cb3453dbab1c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED";
export type RiskGraphNodeQuery$variables = {
  riskId: string;
};
export type RiskGraphNodeQuery$data = {
  readonly node: {
    readonly controlsInfo?: {
      readonly totalCount: number;
    };
    readonly description?: string;
    readonly documentsInfo?: {
      readonly totalCount: number;
    };
    readonly id?: string;
    readonly inherentRiskScore?: number;
    readonly measuresInfo?: {
      readonly totalCount: number;
    };
    readonly name?: string;
    readonly note?: string;
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    } | null | undefined;
    readonly residualRiskScore?: number;
    readonly snapshotId?: string | null | undefined;
    readonly treatment?: RiskTreatment;
    readonly " $fragmentSpreads": FragmentRefs<"RiskControlsTabFragment" | "RiskDocumentsTabFragment" | "RiskMeasuresTabFragment" | "RiskOverviewTabFragment" | "useRiskFormFragment">;
  };
};
export type RiskGraphNodeQuery = {
  response: RiskGraphNodeQuery$data;
  variables: RiskGraphNodeQuery$variables;
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
  "name": "snapshotId",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "treatment",
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
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "note",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "inherentRiskScore",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "residualRiskScore",
  "storageKey": null
},
v11 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 0
  }
],
v12 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
],
v13 = {
  "alias": "measuresInfo",
  "args": (v11/*: any*/),
  "concreteType": "MeasureConnection",
  "kind": "LinkedField",
  "name": "measures",
  "plural": false,
  "selections": (v12/*: any*/),
  "storageKey": "measures(first:0)"
},
v14 = {
  "alias": "documentsInfo",
  "args": (v11/*: any*/),
  "concreteType": "DocumentConnection",
  "kind": "LinkedField",
  "name": "documents",
  "plural": false,
  "selections": (v12/*: any*/),
  "storageKey": "documents(first:0)"
},
v15 = {
  "alias": "controlsInfo",
  "args": (v11/*: any*/),
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v12/*: any*/),
  "storageKey": "controls(first:0)"
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v17 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v19/*: any*/),
    (v20/*: any*/)
  ],
  "storageKey": null
},
v22 = {
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
v23 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "RiskGraphNodeQuery",
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
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "useRiskFormFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "RiskOverviewTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "RiskMeasuresTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "RiskDocumentsTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "RiskControlsTabFragment"
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
    "name": "RiskGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v16/*: any*/),
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
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
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
                "name": "inherentLikelihood",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "inherentImpact",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "residualLikelihood",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "residualImpact",
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "concreteType": "MeasureConnection",
                "kind": "LinkedField",
                "name": "measures",
                "plural": false,
                "selections": [
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
                          (v4/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "state",
                            "storageKey": null
                          },
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v21/*: any*/),
                  (v22/*: any*/)
                ],
                "storageKey": "measures(first:100)"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__measures",
                "kind": "LinkedHandle",
                "name": "measures"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": [
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
                            "name": "documentType",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "first",
                                "value": 1
                              }
                            ],
                            "concreteType": "DocumentVersionConnection",
                            "kind": "LinkedField",
                            "name": "versions",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "DocumentVersionEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "DocumentVersion",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v2/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "status",
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "versions(first:1)"
                          },
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v21/*: any*/),
                  (v22/*: any*/)
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v17/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__documents",
                "kind": "LinkedHandle",
                "name": "documents"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
                "concreteType": "ControlConnection",
                "kind": "LinkedField",
                "name": "controls",
                "plural": false,
                "selections": [
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
                          (v4/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Framework",
                            "kind": "LinkedField",
                            "name": "framework",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v4/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
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
                      (v19/*: any*/),
                      (v20/*: any*/),
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
                  }
                ],
                "storageKey": "controls(first:20)"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "RiskControlsTab_controls",
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
    "cacheID": "03ef938803c52b572e23e691d50c4188",
    "id": null,
    "metadata": {},
    "name": "RiskGraphNodeQuery",
    "operationKind": "query",
    "text": "query RiskGraphNodeQuery(\n  $riskId: ID!\n) {\n  node(id: $riskId) {\n    __typename\n    ... on Risk {\n      id\n      snapshotId\n      name\n      description\n      treatment\n      owner {\n        id\n        fullName\n      }\n      note\n      inherentRiskScore\n      residualRiskScore\n      measuresInfo: measures(first: 0) {\n        totalCount\n      }\n      documentsInfo: documents(first: 0) {\n        totalCount\n      }\n      controlsInfo: controls(first: 0) {\n        totalCount\n      }\n      ...useRiskFormFragment\n      ...RiskOverviewTabFragment\n      ...RiskMeasuresTabFragment\n      ...RiskDocumentsTabFragment\n      ...RiskControlsTabFragment\n    }\n    id\n  }\n}\n\nfragment LinkedDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n\nfragment RiskControlsTabFragment on Risk {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        sectionTitle\n        name\n        framework {\n          id\n          name\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment RiskDocumentsTabFragment on Risk {\n  id\n  documents(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedDocumentsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment RiskMeasuresTabFragment on Risk {\n  id\n  measures(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedMeasuresCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment RiskOverviewTabFragment on Risk {\n  inherentLikelihood\n  inherentImpact\n  residualLikelihood\n  residualImpact\n  inherentRiskScore\n  residualRiskScore\n}\n\nfragment useRiskFormFragment on Risk {\n  id\n  name\n  category\n  description\n  treatment\n  inherentLikelihood\n  inherentImpact\n  residualLikelihood\n  residualImpact\n  note\n  owner {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "43dcdf60f1d1b28414650fe5f9294faa";

export default node;
