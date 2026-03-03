/**
 * @generated SignedSource<<cf316aa4cda5fc4c9aebae8f0f10c365>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageOverviewPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageOverviewPageQuery$data = {
  readonly organization: {
    readonly compliancePage?: {
      readonly canGetNDA: boolean;
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageFrameworkList_compliancePageFragment">;
    } | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageNDASectionFragment" | "CompliancePageSlackSectionFragment" | "CompliancePageStatusSectionFragment">;
  };
};
export type CompliancePageOverviewPageQuery = {
  response: CompliancePageOverviewPageQuery$data;
  variables: CompliancePageOverviewPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
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
  "alias": "canGetNDA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center:get-nda"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center:get-nda\")"
},
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
  "value": 100
},
v6 = [
  (v5/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "RANK"
    }
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageOverviewPageQuery",
    "selections": [
      {
        "alias": "organization",
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
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "CompliancePageFrameworkList_compliancePageFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageStatusSectionFragment"
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageNDASectionFragment"
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageSlackSectionFragment"
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
    "name": "CompliancePageOverviewPageQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
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
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v4/*: any*/),
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:update\")"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "ComplianceFrameworkConnection",
                    "kind": "LinkedField",
                    "name": "complianceFrameworks",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceFrameworkEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "ComplianceFramework",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v4/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "rank",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "visibility",
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
                                  (v4/*: any*/),
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
                                    "name": "lightLogoURL",
                                    "storageKey": null
                                  },
                                  {
                                    "alias": null,
                                    "args": null,
                                    "kind": "ScalarField",
                                    "name": "darkLogoURL",
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              },
                              (v3/*: any*/)
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
                      },
                      {
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
                      }
                    ],
                    "storageKey": "complianceFrameworks(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "filters": [
                      "orderBy"
                    ],
                    "handle": "connection",
                    "key": "CompliancePageFrameworkList_complianceFrameworks",
                    "kind": "LinkedHandle",
                    "name": "complianceFrameworks"
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "active",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "ndaFileName",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "ndaFileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": "canUploadNDA",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:upload-nda"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:upload-nda\")"
                  },
                  {
                    "alias": "canDeleteNDA",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:delete-nda"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:delete-nda\")"
                  }
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "CustomDomain",
                "kind": "LinkedField",
                "name": "customDomain",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "domain",
                    "storageKey": null
                  },
                  (v4/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": [
                  (v5/*: any*/)
                ],
                "concreteType": "SlackConnectionConnection",
                "kind": "LinkedField",
                "name": "slackConnections",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SlackConnectionEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "SlackConnection",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v4/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "channel",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "createdAt",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "slackConnections(first:100)"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "e6cc68fe4df4a8dffc558f6db82ead9c",
    "id": null,
    "metadata": {},
    "name": "CompliancePageOverviewPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageOverviewPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        canGetNDA: permission(action: \"core:trust-center:get-nda\")\n        ...CompliancePageFrameworkList_compliancePageFragment\n        id\n      }\n    }\n    ...CompliancePageStatusSectionFragment\n    ...CompliancePageNDASectionFragment\n    ...CompliancePageSlackSectionFragment\n    id\n  }\n}\n\nfragment CompliancePageFrameworkList_compliancePageFragment on TrustCenter {\n  id\n  canUpdate: permission(action: \"core:trust-center:update\")\n  complianceFrameworks(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        rank\n        visibility\n        framework {\n          id\n          name\n          lightLogoURL\n          darkLogoURL\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment CompliancePageNDASectionFragment on Organization {\n  compliancePage: trustCenter {\n    id\n    ndaFileName\n    ndaFileUrl\n    canUpdate: permission(action: \"core:trust-center:update\")\n    canUploadNDA: permission(action: \"core:trust-center:upload-nda\")\n    canDeleteNDA: permission(action: \"core:trust-center:delete-nda\")\n  }\n}\n\nfragment CompliancePageSlackSectionFragment on Organization {\n  compliancePage: trustCenter {\n    canUpdate: permission(action: \"core:trust-center:update\")\n    id\n  }\n  slackConnections(first: 100) {\n    edges {\n      node {\n        id\n        channel\n        createdAt\n      }\n    }\n  }\n}\n\nfragment CompliancePageStatusSectionFragment on Organization {\n  customDomain {\n    domain\n    id\n  }\n  compliancePage: trustCenter {\n    id\n    active\n    canUpdate: permission(action: \"core:trust-center:update\")\n  }\n}\n"
  }
};
})();

(node as any).hash = "be0dff83941c21755ea2ff77a865805b";

export default node;
