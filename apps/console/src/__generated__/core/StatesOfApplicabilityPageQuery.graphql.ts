/**
 * @generated SignedSource<<50cd52d641383a190d37f3e430cbe3f7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StatesOfApplicabilityPageQuery$variables = {
  organizationId: string;
};
export type StatesOfApplicabilityPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly canCreateStateOfApplicability: boolean;
    readonly id: string;
    readonly " $fragmentSpreads": FragmentRefs<"StatesOfApplicabilityPageFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type StatesOfApplicabilityPageQuery = {
  response: StatesOfApplicabilityPageQuery$data;
  variables: StatesOfApplicabilityPageQuery$variables;
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
  "alias": "canCreateStateOfApplicability",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:create\")"
},
v5 = [
  {
    "kind": "Literal",
    "name": "filter",
    "value": {
      "snapshotId": null
    }
  },
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
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
    "name": "StatesOfApplicabilityPageQuery",
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
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "StatesOfApplicabilityPageFragment"
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
    "name": "StatesOfApplicabilityPageQuery",
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
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v4/*: any*/),
              {
                "alias": null,
                "args": (v5/*: any*/),
                "concreteType": "StateOfApplicabilityConnection",
                "kind": "LinkedField",
                "name": "statesOfApplicability",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "StateOfApplicabilityEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "StateOfApplicability",
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
                            "name": "createdAt",
                            "storageKey": null
                          },
                          {
                            "alias": "applicabilityStatementsInfo",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "first",
                                "value": 0
                              }
                            ],
                            "concreteType": "ApplicabilityStatementConnection",
                            "kind": "LinkedField",
                            "name": "applicabilityStatements",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "totalCount",
                                "storageKey": null
                              }
                            ],
                            "storageKey": "applicabilityStatements(first:0)"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:state-of-applicability:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:state-of-applicability:delete\")"
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
                "storageKey": "statesOfApplicability(filter:{\"snapshotId\":null},first:50,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v5/*: any*/),
                "filters": [
                  "orderBy",
                  "filter"
                ],
                "handle": "connection",
                "key": "StatesOfApplicabilityPage_statesOfApplicability",
                "kind": "LinkedHandle",
                "name": "statesOfApplicability"
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
    "cacheID": "ba258706565308bd0be8b9f0fb192fb3",
    "id": null,
    "metadata": {},
    "name": "StatesOfApplicabilityPageQuery",
    "operationKind": "query",
    "text": "query StatesOfApplicabilityPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      canCreateStateOfApplicability: permission(action: \"core:state-of-applicability:create\")\n      ...StatesOfApplicabilityPageFragment\n    }\n    id\n  }\n}\n\nfragment StatesOfApplicabilityPageFragment on Organization {\n  statesOfApplicability(first: 50, orderBy: {direction: DESC, field: CREATED_AT}, filter: {snapshotId: null}) {\n    edges {\n      node {\n        id\n        ...StatesOfApplicabilityPageRowFragment\n        canDelete: permission(action: \"core:state-of-applicability:delete\")\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment StatesOfApplicabilityPageRowFragment on StateOfApplicability {\n  id\n  name\n  createdAt\n  applicabilityStatementsInfo: applicabilityStatements(first: 0) {\n    totalCount\n  }\n}\n"
  }
};
})();

(node as any).hash = "10fd3ffd276b002b98c3b9aa437ccbed";

export default node;
