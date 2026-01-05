/**
 * @generated SignedSource<<5c57612f8793f7b056bf0046febe42ca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ProcessingActivityGraphListQuery$variables = {
  organizationId: string;
  snapshotId?: string | null | undefined;
};
export type ProcessingActivityGraphListQuery$data = {
  readonly node: {
    readonly canCreateProcessingActivity?: boolean;
    readonly canExportDataProtectionImpactAssessments?: boolean;
    readonly canExportProcessingActivities?: boolean;
    readonly canExportTransferImpactAssessments?: boolean;
    readonly " $fragmentSpreads": FragmentRefs<"ProcessingActivitiesPageDPIAFragment" | "ProcessingActivitiesPageFragment" | "ProcessingActivitiesPageTIAFragment">;
  };
};
export type ProcessingActivityGraphListQuery = {
  response: ProcessingActivityGraphListQuery$data;
  variables: ProcessingActivityGraphListQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  },
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "snapshotId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v2 = {
  "alias": "canCreateProcessingActivity",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:processing-activity:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:processing-activity:create\")"
},
v3 = {
  "alias": "canExportProcessingActivities",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:processing-activity:export"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:processing-activity:export\")"
},
v4 = {
  "alias": "canExportDataProtectionImpactAssessments",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:data-protection-impact-assessment:export"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:data-protection-impact-assessment:export\")"
},
v5 = {
  "alias": "canExportTransferImpactAssessments",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:transfer-impact-assessment:export"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:transfer-impact-assessment:export\")"
},
v6 = [
  {
    "kind": "Variable",
    "name": "snapshotId",
    "variableName": "snapshotId"
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v9 = [
  {
    "fields": (v6/*: any*/),
    "kind": "ObjectValue",
    "name": "filter"
  },
  {
    "kind": "Literal",
    "name": "first",
    "value": 10
  }
],
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
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
  "name": "cursor",
  "storageKey": null
},
v15 = {
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
      "name": "hasNextPage",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "endCursor",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v16 = {
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
v17 = [
  "filter"
],
v18 = {
  "alias": null,
  "args": null,
  "concreteType": "ProcessingActivity",
  "kind": "LinkedField",
  "name": "processingActivity",
  "plural": false,
  "selections": [
    (v8/*: any*/),
    (v11/*: any*/)
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ProcessingActivityGraphListQuery",
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
              {
                "args": (v6/*: any*/),
                "kind": "FragmentSpread",
                "name": "ProcessingActivitiesPageFragment"
              },
              {
                "args": (v6/*: any*/),
                "kind": "FragmentSpread",
                "name": "ProcessingActivitiesPageDPIAFragment"
              },
              {
                "args": (v6/*: any*/),
                "kind": "FragmentSpread",
                "name": "ProcessingActivitiesPageTIAFragment"
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ProcessingActivityGraphListQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v7/*: any*/),
          (v8/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "ProcessingActivityConnection",
                "kind": "LinkedField",
                "name": "processingActivities",
                "plural": false,
                "selections": [
                  (v10/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ProcessingActivityEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ProcessingActivity",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v8/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "snapshotId",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "sourceId",
                            "storageKey": null
                          },
                          (v11/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "purpose",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "dataSubjectCategory",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "personalDataCategory",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "lawfulBasis",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "location",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "internationalTransfers",
                            "storageKey": null
                          },
                          (v12/*: any*/),
                          (v13/*: any*/),
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:processing-activity:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:processing-activity:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:processing-activity:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:processing-activity:delete\")"
                          },
                          (v7/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v15/*: any*/),
                  (v16/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v17/*: any*/),
                "handle": "connection",
                "key": "ProcessingActivitiesPage_processingActivities",
                "kind": "LinkedHandle",
                "name": "processingActivities"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "DataProtectionImpactAssessmentConnection",
                "kind": "LinkedField",
                "name": "dataProtectionImpactAssessments",
                "plural": false,
                "selections": [
                  (v10/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DataProtectionImpactAssessmentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DataProtectionImpactAssessment",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v8/*: any*/),
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
                            "name": "potentialRisk",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "residualRisk",
                            "storageKey": null
                          },
                          (v18/*: any*/),
                          (v12/*: any*/),
                          (v13/*: any*/),
                          (v7/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v15/*: any*/),
                  (v16/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v17/*: any*/),
                "handle": "connection",
                "key": "ProcessingActivitiesPage_dataProtectionImpactAssessments",
                "kind": "LinkedHandle",
                "name": "dataProtectionImpactAssessments"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "TransferImpactAssessmentConnection",
                "kind": "LinkedField",
                "name": "transferImpactAssessments",
                "plural": false,
                "selections": [
                  (v10/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TransferImpactAssessmentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TransferImpactAssessment",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v8/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "dataSubjects",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "transfer",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "localLawRisk",
                            "storageKey": null
                          },
                          (v18/*: any*/),
                          (v12/*: any*/),
                          (v13/*: any*/),
                          (v7/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v15/*: any*/),
                  (v16/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v17/*: any*/),
                "handle": "connection",
                "key": "ProcessingActivitiesPage_transferImpactAssessments",
                "kind": "LinkedHandle",
                "name": "transferImpactAssessments"
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
    "cacheID": "84a26e65e3de62f76f034786cf174f80",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphListQuery",
    "operationKind": "query",
    "text": "query ProcessingActivityGraphListQuery(\n  $organizationId: ID!\n  $snapshotId: ID\n) {\n  node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canCreateProcessingActivity: permission(action: \"core:processing-activity:create\")\n      canExportProcessingActivities: permission(action: \"core:processing-activity:export\")\n      canExportDataProtectionImpactAssessments: permission(action: \"core:data-protection-impact-assessment:export\")\n      canExportTransferImpactAssessments: permission(action: \"core:transfer-impact-assessment:export\")\n      ...ProcessingActivitiesPageFragment_3iomuz\n      ...ProcessingActivitiesPageDPIAFragment_3iomuz\n      ...ProcessingActivitiesPageTIAFragment_3iomuz\n    }\n    id\n  }\n}\n\nfragment ProcessingActivitiesPageDPIAFragment_3iomuz on Organization {\n  id\n  dataProtectionImpactAssessments(first: 10, filter: {snapshotId: $snapshotId}) {\n    totalCount\n    edges {\n      node {\n        id\n        description\n        potentialRisk\n        residualRisk\n        processingActivity {\n          id\n          name\n        }\n        createdAt\n        updatedAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n\nfragment ProcessingActivitiesPageFragment_3iomuz on Organization {\n  id\n  processingActivities(first: 10, filter: {snapshotId: $snapshotId}) {\n    totalCount\n    edges {\n      node {\n        id\n        snapshotId\n        sourceId\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        lawfulBasis\n        location\n        internationalTransfers\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:processing-activity:update\")\n        canDelete: permission(action: \"core:processing-activity:delete\")\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n\nfragment ProcessingActivitiesPageTIAFragment_3iomuz on Organization {\n  id\n  transferImpactAssessments(first: 10, filter: {snapshotId: $snapshotId}) {\n    totalCount\n    edges {\n      node {\n        id\n        dataSubjects\n        transfer\n        localLawRisk\n        processingActivity {\n          id\n          name\n        }\n        createdAt\n        updatedAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6efb688e3142c0eceb054ded9df044de";

export default node;
