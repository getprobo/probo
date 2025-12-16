/**
 * @generated SignedSource<<e0b1ec4a6faf9e19b0d001a9ef481ac7>>
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
v2 = [
  {
    "kind": "Variable",
    "name": "snapshotId",
    "variableName": "snapshotId"
  }
],
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
  "name": "id",
  "storageKey": null
},
v5 = {
  "kind": "Literal",
  "name": "first",
  "value": 10
},
v6 = [
  {
    "fields": (v2/*: any*/),
    "kind": "ObjectValue",
    "name": "filter"
  },
  (v5/*: any*/)
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v12 = {
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
v13 = {
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
v14 = [
  (v5/*: any*/)
],
v15 = {
  "alias": null,
  "args": null,
  "concreteType": "ProcessingActivity",
  "kind": "LinkedField",
  "name": "processingActivity",
  "plural": false,
  "selections": [
    (v4/*: any*/),
    (v8/*: any*/)
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
              {
                "args": (v2/*: any*/),
                "kind": "FragmentSpread",
                "name": "ProcessingActivitiesPageFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "ProcessingActivitiesPageDPIAFragment"
              },
              {
                "args": null,
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
          (v3/*: any*/),
          (v4/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "ProcessingActivityConnection",
                "kind": "LinkedField",
                "name": "processingActivities",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
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
                          (v4/*: any*/),
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
                          (v8/*: any*/),
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
                          (v9/*: any*/),
                          (v10/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v11/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v12/*: any*/),
                  (v13/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": [
                  "filter"
                ],
                "handle": "connection",
                "key": "ProcessingActivitiesPage_processingActivities",
                "kind": "LinkedHandle",
                "name": "processingActivities"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "concreteType": "ProcessingActivityDPIAConnection",
                "kind": "LinkedField",
                "name": "dataProtectionImpactAssessments",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ProcessingActivityDPIAEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ProcessingActivityDPIA",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v4/*: any*/),
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
                          (v15/*: any*/),
                          (v9/*: any*/),
                          (v10/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v11/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v12/*: any*/),
                  (v13/*: any*/)
                ],
                "storageKey": "dataProtectionImpactAssessments(first:10)"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "ProcessingActivitiesPage_dataProtectionImpactAssessments",
                "kind": "LinkedHandle",
                "name": "dataProtectionImpactAssessments"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "concreteType": "ProcessingActivityTIAConnection",
                "kind": "LinkedField",
                "name": "transferImpactAssessments",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ProcessingActivityTIAEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ProcessingActivityTIA",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v4/*: any*/),
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
                          (v15/*: any*/),
                          (v9/*: any*/),
                          (v10/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v11/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v12/*: any*/),
                  (v13/*: any*/)
                ],
                "storageKey": "transferImpactAssessments(first:10)"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "filters": null,
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
    "cacheID": "e518b4737308d2e7528a2f68e26d0e32",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphListQuery",
    "operationKind": "query",
    "text": "query ProcessingActivityGraphListQuery(\n  $organizationId: ID!\n  $snapshotId: ID\n) {\n  node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...ProcessingActivitiesPageFragment_3iomuz\n      ...ProcessingActivitiesPageDPIAFragment\n      ...ProcessingActivitiesPageTIAFragment\n    }\n    id\n  }\n}\n\nfragment ProcessingActivitiesPageDPIAFragment on Organization {\n  id\n  dataProtectionImpactAssessments(first: 10) {\n    totalCount\n    edges {\n      node {\n        id\n        description\n        potentialRisk\n        residualRisk\n        processingActivity {\n          id\n          name\n        }\n        createdAt\n        updatedAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n\nfragment ProcessingActivitiesPageFragment_3iomuz on Organization {\n  id\n  processingActivities(first: 10, filter: {snapshotId: $snapshotId}) {\n    totalCount\n    edges {\n      node {\n        id\n        snapshotId\n        sourceId\n        name\n        purpose\n        dataSubjectCategory\n        personalDataCategory\n        lawfulBasis\n        location\n        internationalTransfers\n        createdAt\n        updatedAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n\nfragment ProcessingActivitiesPageTIAFragment on Organization {\n  id\n  transferImpactAssessments(first: 10) {\n    totalCount\n    edges {\n      node {\n        id\n        dataSubjects\n        transfer\n        localLawRisk\n        processingActivity {\n          id\n          name\n        }\n        createdAt\n        updatedAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bbd2c1907bd9e4607a0beb1bb12d0f11";

export default node;
