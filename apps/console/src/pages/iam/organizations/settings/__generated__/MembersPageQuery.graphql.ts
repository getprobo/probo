/**
 * @generated SignedSource<<e16e82aa25ced17faf9e4c40f7e6130e>>
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
    readonly members: {
      readonly totalCount: number | null | undefined;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MemberListFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly viewer: {
    readonly " $fragmentSpreads": FragmentRefs<"MemberListItem_permissionsFragment">;
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
v2 = [
  (v1/*: any*/)
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 20
},
v5 = {
  "direction": "ASC",
  "field": "CREATED_AT"
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "role",
  "storageKey": null
};
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
            {
              "args": [
                {
                  "kind": "Variable",
                  "name": "organizationId",
                  "variableName": "organizationId"
                }
              ],
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
          "args": (v2/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v3/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
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
                    "selections": [
                      (v7/*: any*/)
                    ],
                    "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
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
          (v8/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          (v8/*: any*/),
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
                  (v9/*: any*/),
                  (v8/*: any*/)
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
                          (v8/*: any*/),
                          (v9/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "MembershipProfile",
                            "kind": "LinkedField",
                            "name": "profile",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "fullName",
                                "storageKey": null
                              },
                              (v8/*: any*/)
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
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "email",
                                "storageKey": null
                              },
                              (v8/*: any*/)
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "createdAt",
                            "storageKey": null
                          },
                          (v3/*: any*/)
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
                "storageKey": "members(first:20,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
              },
              {
                "alias": null,
                "args": (v6/*: any*/),
                "filters": [
                  "orderBy"
                ],
                "handle": "connection",
                "key": "MembersListFragment_members",
                "kind": "LinkedHandle",
                "name": "members"
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
    "cacheID": "b125343184bf4d82a7d0e264faf7ac34",
    "id": null,
    "metadata": {},
    "name": "MembersPageQuery",
    "operationKind": "query",
    "text": "query MembersPageQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    ...MemberListItem_permissionsFragment_4xMPKw\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...MemberListFragment_1jRT0c\n      members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n        totalCount\n      }\n    }\n    id\n  }\n}\n\nfragment MemberListFragment_1jRT0c on Organization {\n  ...MemberListItem_currentRoleFragment\n  members(first: 20, orderBy: {direction: ASC, field: CREATED_AT}) {\n    totalCount\n    edges {\n      node {\n        id\n        ...MemberListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment MemberListItemFragment on Membership {\n  id\n  role\n  profile {\n    fullName\n    id\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n}\n\nfragment MemberListItem_currentRoleFragment on Organization {\n  viewerMembership {\n    role\n    id\n  }\n}\n\nfragment MemberListItem_permissionsFragment_4xMPKw on Identity {\n  canUpdateMembership: permission(action: \"iam:membership:update\", id: $organizationId)\n  canDeleteMembership: permission(action: \"iam:membership:delete\", id: $organizationId)\n}\n"
  }
};
})();

(node as any).hash = "8c40640ce96580ebae677d117741094e";

export default node;
