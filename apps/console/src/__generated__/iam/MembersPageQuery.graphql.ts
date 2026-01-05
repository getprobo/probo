/**
 * @generated SignedSource<<d7435d1c1ee31d5ddc186fd67fa9401f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembersPageQuery$variables = {
  organizationId: string;
};
export type MembersPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly canInviteUser: boolean;
    readonly invitations: {
      readonly totalCount: number | null | undefined;
    };
    readonly members: {
      readonly totalCount: number | null | undefined;
    };
    readonly " $fragmentSpreads": FragmentRefs<"InvitationListFragment" | "InviteUserDialog_currentRoleFragment" | "MemberListFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type MembersPageQuery = {
  response: MembersPageQuery$data;
  variables: MembersPageQuery$variables;
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
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v4 = {
  "direction": "ASC",
  "field": "CREATED_AT"
},
v5 = [
  (v3/*: any*/),
  {
    "kind": "Literal",
    "name": "order",
    "value": (v4/*: any*/)
  }
],
v6 = [
  (v3/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v4/*: any*/)
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v8 = [
  (v7/*: any*/)
],
v9 = {
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
    "name": "MembersPageQuery",
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
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "InviteUserDialog_currentRoleFragment"
                },
                {
                  "args": (v5/*: any*/),
                  "kind": "FragmentSpread",
                  "name": "MemberListFragment"
                },
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "MembershipConnection",
                    "kind": "LinkedField",
                    "name": "members",
                    "plural": false,
                    "selections": (v8/*: any*/),
                    "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
                  },
                  "action": "THROW"
                },
                {
                  "args": (v5/*: any*/),
                  "kind": "FragmentSpread",
                  "name": "InvitationListFragment"
                },
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v6/*: any*/),
                    "concreteType": "InvitationConnection",
                    "kind": "LinkedField",
                    "name": "invitations",
                    "plural": false,
                    "selections": (v8/*: any*/),
                    "storageKey": "invitations(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
                  },
                  "action": "THROW"
                },
                (v9/*: any*/)
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
    "name": "MembersPageQuery",
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
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "viewerMembership",
                "plural": false,
                "selections": [
                  (v11/*: any*/),
                  (v10/*: any*/)
                ],
                "storageKey": null
              },
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
                              (v12/*: any*/),
                              (v10/*: any*/)
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
                "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "MemberListFragment_members",
                "kind": "LinkedHandle",
                "name": "members"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "invitations",
                "plural": false,
                "selections": [
                  (v7/*: any*/),
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
                            "name": "expiresAt",
                            "storageKey": null
                          },
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
                "storageKey": "invitations(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": (v18/*: any*/),
                "handle": "connection",
                "key": "InvitationListFragment_invitations",
                "kind": "LinkedHandle",
                "name": "invitations"
              },
              (v9/*: any*/)
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
    "cacheID": "9cc9ff1663db9c476c1cd67f097a4ab9",
    "id": null,
    "metadata": {},
    "name": "MembersPageQuery",
    "operationKind": "query",
    "text": "query MembersPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...InviteUserDialog_currentRoleFragment\n      ...MemberListFragment_1jRT0c\n      members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n        totalCount\n      }\n      ...InvitationListFragment_1jRT0c\n      invitations(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n        totalCount\n      }\n      canInviteUser: permission(action: \"iam:invitation:create\")\n    }\n    id\n  }\n}\n\nfragment InvitationListFragment_1jRT0c on Organization {\n  invitations(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...InvitationListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment InvitationListItemFragment on Invitation {\n  id\n  fullName\n  email\n  role\n  status\n  createdAt\n  expiresAt\n  acceptedAt\n  canDelete: permission(action: \"iam:invitation:delete\")\n}\n\nfragment InviteUserDialog_currentRoleFragment on Organization {\n  viewerMembership {\n    role\n    id\n  }\n}\n\nfragment MemberListFragment_1jRT0c on Organization {\n  members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...MemberListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment MemberListItemFragment on Membership {\n  id\n  role\n  source\n  state\n  profile {\n    fullName\n    id\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership:update\")\n  canDelete: permission(action: \"iam:membership:delete\")\n}\n"
  }
};
})();

(node as any).hash = "591a88a5e1b0bd427daff74dc27ac512";

export default node;
