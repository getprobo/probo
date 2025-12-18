/**
 * @generated SignedSource<<c7e497311fbff76ab3b574012b9b9660>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationsQuery$variables = Record<PropertyKey, never>;
export type OrganizationsQuery$data = {
  readonly viewer: {
    readonly memberships: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: any;
          readonly id: string;
          readonly organization: {
            readonly name: string;
            readonly " $fragmentSpreads": FragmentRefs<"OrganizationCardFragment">;
          };
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
export type OrganizationsQuery = {
  response: OrganizationsQuery$data;
  variables: OrganizationsQuery$variables;
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
  "name": "createdAt",
  "storageKey": null
},
v3 = {
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
    "name": "OrganizationsQuery",
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
                        (v2/*: any*/),
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "Organization",
                          "kind": "LinkedField",
                          "name": "organization",
                          "plural": false,
                          "selections": [
                            (v3/*: any*/),
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "OrganizationCardFragment"
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
                          "args": null,
                          "kind": "FragmentSpread",
                          "name": "InvitationCardFragment"
                        }
                      ],
                      "storageKey": null
                    }
                  ],
                  "storageKey": null
                }
              ],
              "storageKey": "pendingInvitations(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
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
    "name": "OrganizationsQuery",
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Organization",
                        "kind": "LinkedField",
                        "name": "organization",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          (v1/*: any*/),
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Organization",
                        "kind": "LinkedField",
                        "name": "organization",
                        "plural": false,
                        "selections": [
                          (v1/*: any*/),
                          (v3/*: any*/)
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
    "cacheID": "145dddcd59c5c55a1c61913d9f43b38b",
    "id": null,
    "metadata": {},
    "name": "OrganizationsQuery",
    "operationKind": "query",
    "text": "query OrganizationsQuery {\n  viewer {\n    memberships(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          createdAt\n          organization {\n            name\n            ...OrganizationCardFragment\n            id\n          }\n        }\n      }\n    }\n    pendingInvitations(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          ...InvitationCardFragment\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment InvitationCardFragment on Invitation {\n  id\n  role\n  createdAt\n  organization {\n    id\n    name\n  }\n}\n\nfragment OrganizationCardFragment on Organization {\n  id\n  name\n  logoUrl\n}\n"
  }
};
})();

(node as any).hash = "f85219c0fcd69be1b5cfb0a5b9c9b8db";

export default node;
