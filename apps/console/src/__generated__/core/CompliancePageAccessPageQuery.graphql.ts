/**
 * @generated SignedSource<<780d27bc715545a78041c8202835912d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAccessPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageAccessPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly compliancePage: {
      readonly canCreateAccess: boolean;
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAccessListFragment">;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type CompliancePageAccessPageQuery = {
  response: CompliancePageAccessPageQuery$data;
  variables: CompliancePageAccessPageQuery$variables;
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
  "alias": "canCreateAccess",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center-access:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center-access:create\")"
},
v5 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 10
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageAccessPageQuery",
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
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageAccessListFragment"
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
    "name": "CompliancePageAccessPageQuery",
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
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "concreteType": "TrustCenterAccessConnection",
                    "kind": "LinkedField",
                    "name": "accesses",
                    "plural": false,
                    "selections": [
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
                            "name": "hasNextPage",
                            "storageKey": null
                          },
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
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterAccessEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "cursor",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "TrustCenterAccess",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
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
                                "name": "email",
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
                                "name": "state",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "activeCount",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "pendingRequestCount",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "hasAcceptedNonDisclosureAgreement",
                                "storageKey": null
                              },
                              {
                                "alias": "canUpdate",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:trust-center-access:update"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:trust-center-access:update\")"
                              },
                              (v2/*: any*/)
                            ],
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
                    "storageKey": "accesses(first:10,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  },
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "filters": [
                      "orderBy"
                    ],
                    "handle": "connection",
                    "key": "CompliancePageAccessList_accesses",
                    "kind": "LinkedHandle",
                    "name": "accesses"
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
    "cacheID": "e8e8a8bf8523698493434cd320860613",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAccessPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageAccessPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        id\n        canCreateAccess: permission(action: \"core:trust-center-access:create\")\n        ...CompliancePageAccessListFragment\n      }\n    }\n    id\n  }\n}\n\nfragment CompliancePageAccessListFragment on TrustCenter {\n  accesses(first: 10, orderBy: {field: CREATED_AT, direction: DESC}) {\n    pageInfo {\n      hasNextPage\n      hasPreviousPage\n      startCursor\n      endCursor\n    }\n    edges {\n      cursor\n      node {\n        id\n        ...CompliancePageAccessListItemFragment\n        __typename\n      }\n    }\n  }\n  id\n}\n\nfragment CompliancePageAccessListItemFragment on TrustCenterAccess {\n  id\n  name\n  email\n  createdAt\n  state\n  activeCount\n  pendingRequestCount\n  hasAcceptedNonDisclosureAgreement\n  canUpdate: permission(action: \"core:trust-center-access:update\")\n}\n"
  }
};
})();

(node as any).hash = "4cc473e91fd49ddd83b6cdb0c8da0abf";

export default node;
