/**
 * @generated SignedSource<<06a60a60cf7ca339c25395a94906774a>>
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
    readonly canInviteUser: boolean;
    readonly invitations: {
      readonly $updatableFragmentSpreads: FragmentRefs<"PeoplePage_invitationsTotalCountFragment">;
      readonly totalCount: number | null | undefined;
    };
    readonly members: {
      readonly totalCount: number | null | undefined;
    };
    readonly " $fragmentSpreads": FragmentRefs<"InvitationListFragment" | "PeopleListFragment">;
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
  "alias": "canInviteUser",
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
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v5 = {
  "direction": "ASC",
  "field": "FULL_NAME"
},
v6 = [
  (v4/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v5/*: any*/)
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v8 = {
  "direction": "DESC",
  "field": "CREATED_AT"
},
v9 = [
  (v4/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v8/*: any*/)
  }
],
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "role",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v16 = {
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
v17 = {
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
v18 = [
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
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "MembershipConnection",
                    "kind": "LinkedField",
                    "name": "members",
                    "plural": false,
                    "selections": [
                      (v7/*: any*/)
                    ],
                    "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
                  },
                  "action": "THROW"
                },
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
                },
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v9/*: any*/),
                    "concreteType": "InvitationConnection",
                    "kind": "LinkedField",
                    "name": "invitations",
                    "plural": false,
                    "selections": [
                      (v7/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "PeoplePage_invitationsTotalCountFragment"
                      }
                    ],
                    "storageKey": "invitations(first:20,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  },
                  "action": "THROW"
                },
                {
                  "args": [
                    (v4/*: any*/),
                    {
                      "kind": "Literal",
                      "name": "order",
                      "value": (v8/*: any*/)
                    }
                  ],
                  "kind": "FragmentSpread",
                  "name": "InvitationListFragment"
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
          (v10/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "MembershipConnection",
                "kind": "LinkedField",
                "name": "members",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "MembershipEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Membership",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v10/*: any*/),
                          (v11/*: any*/),
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
                            "concreteType": "MembershipProfile",
                            "kind": "LinkedField",
                            "name": "profile",
                            "plural": false,
                            "selections": [
                              (v10/*: any*/),
                              (v12/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "kind",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "position",
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Identity",
                            "kind": "LinkedField",
                            "name": "identity",
                            "plural": false,
                            "selections": [
                              (v13/*: any*/),
                              (v10/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v14/*: any*/),
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
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "iam:membership:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"iam:membership:delete\")"
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v15/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v16/*: any*/),
                  (v17/*: any*/)
                ],
                "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "PeopleListFragment_members",
                "kind": "LinkedHandle",
                "name": "members"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "invitations",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
                  (v2/*: any*/),
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
                          (v10/*: any*/),
                          (v12/*: any*/),
                          (v13/*: any*/),
                          (v11/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "status",
                            "storageKey": null
                          },
                          (v14/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "acceptedAt",
                            "storageKey": null
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "iam:invitation:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"iam:invitation:delete\")"
                          },
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v15/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v16/*: any*/),
                  (v17/*: any*/)
                ],
                "storageKey": "invitations(first:20,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "InvitationListFragment_invitations",
                "kind": "LinkedHandle",
                "name": "invitations"
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
    "cacheID": "425e73a84c96979c3b380c17512c744a",
    "id": null,
    "metadata": {},
    "name": "PeoplePageQuery",
    "operationKind": "query",
    "text": "query PeoplePageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canInviteUser: permission(action: \"iam:invitation:create\")\n      members(first: 20, orderBy: {direction: ASC, field: FULL_NAME}) {\n        totalCount\n      }\n      ...PeopleListFragment_8lnpd\n      invitations(first: 20, orderBy: {direction: DESC, field: CREATED_AT}) {\n        totalCount\n        __typename\n      }\n      ...InvitationListFragment_1PypFi\n    }\n    id\n  }\n}\n\nfragment InvitationListFragment_1PypFi on Organization {\n  invitations(first: 20, orderBy: {direction: DESC, field: CREATED_AT}) {\n    edges {\n      node {\n        id\n        ...InvitationListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment InvitationListItemFragment on Invitation {\n  id\n  fullName\n  email\n  role\n  status\n  createdAt\n  acceptedAt\n  canDelete: permission(action: \"iam:invitation:delete\")\n}\n\nfragment PeopleListFragment_8lnpd on Organization {\n  members(first: 20, orderBy: {direction: ASC, field: FULL_NAME}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...PeopleListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment PeopleListItemFragment on Membership {\n  id\n  role\n  source\n  state\n  profile {\n    id\n    fullName\n    kind\n    position\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership:update\")\n  canDelete: permission(action: \"iam:membership:delete\")\n}\n"
  }
};
})();

(node as any).hash = "9ab9d1a720b23dd78bdb82530a30b6f3";

export default node;
