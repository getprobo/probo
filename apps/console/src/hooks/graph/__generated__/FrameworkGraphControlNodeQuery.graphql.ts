/**
 * @generated SignedSource<<ec1b843d1afb3a82c747b8580b528fda>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ControlStatus = "EXCLUDED" | "INCLUDED";
export type FrameworkGraphControlNodeQuery$variables = {
  controlId: string;
};
export type FrameworkGraphControlNodeQuery$data = {
  readonly node: {
    readonly audits?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedAuditsCardFragment">;
        };
      }>;
    };
    readonly description?: string | null | undefined;
    readonly documents?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedDocumentsCardFragment">;
        };
      }>;
    };
    readonly exclusionJustification?: string | null | undefined;
    readonly id?: string;
    readonly measures?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedMeasuresCardFragment">;
        };
      }>;
    };
    readonly name?: string;
    readonly obligations?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedObligationsCardFragment">;
        };
      }>;
    };
    readonly sectionTitle?: string;
    readonly snapshots?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedSnapshotsCardFragment">;
        };
      }>;
    };
    readonly stateOfApplicabilityControls?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"LinkedStatesOfApplicabilityCardFragment">;
        };
      }>;
    };
    readonly status?: ControlStatus;
    readonly " $fragmentSpreads": FragmentRefs<"FrameworkControlDialogFragment">;
  };
};
export type FrameworkGraphControlNodeQuery = {
  response: FrameworkGraphControlNodeQuery$data;
  variables: FrameworkGraphControlNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "controlId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "controlId"
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
  "name": "sectionTitle",
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
  "name": "status",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "exclusionJustification",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v10 = {
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
v11 = {
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
v12 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v13 = [
  (v2/*: any*/),
  (v3/*: any*/)
],
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkGraphControlNodeQuery",
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
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "FrameworkControlDialogFragment"
              },
              {
                "alias": "stateOfApplicabilityControls",
                "args": null,
                "concreteType": "StateOfApplicabilityControlConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_stateOfApplicabilityControls_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "StateOfApplicabilityControlEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "StateOfApplicabilityControl",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedStatesOfApplicabilityCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": "measures",
                "args": null,
                "concreteType": "MeasureConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_measures_connection",
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
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedMeasuresCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": "documents",
                "args": null,
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_documents_connection",
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
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedDocumentsCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": "audits",
                "args": null,
                "concreteType": "AuditConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_audits_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AuditEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Audit",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedAuditsCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": "obligations",
                "args": null,
                "concreteType": "ObligationConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_obligations_connection",
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
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedObligationsCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": "snapshots",
                "args": null,
                "concreteType": "SnapshotConnection",
                "kind": "LinkedField",
                "name": "__FrameworkGraphControl_snapshots_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SnapshotEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Snapshot",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "LinkedSnapshotsCardFragment"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "type": "Control",
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
    "name": "FrameworkGraphControlNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v8/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "bestPractice",
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "concreteType": "StateOfApplicabilityControlConnection",
                "kind": "LinkedField",
                "name": "stateOfApplicabilityControls",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "StateOfApplicabilityControlEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "StateOfApplicabilityControl",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "stateOfApplicabilityId",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "controlId",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "StateOfApplicability",
                            "kind": "LinkedField",
                            "name": "stateOfApplicability",
                            "plural": false,
                            "selections": (v13/*: any*/),
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "applicability",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "justification",
                            "storageKey": null
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "stateOfApplicabilityControls(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_stateOfApplicabilityControls",
                "kind": "LinkedHandle",
                "name": "stateOfApplicabilityControls"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
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
                          (v3/*: any*/),
                          (v14/*: any*/),
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "measures(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_measures",
                "kind": "LinkedHandle",
                "name": "measures"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
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
                          (v15/*: any*/),
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
                                      (v6/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "versions(first:1)"
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_documents",
                "kind": "LinkedHandle",
                "name": "documents"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "concreteType": "AuditConnection",
                "kind": "LinkedField",
                "name": "audits",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AuditEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Audit",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v15/*: any*/),
                          (v14/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "validFrom",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "validUntil",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Framework",
                            "kind": "LinkedField",
                            "name": "framework",
                            "plural": false,
                            "selections": (v13/*: any*/),
                            "storageKey": null
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "audits(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_audits",
                "kind": "LinkedHandle",
                "name": "audits"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
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
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "People",
                            "kind": "LinkedField",
                            "name": "owner",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "fullName",
                                "storageKey": null
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "obligations(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_obligations",
                "kind": "LinkedHandle",
                "name": "obligations"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "concreteType": "SnapshotConnection",
                "kind": "LinkedField",
                "name": "snapshots",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SnapshotEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Snapshot",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "type",
                            "storageKey": null
                          },
                          (v15/*: any*/),
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v10/*: any*/),
                  (v11/*: any*/)
                ],
                "storageKey": "snapshots(first:100)"
              },
              {
                "alias": null,
                "args": (v12/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_snapshots",
                "kind": "LinkedHandle",
                "name": "snapshots"
              }
            ],
            "type": "Control",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f7773b474bf5f22e9a7075c685839001",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "stateOfApplicabilityControls"
          ]
        },
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
            "audits"
          ]
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "obligations"
          ]
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "snapshots"
          ]
        }
      ]
    },
    "name": "FrameworkGraphControlNodeQuery",
    "operationKind": "query",
    "text": "query FrameworkGraphControlNodeQuery(\n  $controlId: ID!\n) {\n  node(id: $controlId) {\n    __typename\n    ... on Control {\n      id\n      name\n      sectionTitle\n      description\n      status\n      exclusionJustification\n      ...FrameworkControlDialogFragment\n      stateOfApplicabilityControls(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedStatesOfApplicabilityCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      measures(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedMeasuresCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      documents(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedDocumentsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      audits(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedAuditsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      obligations(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedObligationsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      snapshots(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedSnapshotsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment FrameworkControlDialogFragment on Control {\n  id\n  name\n  description\n  sectionTitle\n  status\n  exclusionJustification\n  bestPractice\n}\n\nfragment LinkedAuditsCardFragment on Audit {\n  id\n  name\n  createdAt\n  state\n  validFrom\n  validUntil\n  framework {\n    id\n    name\n  }\n}\n\nfragment LinkedDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n\nfragment LinkedObligationsCardFragment on Obligation {\n  id\n  requirement\n  area\n  source\n  status\n  owner {\n    fullName\n    id\n  }\n}\n\nfragment LinkedSnapshotsCardFragment on Snapshot {\n  id\n  name\n  description\n  type\n  createdAt\n}\n\nfragment LinkedStatesOfApplicabilityCardFragment on StateOfApplicabilityControl {\n  id\n  stateOfApplicabilityId\n  controlId\n  stateOfApplicability {\n    id\n    name\n  }\n  applicability\n  justification\n}\n"
  }
};
})();

(node as any).hash = "b28838e3c3c4cdd68906247d129799c8";

export default node;
