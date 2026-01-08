/**
 * @generated SignedSource<<1d247c2fede04c2292000421a9e49fa8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsPageQuery$variables = Record<PropertyKey, never>;
export type MembershipsPageQuery$data = {
  readonly viewer: {
    readonly memberships: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly organization: {
            readonly name: string;
          };
          readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
        };
      }>;
    };
    readonly pendingInvitations: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"InvitationCardFragment">;
        };
      }>;
    };
  };
};
export type MembershipsPageQuery = {
  response: MembershipsPageQuery$data;
  variables: MembershipsPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "DESC",
    "field": "CREATED_AT"
  }
},
v1 = [
  (v0/*: any*/)
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
  "name": "__typename",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v6 = {
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
v7 = {
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
v8 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  (v0/*: any*/)
],
v9 = [
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "MembershipsPageQuery",
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
              "kind": "RequiredField",
              "field": {
                "alias": "memberships",
                "args": (v1/*: any*/),
                "concreteType": "MembershipConnection",
                "kind": "LinkedField",
                "name": "__MembershipsPage_memberships_connection",
                "plural": false,
                "selections": [
                  {
                    "kind": "RequiredField",
                    "field": {
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
                            (v2/*: any*/),
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "MembershipCardFragment"
                            },
                            {
                              "kind": "RequiredField",
                              "field": {
                                "alias": null,
                                "args": null,
                                "concreteType": "Organization",
                                "kind": "LinkedField",
                                "name": "organization",
                                "plural": false,
                                "selections": [
                                  (v3/*: any*/)
                                ],
                                "storageKey": null
                              },
                              "action": "THROW"
                            },
                            (v4/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v5/*: any*/)
                      ],
                      "storageKey": null
                    },
                    "action": "THROW"
                  },
                  (v6/*: any*/),
                  (v7/*: any*/)
                ],
                "storageKey": "__MembershipsPage_memberships_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              "action": "THROW"
            },
            {
              "kind": "RequiredField",
              "field": {
                "alias": "pendingInvitations",
                "args": (v1/*: any*/),
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "__MembershipsPage_pendingInvitations_connection",
                "plural": false,
                "selections": [
                  {
                    "kind": "RequiredField",
                    "field": {
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
                            (v2/*: any*/),
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "InvitationCardFragment"
                            },
                            (v4/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v5/*: any*/)
                      ],
                      "storageKey": null
                    },
                    "action": "THROW"
                  },
                  (v6/*: any*/),
                  (v7/*: any*/)
                ],
                "storageKey": "__MembershipsPage_pendingInvitations_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              "action": "THROW"
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
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "MembershipsPageQuery",
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
            "alias": null,
            "args": (v8/*: any*/),
            "concreteType": "MembershipConnection",
            "kind": "LinkedField",
            "name": "memberships",
            "plural": false,
            "selections": [
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Session",
                        "kind": "LinkedField",
                        "name": "lastSession",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "expiresAt",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Organization",
                        "kind": "LinkedField",
                        "name": "organization",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "logoUrl",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      },
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v5/*: any*/)
                ],
                "storageKey": null
              },
              (v6/*: any*/),
              (v7/*: any*/)
            ],
            "storageKey": "memberships(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          {
            "alias": null,
            "args": (v8/*: any*/),
            "filters": (v9/*: any*/),
            "handle": "connection",
            "key": "MembershipsPage_memberships",
            "kind": "LinkedHandle",
            "name": "memberships"
          },
          {
            "alias": null,
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "role",
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
                        "alias": null,
                        "args": null,
                        "concreteType": "Organization",
                        "kind": "LinkedField",
                        "name": "organization",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v5/*: any*/)
                ],
                "storageKey": null
              },
              (v6/*: any*/),
              (v7/*: any*/)
            ],
            "storageKey": "pendingInvitations(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          {
            "alias": null,
            "args": (v8/*: any*/),
            "filters": (v9/*: any*/),
            "handle": "connection",
            "key": "MembershipsPage_pendingInvitations",
            "kind": "LinkedHandle",
            "name": "pendingInvitations"
          },
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "a3dc3f3e775fa02b7d181fa57ade06b9",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "viewer",
            "memberships"
          ]
        },
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "viewer",
            "pendingInvitations"
          ]
        }
      ]
    },
    "name": "MembershipsPageQuery",
    "operationKind": "query",
    "text": "query MembershipsPageQuery {\n  viewer {\n    memberships(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          ...MembershipCardFragment\n          organization {\n            name\n            id\n          }\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    pendingInvitations(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          ...InvitationCardFragment\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    id\n  }\n}\n\nfragment InvitationCardFragment on Invitation {\n  id\n  role\n  createdAt\n  organization {\n    id\n    name\n  }\n}\n\nfragment MembershipCardFragment on Membership {\n  lastSession {\n    id\n    expiresAt\n  }\n  organization {\n    id\n    name\n    logoUrl\n  }\n}\n"
  }
};
})();

(node as any).hash = "b21b9cae0ab7a4afdf12fb6ae516b64e";

export default node;
