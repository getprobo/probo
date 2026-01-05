/**
 * @generated SignedSource<<ab48c4d454e4f9fa4472a103b8144b7c>>
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
var v0 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
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
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
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
                "alias": null,
                "args": (v0/*: any*/),
                "concreteType": "MembershipConnection",
                "kind": "LinkedField",
                "name": "memberships",
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
                          "kind": "RequiredField",
                          "field": {
                            "alias": null,
                            "args": null,
                            "concreteType": "Membership",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v1/*: any*/),
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
                                    (v2/*: any*/)
                                  ],
                                  "storageKey": null
                                },
                                "action": "THROW"
                              }
                            ],
                            "storageKey": null
                          },
                          "action": "THROW"
                        }
                      ],
                      "storageKey": null
                    },
                    "action": "THROW"
                  }
                ],
                "storageKey": "memberships(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              },
              "action": "THROW"
            },
            {
              "kind": "RequiredField",
              "field": {
                "alias": null,
                "args": (v0/*: any*/),
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "pendingInvitations",
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
                          "kind": "RequiredField",
                          "field": {
                            "alias": null,
                            "args": null,
                            "concreteType": "Invitation",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v1/*: any*/),
                              {
                                "args": null,
                                "kind": "FragmentSpread",
                                "name": "InvitationCardFragment"
                              }
                            ],
                            "storageKey": null
                          },
                          "action": "THROW"
                        }
                      ],
                      "storageKey": null
                    },
                    "action": "THROW"
                  }
                ],
                "storageKey": "pendingInvitations(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
            "args": (v0/*: any*/),
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
                      (v1/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Session",
                        "kind": "LinkedField",
                        "name": "lastSession",
                        "plural": false,
                        "selections": [
                          (v1/*: any*/),
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
                          (v1/*: any*/),
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "logoUrl",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "memberships(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          {
            "alias": null,
            "args": (v0/*: any*/),
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
                      (v1/*: any*/),
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
                          (v1/*: any*/),
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "pendingInvitations(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "70ba2b8cec46ba8154cf7a6a51e03932",
    "id": null,
    "metadata": {},
    "name": "MembershipsPageQuery",
    "operationKind": "query",
    "text": "query MembershipsPageQuery {\n  viewer {\n    memberships(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          ...MembershipCardFragment\n          organization {\n            name\n            id\n          }\n        }\n      }\n    }\n    pendingInvitations(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          ...InvitationCardFragment\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment InvitationCardFragment on Invitation {\n  id\n  role\n  createdAt\n  organization {\n    id\n    name\n  }\n}\n\nfragment MembershipCardFragment on Membership {\n  lastSession {\n    id\n    expiresAt\n  }\n  organization {\n    id\n    name\n    logoUrl\n  }\n}\n"
  }
};
})();

(node as any).hash = "4148521eb3eaba1670a2d181b6d2468f";

export default node;
