/**
 * @generated SignedSource<<0a0f91bc4b1622f7dea5c39185904169>>
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
      readonly canCreateReference: boolean;
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageReferenceListFragment">;
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
v5 = [
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
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageReferenceListFragment"
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
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
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
                                "name": "logoUrl",
                                "storageKey": null
                              },
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
                    "storageKey": "references(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "filters": [
                      "orderBy"
                    ],
                    "handle": "connection",
                    "key": "CompliancePageReferenceList_references",
                    "kind": "LinkedHandle",
                    "name": "references"
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
    "cacheID": "d352c6f8d833eb8c105e72184c070a85",
    "id": null,
    "metadata": {},
    "name": "CompliancePageReferencesPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageReferencesPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        id\n        canCreateReference: permission(action: \"core:trust-center-reference:create\")\n        ...CompliancePageReferenceListFragment\n      }\n    }\n    id\n  }\n}\n\nfragment CompliancePageReferenceListFragment on TrustCenter {\n  references(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        rank\n        ...CompliancePageReferenceListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  id\n}\n\nfragment CompliancePageReferenceListItemFragment on TrustCenterReference {\n  id\n  logoUrl\n  name\n  description\n  rank\n  websiteUrl\n  canUpdate: permission(action: \"core:trust-center-reference:update\")\n  canDelete: permission(action: \"core:trust-center-reference:delete\")\n}\n"
  }
};
})();

(node as any).hash = "9c5e008e8f12d028e3b280a833ec3fc5";

export default node;
