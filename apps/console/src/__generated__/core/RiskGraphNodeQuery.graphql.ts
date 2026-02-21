/**
 * @generated SignedSource<<8159d097afc6c907c5c2e2f789bbd7e8>>
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
    readonly canCreateDocumentMapping?: boolean;
    readonly canCreateMeasureMapping?: boolean;
    readonly canCreateObligationMapping?: boolean;
    readonly canDelete?: boolean;
    readonly canDeleteDocumentMapping?: boolean;
    readonly canDeleteMeasureMapping?: boolean;
    readonly canDeleteObligationMapping?: boolean;
    readonly canUpdate?: boolean;
    readonly controlsInfo?: {
      readonly totalCount: number;
    };
    readonly description?: string | null | undefined;
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
    readonly obligationsInfo?: {
      readonly totalCount: number;
    };
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    } | null | undefined;
    readonly residualRiskScore?: number;
    readonly snapshotId?: string | null | undefined;
    readonly treatment?: RiskTreatment;
    readonly " $fragmentSpreads": FragmentRefs<"RiskControlsTabFragment" | "RiskDocumentsTabFragment" | "RiskMeasuresTabFragment" | "RiskObligationsTabFragment" | "RiskOverviewTabFragment" | "useRiskFormFragment">;
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
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v7/*: any*/)
  ],
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "note",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "inherentRiskScore",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "residualRiskScore",
  "storageKey": null
},
v12 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 0
  }
],
v13 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
],
v14 = {
  "alias": "measuresInfo",
  "args": (v12/*: any*/),
  "concreteType": "MeasureConnection",
  "kind": "LinkedField",
  "name": "measures",
  "plural": false,
  "selections": (v13/*: any*/),
  "storageKey": "measures(first:0)"
},
v15 = {
  "alias": "documentsInfo",
  "args": (v12/*: any*/),
  "concreteType": "DocumentConnection",
  "kind": "LinkedField",
  "name": "documents",
  "plural": false,
  "selections": (v13/*: any*/),
  "storageKey": "documents(first:0)"
},
v16 = {
  "alias": "controlsInfo",
  "args": (v12/*: any*/),
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v13/*: any*/),
  "storageKey": "controls(first:0)"
},
v17 = {
  "alias": "obligationsInfo",
  "args": (v12/*: any*/),
  "concreteType": "ObligationConnection",
  "kind": "LinkedField",
  "name": "obligations",
  "plural": false,
  "selections": (v13/*: any*/),
  "storageKey": "obligations(first:0)"
},
v18 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:update\")"
},
v19 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:delete\")"
},
v20 = {
  "alias": "canCreateDocumentMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:create-document-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:create-document-mapping\")"
},
v21 = {
  "alias": "canDeleteDocumentMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:delete-document-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:delete-document-mapping\")"
},
v22 = {
  "alias": "canCreateMeasureMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:create-measure-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:create-measure-mapping\")"
},
v23 = {
  "alias": "canDeleteMeasureMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:delete-measure-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:delete-measure-mapping\")"
},
v24 = {
  "alias": "canCreateObligationMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:create-obligation-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:create-obligation-mapping\")"
},
v25 = {
  "alias": "canDeleteObligationMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:risk:delete-obligation-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:risk:delete-obligation-mapping\")"
},
v26 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v27 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v28 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v29 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v30 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v31 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v29/*: any*/),
    (v30/*: any*/)
  ],
  "storageKey": null
},
v32 = {
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
v33 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v34 = [
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
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/),
              (v23/*: any*/),
              (v24/*: any*/),
              (v25/*: any*/),
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
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "RiskObligationsTabFragment"
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
          (v26/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/),
              (v20/*: any*/),
              (v21/*: any*/),
              (v22/*: any*/),
              (v23/*: any*/),
              (v24/*: any*/),
              (v25/*: any*/),
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
                "args": (v27/*: any*/),
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
                          (v26/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v28/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v31/*: any*/),
                  (v32/*: any*/)
                ],
                "storageKey": "measures(first:100)"
              },
              {
                "alias": null,
                "args": (v27/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__measures",
                "kind": "LinkedHandle",
                "name": "measures"
              },
              {
                "alias": null,
                "args": (v27/*: any*/),
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
                                      (v33/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "versions(first:1)"
                          },
                          (v26/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v28/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v31/*: any*/),
                  (v32/*: any*/)
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v27/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__documents",
                "kind": "LinkedHandle",
                "name": "documents"
              },
              {
                "alias": null,
                "args": (v34/*: any*/),
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
                          (v26/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v28/*: any*/)
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
                      (v29/*: any*/),
                      (v30/*: any*/),
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
                "args": (v34/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "RiskControlsTab_controls",
                "kind": "LinkedHandle",
                "name": "controls"
              },
              {
                "alias": null,
                "args": (v27/*: any*/),
                "concreteType": "ObligationConnection",
                "kind": "LinkedField",
                "name": "obligations",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ObligationEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Obligation",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "requirement",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "area",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "source",
                            "storageKey": null
                          },
                          (v33/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Profile",
                            "kind": "LinkedField",
                            "name": "owner",
                            "plural": false,
                            "selections": [
                              (v7/*: any*/),
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v26/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v28/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v31/*: any*/),
                  (v32/*: any*/)
                ],
                "storageKey": "obligations(first:100)"
              },
              {
                "alias": null,
                "args": (v27/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Risk__obligations",
                "kind": "LinkedHandle",
                "name": "obligations"
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
    "cacheID": "81ab255e9f6f1e3f666f5dd77920689b",
    "id": null,
    "metadata": {},
    "name": "RiskGraphNodeQuery",
    "operationKind": "query",
    "text": "query RiskGraphNodeQuery(\n  $riskId: ID!\n) {\n  node(id: $riskId) {\n    __typename\n    ... on Risk {\n      id\n      snapshotId\n      name\n      description\n      treatment\n      owner {\n        id\n        fullName\n      }\n      note\n      inherentRiskScore\n      residualRiskScore\n      measuresInfo: measures(first: 0) {\n        totalCount\n      }\n      documentsInfo: documents(first: 0) {\n        totalCount\n      }\n      controlsInfo: controls(first: 0) {\n        totalCount\n      }\n      obligationsInfo: obligations(first: 0) {\n        totalCount\n      }\n      canUpdate: permission(action: \"core:risk:update\")\n      canDelete: permission(action: \"core:risk:delete\")\n      canCreateDocumentMapping: permission(action: \"core:risk:create-document-mapping\")\n      canDeleteDocumentMapping: permission(action: \"core:risk:delete-document-mapping\")\n      canCreateMeasureMapping: permission(action: \"core:risk:create-measure-mapping\")\n      canDeleteMeasureMapping: permission(action: \"core:risk:delete-measure-mapping\")\n      canCreateObligationMapping: permission(action: \"core:risk:create-obligation-mapping\")\n      canDeleteObligationMapping: permission(action: \"core:risk:delete-obligation-mapping\")\n      ...useRiskFormFragment\n      ...RiskOverviewTabFragment\n      ...RiskMeasuresTabFragment\n      ...RiskDocumentsTabFragment\n      ...RiskControlsTabFragment\n      ...RiskObligationsTabFragment\n    }\n    id\n  }\n}\n\nfragment LinkedDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n\nfragment LinkedObligationsCardFragment on Obligation {\n  id\n  requirement\n  area\n  source\n  status\n  owner {\n    fullName\n    id\n  }\n}\n\nfragment RiskControlsTabFragment on Risk {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        sectionTitle\n        name\n        framework {\n          id\n          name\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment RiskDocumentsTabFragment on Risk {\n  id\n  documents(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedDocumentsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment RiskMeasuresTabFragment on Risk {\n  id\n  measures(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedMeasuresCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment RiskObligationsTabFragment on Risk {\n  id\n  obligations(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedObligationsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment RiskOverviewTabFragment on Risk {\n  inherentLikelihood\n  inherentImpact\n  residualLikelihood\n  residualImpact\n  inherentRiskScore\n  residualRiskScore\n}\n\nfragment useRiskFormFragment on Risk {\n  id\n  name\n  category\n  description\n  treatment\n  inherentLikelihood\n  inherentImpact\n  residualLikelihood\n  residualImpact\n  inherentRiskScore\n  residualRiskScore\n  note\n  owner {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "3ea74f7158cc3634d47af761c18dbadf";

export default node;
