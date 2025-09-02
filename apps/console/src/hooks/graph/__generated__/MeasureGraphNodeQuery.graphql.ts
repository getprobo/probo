/**
 * @generated SignedSource<<c813fdce5ffc6c85f717413f4b87f0f4>>
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
    readonly category?: string;
    readonly controlsInfos?: {
      readonly totalCount: number;
    };
    readonly description?: string;
    readonly evidencesInfos?: {
      readonly totalCount: number;
    };
    readonly id?: string;
    readonly name?: string;
    readonly risksInfos?: {
      readonly totalCount: number;
    };
    readonly snapshotId?: string | null | undefined;
    readonly state?: MeasureState;
    readonly tasksInfos?: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MeasureControlsTabFragment" | "MeasureEvidencesTabFragment" | "MeasureFormDialogMeasureFragment" | "MeasureRisksTabFragment" | "MeasureTasksTabFragment">;
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
  "name": "state",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "category",
  "storageKey": null
},
v8 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 0
  }
],
v9 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
],
v10 = {
  "alias": "evidencesInfos",
  "args": (v8/*: any*/),
  "concreteType": "EvidenceConnection",
  "kind": "LinkedField",
  "name": "evidences",
  "plural": false,
  "selections": (v9/*: any*/),
  "storageKey": "evidences(first:0)"
},
v11 = {
  "alias": "risksInfos",
  "args": (v8/*: any*/),
  "concreteType": "RiskConnection",
  "kind": "LinkedField",
  "name": "risks",
  "plural": false,
  "selections": (v9/*: any*/),
  "storageKey": "risks(first:0)"
},
v12 = {
  "alias": "tasksInfos",
  "args": (v8/*: any*/),
  "concreteType": "TaskConnection",
  "kind": "LinkedField",
  "name": "tasks",
  "plural": false,
  "selections": (v9/*: any*/),
  "storageKey": "tasks(first:0)"
},
v13 = {
  "alias": "controlsInfos",
  "args": (v8/*: any*/),
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": (v9/*: any*/),
  "storageKey": "controls(first:0)"
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v15 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v17/*: any*/),
    (v18/*: any*/)
  ],
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
    (v17/*: any*/),
    (v18/*: any*/),
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
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureRisksTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "MeasureTasksTabFragment"
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
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              {
                "alias": null,
                "args": (v15/*: any*/),
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
                          (v4/*: any*/),
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
                          (v14/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "risks(first:100)"
              },
              {
                "alias": null,
                "args": (v15/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Measure__risks",
                "kind": "LinkedHandle",
                "name": "risks"
              },
              {
                "alias": null,
                "args": (v15/*: any*/),
                "concreteType": "TaskConnection",
                "kind": "LinkedField",
                "name": "tasks",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TaskEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Task",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v4/*: any*/),
                          (v6/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "timeEstimate",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "deadline",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "People",
                            "kind": "LinkedField",
                            "name": "assignedTo",
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
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Measure",
                            "kind": "LinkedField",
                            "name": "measure",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v14/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "tasks(first:100)"
              },
              {
                "alias": null,
                "args": (v15/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Measure__tasks",
                "kind": "LinkedHandle",
                "name": "tasks"
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
                          (v4/*: any*/),
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
                              (v4/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v14/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
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
                            "kind": "ScalarField",
                            "name": "filename",
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
                          (v14/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
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
    "cacheID": "265b74c861fec8f3101b207fdf9f6219",
    "id": null,
    "metadata": {},
    "name": "MeasureGraphNodeQuery",
    "operationKind": "query",
    "text": "query MeasureGraphNodeQuery(\n  $measureId: ID!\n) {\n  node(id: $measureId) {\n    __typename\n    ... on Measure {\n      id\n      snapshotId\n      name\n      description\n      state\n      category\n      evidencesInfos: evidences(first: 0) {\n        totalCount\n      }\n      risksInfos: risks(first: 0) {\n        totalCount\n      }\n      tasksInfos: tasks(first: 0) {\n        totalCount\n      }\n      controlsInfos: controls(first: 0) {\n        totalCount\n      }\n      ...MeasureRisksTabFragment\n      ...MeasureTasksTabFragment\n      ...MeasureControlsTabFragment\n      ...MeasureFormDialogMeasureFragment\n      ...MeasureEvidencesTabFragment\n    }\n    id\n  }\n}\n\nfragment LinkedControlsCardFragment on Control {\n  id\n  name\n  sectionTitle\n  framework {\n    id\n    name\n  }\n}\n\nfragment LinkedRisksCardFragment on Risk {\n  id\n  name\n  inherentRiskScore\n  residualRiskScore\n}\n\nfragment MeasureControlsTabFragment on Measure {\n  id\n  controls(first: 20) {\n    edges {\n      node {\n        id\n        ...LinkedControlsCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment on Measure {\n  id\n  evidences(first: 50) {\n    edges {\n      node {\n        id\n        filename\n        mimeType\n        ...MeasureEvidencesTabFragment_evidence\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment MeasureEvidencesTabFragment_evidence on Evidence {\n  id\n  filename\n  size\n  type\n  createdAt\n  mimeType\n}\n\nfragment MeasureFormDialogMeasureFragment on Measure {\n  id\n  description\n  name\n  category\n  state\n}\n\nfragment MeasureRisksTabFragment on Measure {\n  id\n  risks(first: 100) {\n    edges {\n      node {\n        id\n        ...LinkedRisksCardFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment MeasureTasksTabFragment on Measure {\n  tasks(first: 100) {\n    edges {\n      node {\n        id\n        name\n        state\n        description\n        ...TaskFormDialogFragment\n        assignedTo {\n          id\n          fullName\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment TaskFormDialogFragment on Task {\n  id\n  description\n  name\n  state\n  timeEstimate\n  deadline\n  assignedTo {\n    id\n  }\n  measure {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "fbb5b1c2b357020be17bdbf0398a15fc";

export default node;
