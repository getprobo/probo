/**
 * @generated SignedSource<<73338f546d62947cb3d12159cb651e0a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageReferencesPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageReferencesPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly compliancePage: {
      readonly canCreateComplianceBadge: boolean;
      readonly canCreateReference: boolean;
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageBadgeListFragment" | "CompliancePageReferenceListFragment">;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type CompliancePageReferencesPageQuery = {
  response: CompliancePageReferencesPageQuery$data;
  variables: CompliancePageReferencesPageQuery$variables;
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": "canCreateReference",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center-reference:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center-reference:create\")"
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "canCreateComplianceBadge",
  "storageKey": null
},
v6 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "RANK"
    }
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "rank",
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
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageReferencesPageQuery",
    "selections": [
      {
        "alias": "organization",
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
              {
                "kind": "RequiredField",
                "field": {
                  "alias": "compliancePage",
                  "args": null,
                  "concreteType": "TrustCenter",
                  "kind": "LinkedField",
                  "name": "trustCenter",
                  "plural": false,
                  "selections": [
                    (v3/*: any*/),
                    (v4/*: any*/),
                    (v5/*: any*/),
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageReferenceListFragment"
                    },
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageBadgeListFragment"
                    }
                  ],
                  "storageKey": null
                },
                "action": "THROW"
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
    "name": "CompliancePageReferencesPageQuery",
    "selections": [
      {
        "alias": "organization",
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
              {
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "TrustCenterReferenceConnection",
                    "kind": "LinkedField",
                    "name": "references",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterReferenceEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "TrustCenterReference",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v7/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "logoUrl",
                                "storageKey": null
                              },
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
                                "name": "websiteUrl",
                                "storageKey": null
                              },
                              {
                                "alias": "canUpdate",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:trust-center-reference:update"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:trust-center-reference:update\")"
                              },
                              {
                                "alias": "canDelete",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:trust-center-reference:delete"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:trust-center-reference:delete\")"
                              },
                              (v2/*: any*/)
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
                    "storageKey": "references(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "filters": (v12/*: any*/),
                    "handle": "connection",
                    "key": "CompliancePageReferenceList_references",
                    "kind": "LinkedHandle",
                    "name": "references"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "ComplianceBadgeConnection",
                    "kind": "LinkedField",
                    "name": "complianceBadges",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceBadgeEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "ComplianceBadge",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v7/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "iconUrl",
                                "storageKey": null
                              },
                              (v8/*: any*/),
                              {
                                "alias": "canUpdate",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:compliance-badge:update"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:compliance-badge:update\")"
                              },
                              {
                                "alias": "canDelete",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:compliance-badge:delete"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:compliance-badge:delete\")"
                              },
                              (v2/*: any*/)
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
                    "storageKey": "complianceBadges(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "filters": (v12/*: any*/),
                    "handle": "connection",
                    "key": "CompliancePageBadgeList_complianceBadges",
                    "kind": "LinkedHandle",
                    "name": "complianceBadges"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "525264edede7cea2f9c951342a28c9d0",
    "id": null,
    "metadata": {},
    "name": "CompliancePageReferencesPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageReferencesPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        id\n        canCreateReference: permission(action: \"core:trust-center-reference:create\")\n        canCreateComplianceBadge\n        ...CompliancePageReferenceListFragment\n        ...CompliancePageBadgeListFragment\n      }\n    }\n    id\n  }\n}\n\nfragment CompliancePageBadgeListFragment on TrustCenter {\n  complianceBadges(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        rank\n        ...CompliancePageBadgeListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment CompliancePageBadgeListItemFragment on ComplianceBadge {\n  id\n  iconUrl\n  name\n  rank\n  canUpdate: permission(action: \"core:compliance-badge:update\")\n  canDelete: permission(action: \"core:compliance-badge:delete\")\n}\n\nfragment CompliancePageReferenceListFragment on TrustCenter {\n  references(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        rank\n        ...CompliancePageReferenceListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment CompliancePageReferenceListItemFragment on TrustCenterReference {\n  id\n  logoUrl\n  name\n  description\n  rank\n  websiteUrl\n  canUpdate: permission(action: \"core:trust-center-reference:update\")\n  canDelete: permission(action: \"core:trust-center-reference:delete\")\n}\n"
  }
};
})();

(node as any).hash = "1e87595675566c24388d86008fe92668";

export default node;
