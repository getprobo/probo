/**
 * @generated SignedSource<<bf973ac1452f812480abe7160f4529e8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type MeasureGraphNodeQuery$variables = {
  measureId: string;
};
export type MeasureGraphNodeQuery$data = {
  readonly node: {
    readonly canDelete?: boolean;
    readonly canListTasks?: boolean;
    readonly canUpdate?: boolean;
    readonly category?: string;
    readonly controlsInfos?: {
      readonly totalCount: number;
    };
    readonly description?: string | null | undefined;
    readonly evidencesInfos?: {
      readonly totalCount: number;
    };
    readonly id?: string;
    readonly name?: string;
    readonly risksInfos?: {
      readonly totalCount: number;
    };
    readonly state?: MeasureState;
    readonly " $fragmentSpreads": FragmentRefs<"MeasureControlsTabFragment" | "MeasureEvidencesTabFragment" | "MeasureFormDialogMeasureFragment" | "MeasureRisksTabFragment">;
  };
};
export type MeasureGraphNodeQuery = {
  response: MeasureGraphNodeQuery$data;
  variables: MeasureGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "measureId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "measureId"
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
  "name": "state",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "category",
  "storageKey": null
},
v7 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:measure:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:measure:update\")"
},
v8 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:measure:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:measure:delete\")"
},
v9 = {
  "alias": "canListTasks",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:task:list"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:task:list\")"
},
v10 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 0
  }
],
v11 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
],
v12 = {
  "alias": "evidencesInfos",
  "args": (v10/*: any*/),
  "concreteType": "EvidenceConnection",
  "kind": "LinkedField",
  "name": "evidences",
  "plural": false,
  "selections": (v11/*: any*/),
  "storageKey": "evidences(first:0)"
},
v13 = {
  "alias": "risksInfos",
  "args": (v10/*: any*/),
  "concreteType": "RiskConnection",
  "kind": "LinkedField",
  "name": "risks",
  "plural": false,
  "selections": (v11/*: any*/),
  "storageKey": "risks(first:0)"
},
v14 = {
  "alias": "controlsInfos",
  "args": (v10/*: any*/),
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v11/*: any*/),
  "storageKey": "controls(first:0)"
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v16 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v20 = {
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
v21 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
],
v22 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v18/*: any*/),
    (v19/*: any*/),
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
v23 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureGraphNodeQuery",
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
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureRisksTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureControlsTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureFormDialogMeasureFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureEvidencesTabFragment"
              }
            ],
            "type": "Measure",
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
    "name": "MeasureGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v15/*: any*/),
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
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              {
                "alias": null,
                "args": (v16/*: any*/),
                "concreteType": "RiskConnection",
                "kind": "LinkedField",
                "name": "risks",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "RiskEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Risk",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
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
                          {
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
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "inherentRiskScore",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "residualRiskScore",
                            "storageKey": null
                          },
                          (v15/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/)
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
                      (v18/*: any*/),
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/)
                ],
                "storageKey": "risks(first:100)"
              },
              {
                "alias": null,
                "args": (v16/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Measure__risks",
                "kind": "LinkedHandle",
                "name": "risks"
              },
              {
                "alias": null,
                "args": (v21/*: any*/),
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
                            "alias": "canCreateMeasureMapping",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:control:create-measure-mapping"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:control:create-measure-mapping\")"
                          },
                          {
                            "alias": "canDeleteMeasureMapping",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:control:delete-measure-mapping"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:control:delete-measure-mapping\")"
                          },
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "sectionTitle",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Framework",
                            "kind": "LinkedField",
                            "name": "framework",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v3/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v15/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v22/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "controls(first:20)"
              },
              {
                "alias": null,
                "args": (v21/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "MeasureControlsTab_controls",
                "kind": "LinkedHandle",
                "name": "controls"
              },
              {
                "alias": "canUploadEvidence",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:measure:upload-evidence"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:measure:upload-evidence\")"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
                "concreteType": "EvidenceConnection",
                "kind": "LinkedField",
                "name": "evidences",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "EvidenceEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Evidence",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "File",
                            "kind": "LinkedField",
                            "name": "file",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "fileName",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "mimeType",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "size",
                                "storageKey": null
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "type",
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
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:evidence:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:evidence:delete\")"
                          },
                          (v15/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v22/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "evidences(first:50)"
              },
              {
                "alias": null,
                "args": (v23/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "MeasureEvidencesTabFragment_evidences",
                "kind": "LinkedHandle",
                "name": "evidences"
              }
            ],
            "type": "Measure",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "d8ac1210636ffd91eacb1bfe3452e26c",
    "id": null,
    "metadata": {},
    "name": "MeasureGraphNodeQuery",
    "operationKind": "query",
    "text": "query MeasureGraphNodeQuery(\n  $measureId: ID!\n) {\n  node(id: $measureId) {\n    __typename\n    ... on Measure {\n      id\n      name\n      description\n      state\n      category\n      canUpdate: permission(action: \"core:measure:update\")\n      canDelete: permission(action: \"core:measure:delete\")\n      canListTasks: permission(action: \"core:task:list\")\n      evidencesInfos: evidences(first: 0) {\n        totalCount\n      }\n      risksInfos: risks(first: 0) {\n        totalCount\n      }\n      controlsInfos: controls(first: 0) {\n        totalCount\n      }\n      ...MeasureRisksTabFragment\n      ...MeasureControlsTabFragment\n      ...MeasureFormDialogMeasureFragment\n      ...MeasureEvidencesTabFragment\n    }\n    id\n  }\n}\n\nfragment LinkedControlsCardFragment on Control {\n  id\n  name\n  sectionTitle\n  framework {\n    id\n    name\n  }\n}\n\nfragment LinkedRisksCardFragment on Risk {\n  id\n  name\n  inherentRiskScore\n  residualRiskScore\n}\n\nfragment MeasureControlsTabFragment on Measure {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        canCreateMeasureMapping: permission(action: \"core:control:create-measure-mapping\")\n        canDeleteMeasureMapping: permission(action: \"core:control:delete-measure-mapping\")\n        ...LinkedControlsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment on Measure {\n  id\n  canUploadEvidence: permission(action: \"core:measure:upload-evidence\")\n  evidences(first: 50) {\n    edges {\n      node {\n        id\n        file {\n          fileName\n          mimeType\n          size\n          id\n        }\n        ...MeasureEvidencesTabFragment_evidence\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment_evidence on Evidence {\n  id\n  file {\n    fileName\n    mimeType\n    size\n    id\n  }\n  type\n  createdAt\n  canDelete: permission(action: \"core:evidence:delete\")\n}\n\nfragment MeasureFormDialogMeasureFragment on Measure {\n  id\n  description\n  name\n  category\n  state\n}\n\nfragment MeasureRisksTabFragment on Measure {\n  id\n  risks(first: 100) {\n    edges {\n      node {\n        id\n        canCreateMeasureMapping: permission(action: \"core:risk:create-measure-mapping\")\n        canDeleteMeasureMapping: permission(action: \"core:risk:delete-measure-mapping\")\n        ...LinkedRisksCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "0d55da2c159f16130dd3aaf232103b0f";

export default node;
