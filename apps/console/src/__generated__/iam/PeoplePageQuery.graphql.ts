/**
 * @generated SignedSource<<a9d41020a1c8eb8c673aa59fa2772468>>
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
    readonly canInviteUser: boolean;
    readonly invitations: {
      readonly $updatableFragmentSpreads: FragmentRefs<"PeoplePage_invitationsTotalCountFragment">;
      readonly totalCount: number | null | undefined;
    };
    readonly profiles: {
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
v5 = {
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v6 = {
  "direction": "ASC",
  "field": "FULL_NAME"
},
v7 = [
  (v5/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v6/*: any*/)
  }
],
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v9 = {
  "direction": "DESC",
  "field": "CREATED_AT"
},
v10 = [
  (v5/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v9/*: any*/)
  }
],
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
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
  "name": "role",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
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
v18 = {
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
v19 = [
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
                (v4/*: any*/),
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "concreteType": "ProfileConnection",
                    "kind": "LinkedField",
                    "name": "profiles",
                    "plural": false,
                    "selections": [
                      (v8/*: any*/)
                    ],
                    "storageKey": "profiles(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
                  },
                  "action": "THROW"
                },
                {
                  "args": [
                    (v5/*: any*/),
                    {
                      "kind": "Literal",
                      "name": "order",
                      "value": (v6/*: any*/)
                    }
                  ],
                  "kind": "FragmentSpread",
                  "name": "PeopleListFragment"
                },
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v10/*: any*/),
                    "concreteType": "InvitationConnection",
                    "kind": "LinkedField",
                    "name": "invitations",
                    "plural": false,
                    "selections": [
                      (v8/*: any*/),
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
                    (v5/*: any*/),
                    {
                      "kind": "Literal",
                      "name": "order",
                      "value": (v9/*: any*/)
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
          (v11/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "alias": null,
                "args": (v7/*: any*/),
                "concreteType": "ProfileConnection",
                "kind": "LinkedField",
                "name": "profiles",
                "plural": false,
                "selections": [
                  (v8/*: any*/),
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
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Membership",
                            "kind": "LinkedField",
                            "name": "membership",
                            "plural": false,
                            "selections": [
                              (v11/*: any*/),
                              (v13/*: any*/),
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
                                    "value": "iam:membership-profile:delete"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"iam:membership-profile:delete\")"
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
                              (v14/*: any*/),
                              (v11/*: any*/)
                            ],
                            "storageKey": null
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
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v16/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v17/*: any*/),
                  (v18/*: any*/)
                ],
                "storageKey": "profiles(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"FULL_NAME\"})"
              },
              {
                "alias": null,
                "args": (v7/*: any*/),
                "filters": (v19/*: any*/),
                "handle": "connection",
                "key": "PeopleListFragment_profiles",
                "kind": "LinkedHandle",
                "name": "profiles"
              },
              {
                "alias": null,
                "args": (v10/*: any*/),
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "invitations",
                "plural": false,
                "selections": [
                  (v8/*: any*/),
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
                          (v11/*: any*/),
                          (v12/*: any*/),
                          (v14/*: any*/),
                          (v13/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "status",
                            "storageKey": null
                          },
                          (v15/*: any*/),
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
                      (v16/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v17/*: any*/),
                  (v18/*: any*/)
                ],
                "storageKey": "invitations(first:20,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v10/*: any*/),
                "filters": (v19/*: any*/),
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
    "cacheID": "3aef3e0e345bd04d2432c7167c87aa8d",
    "id": null,
    "metadata": {},
    "name": "PeoplePageQuery",
    "operationKind": "query",
    "text": "query PeoplePageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      canCreateUser: permission(action: \"iam:membership-profile:create\")\n      canInviteUser: permission(action: \"iam:invitation:create\")\n      profiles(first: 20, orderBy: {direction: ASC, field: FULL_NAME}) {\n        totalCount\n      }\n      ...PeopleListFragment_8lnpd\n      invitations(first: 20, orderBy: {direction: DESC, field: CREATED_AT}) {\n        totalCount\n        __typename\n      }\n      ...InvitationListFragment_1PypFi\n    }\n    id\n  }\n}\n\nfragment InvitationListFragment_1PypFi on Organization {\n  invitations(first: 20, orderBy: {direction: DESC, field: CREATED_AT}) {\n    edges {\n      node {\n        id\n        ...InvitationListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment InvitationListItemFragment on Invitation {\n  id\n  fullName\n  email\n  role\n  status\n  createdAt\n  acceptedAt\n  canDelete: permission(action: \"iam:invitation:delete\")\n}\n\nfragment PeopleListFragment_8lnpd on Organization {\n  profiles(first: 20, orderBy: {direction: ASC, field: FULL_NAME}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...PeopleListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment PeopleListItemFragment on Profile {\n  id\n  source\n  state\n  fullName\n  kind\n  position\n  membership {\n    id\n    role\n    canUpdate: permission(action: \"iam:membership:update\")\n    canDelete: permission(action: \"iam:membership-profile:delete\")\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n}\n"
  }
};
})();

(node as any).hash = "a89fa762bef0dcc31d05543a782f9aaf";

export default node;
