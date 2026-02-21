/**
 * @generated SignedSource<<72da3c2c267ec4fd2da7aee87f38c51b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
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
    readonly canCreateAuditMapping?: boolean;
    readonly canCreateDocumentMapping?: boolean;
    readonly canCreateMeasureMapping?: boolean;
    readonly canCreateObligationMapping?: boolean;
    readonly canCreateSnapshotMapping?: boolean;
    readonly canDelete?: boolean;
    readonly canDeleteAuditMapping?: boolean;
    readonly canDeleteDocumentMapping?: boolean;
    readonly canDeleteMeasureMapping?: boolean;
    readonly canDeleteObligationMapping?: boolean;
    readonly canDeleteSnapshotMapping?: boolean;
    readonly canUpdate?: boolean;
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
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:update\")"
},
v7 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:delete\")"
},
v8 = {
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
v9 = {
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
v10 = {
  "alias": "canCreateDocumentMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:create-document-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:create-document-mapping\")"
},
v11 = {
  "alias": "canDeleteDocumentMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:delete-document-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:delete-document-mapping\")"
},
v12 = {
  "alias": "canCreateAuditMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:create-audit-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:create-audit-mapping\")"
},
v13 = {
  "alias": "canDeleteAuditMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:delete-audit-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:delete-audit-mapping\")"
},
v14 = {
  "alias": "canCreateSnapshotMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:create-snapshot-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:create-snapshot-mapping\")"
},
v15 = {
  "alias": "canDeleteSnapshotMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:delete-snapshot-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:delete-snapshot-mapping\")"
},
v16 = {
  "alias": "canCreateObligationMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:create-obligation-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:create-obligation-mapping\")"
},
v17 = {
  "alias": "canDeleteObligationMapping",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:control:delete-obligation-mapping"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:control:delete-obligation-mapping\")"
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v20 = {
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
v21 = {
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
v22 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v23 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v24 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v25 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
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
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "FrameworkControlDialogFragment"
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
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
          (v18/*: any*/),
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
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "bestPractice",
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
                          (v23/*: any*/),
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
                ],
                "storageKey": "measures(first:100)"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_measures",
                "kind": "LinkedHandle",
                "name": "measures"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
                          (v24/*: any*/),
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
                                      (v25/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "versions(first:1)"
                          },
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_documents",
                "kind": "LinkedHandle",
                "name": "documents"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
                          (v24/*: any*/),
                          (v23/*: any*/),
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
                            "selections": [
                              (v2/*: any*/),
                              (v3/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
                ],
                "storageKey": "audits(first:100)"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_audits",
                "kind": "LinkedHandle",
                "name": "audits"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
                          (v25/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Profile",
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
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
                ],
                "storageKey": "obligations(first:100)"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "FrameworkGraphControl_obligations",
                "kind": "LinkedHandle",
                "name": "obligations"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
                          (v24/*: any*/),
                          (v18/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v19/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/),
                  (v21/*: any*/)
                ],
                "storageKey": "snapshots(first:100)"
              },
              {
                "alias": null,
                "args": (v22/*: any*/),
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
    "cacheID": "d177e11d82fa12c1622578238707b173",
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
    "text": "query FrameworkGraphControlNodeQuery(\n  $controlId: ID!\n) {\n  node(id: $controlId) {\n    __typename\n    ... on Control {\n      id\n      name\n      sectionTitle\n      description\n      canUpdate: permission(action: \"core:control:update\")\n      canDelete: permission(action: \"core:control:delete\")\n      canCreateMeasureMapping: permission(action: \"core:control:create-measure-mapping\")\n      canDeleteMeasureMapping: permission(action: \"core:control:delete-measure-mapping\")\n      canCreateDocumentMapping: permission(action: \"core:control:create-document-mapping\")\n      canDeleteDocumentMapping: permission(action: \"core:control:delete-document-mapping\")\n      canCreateAuditMapping: permission(action: \"core:control:create-audit-mapping\")\n      canDeleteAuditMapping: permission(action: \"core:control:delete-audit-mapping\")\n      canCreateSnapshotMapping: permission(action: \"core:control:create-snapshot-mapping\")\n      canDeleteSnapshotMapping: permission(action: \"core:control:delete-snapshot-mapping\")\n      canCreateObligationMapping: permission(action: \"core:control:create-obligation-mapping\")\n      canDeleteObligationMapping: permission(action: \"core:control:delete-obligation-mapping\")\n      ...FrameworkControlDialogFragment\n      measures(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedMeasuresCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      documents(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedDocumentsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      audits(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedAuditsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      obligations(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedObligationsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      snapshots(first: 100) {\n        edges {\n          node {\n            id\n            ...LinkedSnapshotsCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment FrameworkControlDialogFragment on Control {\n  id\n  name\n  description\n  sectionTitle\n  bestPractice\n}\n\nfragment LinkedAuditsCardFragment on Audit {\n  id\n  name\n  createdAt\n  state\n  validFrom\n  validUntil\n  framework {\n    id\n    name\n  }\n}\n\nfragment LinkedDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment LinkedMeasuresCardFragment on Measure {\n  id\n  name\n  state\n}\n\nfragment LinkedObligationsCardFragment on Obligation {\n  id\n  requirement\n  area\n  source\n  status\n  owner {\n    fullName\n    id\n  }\n}\n\nfragment LinkedSnapshotsCardFragment on Snapshot {\n  id\n  name\n  description\n  type\n  createdAt\n}\n"
  }
};
})();

(node as any).hash = "c5a03208206812d371f5133c8963a710";

export default node;
