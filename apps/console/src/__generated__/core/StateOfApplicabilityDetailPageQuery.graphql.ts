/**
 * @generated SignedSource<<06d20e72b67288c86573caa22b15adff>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityDetailPageQuery$variables = {
  stateOfApplicabilityId: string;
};
export type StateOfApplicabilityDetailPageQuery$data = {
  readonly node: {
    readonly canDelete?: boolean;
    readonly canExport?: boolean;
    readonly canUpdate?: boolean;
    readonly createdAt?: string;
    readonly id?: string;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
    } | null | undefined;
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly snapshotId?: string | null | undefined;
    readonly sourceId?: string | null | undefined;
    readonly updatedAt?: string;
    readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityControlsTabFragment">;
  };
};
export type StateOfApplicabilityDetailPageQuery = {
  response: StateOfApplicabilityDetailPageQuery$data;
  variables: StateOfApplicabilityDetailPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "stateOfApplicabilityId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "stateOfApplicabilityId"
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
  "name": "sourceId",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v8 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:update\")"
},
v9 = {
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
v10 = {
  "alias": "canExport",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:state-of-applicability:export"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:state-of-applicability:export\")"
},
v11 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/)
  ],
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "owner",
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
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v14 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "CONTROL_SECTION_TITLE"
    }
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "StateOfApplicabilityDetailPageQuery",
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
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "StateOfApplicabilityControlsTabFragment"
              }
            ],
            "type": "StateOfApplicability",
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
    "name": "StateOfApplicabilityDetailPageQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v13/*: any*/),
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
              {
                "alias": "canCreateApplicabilityStatement",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:applicability-statement:create"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:applicability-statement:create\")"
              },
              {
                "alias": "canUpdateApplicabilityStatement",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:applicability-statement:update"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:applicability-statement:update\")"
              },
              {
                "alias": "canDeleteApplicabilityStatement",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "core:applicability-statement:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"core:applicability-statement:delete\")"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "concreteType": "ApplicabilityStatementConnection",
                "kind": "LinkedField",
                "name": "applicabilityStatements",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ApplicabilityStatementEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ApplicabilityStatement",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
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
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Control",
                            "kind": "LinkedField",
                            "name": "control",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "sectionTitle",
                                "storageKey": null
                              },
                              (v3/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "bestPractice",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "regulatory",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "contractual",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "riskAssessment",
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
                              (v11/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v13/*: any*/)
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
                "storageKey": "applicabilityStatements(first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"CONTROL_SECTION_TITLE\"})"
              },
              {
                "alias": null,
                "args": (v14/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "StateOfApplicabilityControlsTab_applicabilityStatements",
                "kind": "LinkedHandle",
                "name": "applicabilityStatements"
              }
            ],
            "type": "StateOfApplicability",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "dceee9f818236953e28441fb1bbde7aa",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityDetailPageQuery",
    "operationKind": "query",
    "text": "query StateOfApplicabilityDetailPageQuery(\n  $stateOfApplicabilityId: ID!\n) {\n  node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      name\n      sourceId\n      snapshotId\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:state-of-applicability:update\")\n      canDelete: permission(action: \"core:state-of-applicability:delete\")\n      canExport: permission(action: \"core:state-of-applicability:export\")\n      organization {\n        id\n      }\n      owner {\n        id\n        fullName\n      }\n      ...StateOfApplicabilityControlsTabFragment\n    }\n    id\n  }\n}\n\nfragment StateOfApplicabilityControlsTabFragment on StateOfApplicability {\n  id\n  organization {\n    id\n  }\n  canCreateApplicabilityStatement: permission(action: \"core:applicability-statement:create\")\n  canUpdateApplicabilityStatement: permission(action: \"core:applicability-statement:update\")\n  canDeleteApplicabilityStatement: permission(action: \"core:applicability-statement:delete\")\n  applicabilityStatements(first: 1000, orderBy: {direction: ASC, field: CONTROL_SECTION_TITLE}) {\n    edges {\n      node {\n        id\n        applicability\n        justification\n        control {\n          id\n          sectionTitle\n          name\n          bestPractice\n          regulatory\n          contractual\n          riskAssessment\n          framework {\n            id\n            name\n          }\n          organization {\n            id\n          }\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c5fc32cde9c8e0ce3b8a77c96e0bb7cd";

export default node;
