/**
 * @generated SignedSource<<86b86e1209fbe58b62c2be028fb0482c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PeoplePageQuery$variables = {
  organizationId: string;
};
export type PeoplePageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly canCreateUser: boolean;
    readonly " $fragmentSpreads": FragmentRefs<"PeopleListFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type PeoplePageQuery = {
  response: PeoplePageQuery$data;
  variables: PeoplePageQuery$variables;
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
  "alias": "canCreateUser",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "iam:membership-profile:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"iam:membership-profile:create\")"
},
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v5 = {
  "direction": "ASC",
  "field": "FULL_NAME"
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v7 = [
  (v4/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v5/*: any*/)
  }
],
v8 = [
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
  "name": "cursor",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
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
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PeoplePageQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
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
                {
                  "args": [
                    (v4/*: any*/),
                    {
                      "kind": "Literal",
                      "name": "order",
                      "value": (v5/*: any*/)
                    }
                  ],
                  "kind": "FragmentSpread",
                  "name": "PeopleListFragment"
                }
              ],
              "type": "Organization",
              "abstractKey": null
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "PeoplePageQuery",
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
          (v6/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": (v7/*: any*/),
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
                          (v6/*: any*/),
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
                              (v6/*: any*/),
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
                            "args": (v8/*: any*/),
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
                                      (v6/*: any*/),
                                      (v9/*: any*/),
                                      (v2/*: any*/)
                                    ],
                                    "storageKey": null
                                  },
                                  (v10/*: any*/)
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
                                  (v11/*: any*/),
                                  (v12/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v13/*: any*/)
                            ],
                            "storageKey": "pendingInvitations(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                          },
                          {
                            "alias": "lastInvitation",
                            "args": (v8/*: any*/),
                            "filters": (v14/*: any*/),
                            "handle": "connection",
                            "key": "PeopleListItem_lastInvitation",
                            "kind": "LinkedHandle",
                            "name": "pendingInvitations"
                          },
                          (v9/*: any*/),
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
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v10/*: any*/)
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
                      (v11/*: any*/),
                      (v12/*: any*/),
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
                  (v13/*: any*/)
                ],
                "storageKey": "profiles(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
                "filters": (v14/*: any*/),
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
    "cacheID": "af366ea24d2c96763b104acc69d9b3ff",
    "id": null,
    "metadata": {},
    "name": "PeoplePageQuery",
    "operationKind": "query",
    "text": "query PeoplePageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canCreateUser: permission(action: \"iam:membership-profile:create\")\n      ...PeopleListFragment_8lnpd\n    }\n    id\n  }\n}\n\nfragment PeopleListFragment_8lnpd on Organization {\n  profiles(first: 20, orderBy: {direction: ASC, field: FULL_NAME}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...PeopleListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment PeopleListItemFragment on Profile {\n  id\n  source\n  state\n  fullName\n  emailAddress\n  membership {\n    id\n    role\n    canUpdate: permission(action: \"iam:membership:update\")\n  }\n  lastInvitation: pendingInvitations(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        createdAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n  canInvite: permission(action: \"iam:invitation:create\")\n  canDelete: permission(action: \"iam:membership-profile:delete\")\n}\n"
  }
};
})();

(node as any).hash = "096ed215c8310f05d842827b5f3aec68";

export default node;
