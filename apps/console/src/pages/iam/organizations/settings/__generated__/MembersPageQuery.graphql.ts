/**
 * @generated SignedSource<<33faed6358cc7fba56a1f493c11f31d4>>
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
  readonly viewer: {
    readonly canInviteUser: boolean;
    readonly " $fragmentSpreads": FragmentRefs<"InvitationListItem_permissionsFragment" | "MemberListItem_permissionsFragment">;
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
v1 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
},
v2 = {
  "alias": "canInviteUser",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "iam:membership:create"
    },
    (v1/*: any*/)
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": null
},
v3 = [
  {
    "kind": "Variable",
    "name": "organizationId",
    "variableName": "organizationId"
  }
],
v4 = [
  (v1/*: any*/)
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v6 = {
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v7 = {
  "direction": "ASC",
  "field": "CREATED_AT"
},
v8 = [
  (v6/*: any*/),
  {
    "kind": "Literal",
    "name": "order",
    "value": (v7/*: any*/)
  }
],
v9 = [
  (v6/*: any*/),
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": (v7/*: any*/)
  }
],
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v11 = [
  (v10/*: any*/)
],
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
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
  "name": "fullName",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v18 = {
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MembersPageQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "Identity",
          "kind": "LinkedField",
          "name": "viewer",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            {
              "args": (v3/*: any*/),
              "kind": "FragmentSpread",
              "name": "InvitationListItem_permissionsFragment"
            },
            {
              "args": (v3/*: any*/),
              "kind": "FragmentSpread",
              "name": "MemberListItem_permissionsFragment"
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      },
      {
        "kind": "RequiredField",
        "field": {
          "alias": "organization",
          "args": (v4/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v5/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "InviteUserDialog_currentRoleFragment"
                },
                {
                  "args": (v8/*: any*/),
                  "kind": "FragmentSpread",
                  "name": "MemberListFragment"
                },
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": (v9/*: any*/),
                    "concreteType": "MembershipConnection",
                    "kind": "LinkedField",
                    "name": "members",
                    "plural": false,
                    "selections": (v11/*: any*/),
                    "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
                  },
                  "action": "THROW"
                },
                {
                  "args": (v8/*: any*/),
                  "kind": "FragmentSpread",
                  "name": "InvitationListFragment"
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
                    "selections": (v11/*: any*/),
                    "storageKey": "invitations(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
                  },
                  "action": "THROW"
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
    "name": "MembersPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          {
            "alias": "canDeleteInvitation",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "iam:invitation:delete"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canUpdateMembership",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "iam:membership:update"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canDeleteMembership",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "iam:membership:delete"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          (v12/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          (v12/*: any*/),
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
                  (v13/*: any*/),
                  (v12/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "MembershipConnection",
                "kind": "LinkedField",
                "name": "members",
                "plural": false,
                "selections": [
                  (v10/*: any*/),
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
                          (v12/*: any*/),
                          (v13/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "MembershipProfile",
                            "kind": "LinkedField",
                            "name": "profile",
                            "plural": false,
                            "selections": [
                              (v14/*: any*/),
                              (v12/*: any*/)
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
                              (v15/*: any*/),
                              (v12/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v16/*: any*/),
                          (v5/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v18/*: any*/),
                  (v19/*: any*/)
                ],
                "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v20/*: any*/),
                "handle": "connection",
                "key": "MemberListFragment_members",
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
                  (v10/*: any*/),
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
                          (v14/*: any*/),
                          (v15/*: any*/),
                          (v13/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "status",
                            "storageKey": null
                          },
                          (v16/*: any*/),
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
                          (v5/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v17/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v18/*: any*/),
                  (v19/*: any*/)
                ],
                "storageKey": "invitations(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": (v20/*: any*/),
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
    "cacheID": "1c006f49a510a658f66a78c66e9289db",
    "id": null,
    "metadata": {},
    "name": "MembersPageQuery",
    "operationKind": "query",
    "text": "query MembersPageQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    canInviteUser: permission(action: \"iam:membership:create\", id: $organizationId)\n    ...InvitationListItem_permissionsFragment_4xMPKw\n    ...MemberListItem_permissionsFragment_4xMPKw\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...InviteUserDialog_currentRoleFragment\n      ...MemberListFragment_1jRT0c\n      members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n        totalCount\n      }\n      ...InvitationListFragment_1jRT0c\n      invitations(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n        totalCount\n      }\n    }\n    id\n  }\n}\n\nfragment InvitationListFragment_1jRT0c on Organization {\n  invitations(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...InvitationListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment InvitationListItemFragment on Invitation {\n  id\n  fullName\n  email\n  role\n  status\n  createdAt\n  expiresAt\n  acceptedAt\n}\n\nfragment InvitationListItem_permissionsFragment_4xMPKw on Identity {\n  canDeleteInvitation: permission(action: \"iam:invitation:delete\", id: $organizationId)\n}\n\nfragment InviteUserDialog_currentRoleFragment on Organization {\n  viewerMembership {\n    role\n    id\n  }\n}\n\nfragment MemberListFragment_1jRT0c on Organization {\n  ...MemberListItem_currentRoleFragment\n  members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...MemberListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment MemberListItemFragment on Membership {\n  id\n  role\n  profile {\n    fullName\n    id\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n}\n\nfragment MemberListItem_currentRoleFragment on Organization {\n  viewerMembership {\n    role\n    id\n  }\n}\n\nfragment MemberListItem_permissionsFragment_4xMPKw on Identity {\n  canUpdateMembership: permission(action: \"iam:membership:update\", id: $organizationId)\n  canDeleteMembership: permission(action: \"iam:membership:delete\", id: $organizationId)\n}\n"
  }
};
})();

(node as any).hash = "6e919357c168bbfbdccfa7375ce71660";

export default node;
