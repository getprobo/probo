/**
 * @generated SignedSource<<a32e2b4330ac1d54762817127289b51d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrderDirection = "ASC" | "DESC";
export type ProfileOrderField = "CREATED_AT" | "FULL_NAME" | "KIND" | "ORGANIZATION_NAME";
export type ProfileOrder = {
  direction: OrderDirection;
  field: ProfileOrderField;
};
export type PeopleListFragment_RefetchQuery$variables = {
  after?: string | null | undefined;
  before?: string | null | undefined;
  first?: number | null | undefined;
  id: string;
  last?: number | null | undefined;
  order?: ProfileOrder | null | undefined;
};
export type PeopleListFragment_RefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"PeopleListFragment">;
  } | null | undefined;
};
export type PeopleListFragment_RefetchQuery = {
  response: PeopleListFragment_RefetchQuery$data;
  variables: PeopleListFragment_RefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "after"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "before"
},
v2 = {
  "defaultValue": 20,
  "kind": "LocalArgument",
  "name": "first"
},
v3 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v4 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "last"
},
v5 = {
  "defaultValue": {
    "direction": "ASC",
    "field": "FULL_NAME"
  },
  "kind": "LocalArgument",
  "name": "order"
},
v6 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
],
v7 = {
  "kind": "Variable",
  "name": "after",
  "variableName": "after"
},
v8 = {
  "kind": "Variable",
  "name": "before",
  "variableName": "before"
},
v9 = {
  "kind": "Variable",
  "name": "first",
  "variableName": "first"
},
v10 = {
  "kind": "Variable",
  "name": "last",
  "variableName": "last"
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v13 = [
  (v7/*: any*/),
  (v8/*: any*/),
  (v9/*: any*/),
  (v10/*: any*/),
  {
    "kind": "Variable",
    "name": "orderBy",
    "variableName": "order"
  }
],
v14 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
],
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v19 = {
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
v20 = [
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/),
      (v3/*: any*/),
      (v4/*: any*/),
      (v5/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "PeopleListFragment_RefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v6/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": [
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              {
                "kind": "Variable",
                "name": "order",
                "variableName": "order"
              }
            ],
            "kind": "FragmentSpread",
            "name": "PeopleListFragment"
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
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/),
      (v2/*: any*/),
      (v4/*: any*/),
      (v5/*: any*/),
      (v3/*: any*/)
    ],
    "kind": "Operation",
    "name": "PeopleListFragment_RefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v6/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v11/*: any*/),
          (v12/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v13/*: any*/),
                "concreteType": "ProfileConnection",
                "kind": "LinkedField",
                "name": "profiles",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "totalCount",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ProfileEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Profile",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v12/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "source",
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
                            "name": "fullName",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "emailAddress",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Membership",
                            "kind": "LinkedField",
                            "name": "membership",
                            "plural": false,
                            "selections": [
                              (v12/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "role",
                                "storageKey": null
                              },
                              {
                                "alias": "canUpdate",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "iam:membership:update"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"iam:membership:update\")"
                              }
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": "lastInvitation",
                            "args": (v14/*: any*/),
                            "concreteType": "InvitationConnection",
                            "kind": "LinkedField",
                            "name": "pendingInvitations",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "InvitationEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "Invitation",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v12/*: any*/),
                                      (v15/*: any*/),
                                      (v11/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v16/*: any*/)
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
                                  (v17/*: any*/),
                                  (v18/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v19/*: any*/)
                            ],
                            "storageKey": "pendingInvitations(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                          },
                          {
                            "alias": "lastInvitation",
                            "args": (v14/*: any*/),
                            "filters": (v20/*: any*/),
                            "handle": "connection",
                            "key": "PeopleListItem_lastInvitation",
                            "kind": "LinkedHandle",
                            "name": "pendingInvitations"
                          },
                          (v15/*: any*/),
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "iam:membership-profile:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"iam:membership-profile:update\")"
                          },
                          {
                            "alias": "canInvite",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "iam:invitation:create"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"iam:invitation:create\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "iam:membership-profile:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"iam:membership-profile:delete\")"
                          },
                          (v11/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
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
                      (v17/*: any*/),
                      (v18/*: any*/),
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
                  (v19/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v13/*: any*/),
                "filters": (v20/*: any*/),
                "handle": "connection",
                "key": "PeopleListFragment_profiles",
                "kind": "LinkedHandle",
                "name": "profiles"
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
    "cacheID": "15a5dd79df1b16aa1a66ccb8024f9627",
    "id": null,
    "metadata": {},
    "name": "PeopleListFragment_RefetchQuery",
    "operationKind": "query",
    "text": "query PeopleListFragment_RefetchQuery(\n  $after: CursorKey = null\n  $before: CursorKey = null\n  $first: Int = 20\n  $last: Int = null\n  $order: ProfileOrder = {direction: ASC, field: FULL_NAME}\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...PeopleListFragment_16fISc\n    id\n  }\n}\n\nfragment PeopleListFragment_16fISc on Organization {\n  profiles(first: $first, after: $after, last: $last, before: $before, orderBy: $order) {\n    totalCount\n    edges {\n      node {\n        id\n        ...PeopleListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment PeopleListItemFragment on Profile {\n  id\n  source\n  state\n  fullName\n  emailAddress\n  membership {\n    id\n    role\n    canUpdate: permission(action: \"iam:membership:update\")\n  }\n  lastInvitation: pendingInvitations(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        createdAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n  canInvite: permission(action: \"iam:invitation:create\")\n  canDelete: permission(action: \"iam:membership-profile:delete\")\n}\n"
  }
};
})();

(node as any).hash = "ee3aeca404063a0660eeb9ff27592e86";

export default node;
