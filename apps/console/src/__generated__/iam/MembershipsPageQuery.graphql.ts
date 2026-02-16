/**
 * @generated SignedSource<<68acd1fa60fc1efe40e3983441ae01da>>
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
    readonly profiles: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly membership: {
            readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
          };
          readonly organization: {
            readonly name: string;
            readonly " $fragmentSpreads": FragmentRefs<"MembershipCard_organizationFragment">;
          };
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
    "direction": "ASC",
    "field": "ORGANIZATION_NAME"
  }
},
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
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v5 = {
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
v6 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  (v0/*: any*/)
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
                "alias": "profiles",
                "args": [
                  (v0/*: any*/)
                ],
                "concreteType": "ProfileConnection",
                "kind": "LinkedField",
                "name": "__MembershipsPage_profiles_connection",
                "plural": false,
                "selections": [
                  {
                    "kind": "RequiredField",
                    "field": {
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
                            (v1/*: any*/),
                            {
                              "kind": "RequiredField",
                              "field": {
                                "alias": null,
                                "args": null,
                                "concreteType": "Membership",
                                "kind": "LinkedField",
                                "name": "membership",
                                "plural": false,
                                "selections": [
                                  {
                                    "args": null,
                                    "kind": "FragmentSpread",
                                    "name": "MembershipCardFragment"
                                  }
                                ],
                                "storageKey": null
                              },
                              "action": "THROW"
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
                                  (v2/*: any*/),
                                  {
                                    "args": null,
                                    "kind": "FragmentSpread",
                                    "name": "MembershipCard_organizationFragment"
                                  }
                                ],
                                "storageKey": null
                              },
                              "action": "THROW"
                            },
                            (v3/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v4/*: any*/)
                      ],
                      "storageKey": null
                    },
                    "action": "THROW"
                  },
                  (v5/*: any*/)
                ],
                "storageKey": "__MembershipsPage_profiles_connection(orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
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
            "args": (v6/*: any*/),
            "concreteType": "ProfileConnection",
            "kind": "LinkedField",
            "name": "profiles",
            "plural": false,
            "selections": [
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
                      (v1/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Membership",
                        "kind": "LinkedField",
                        "name": "membership",
                        "plural": false,
                        "selections": [
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
                          (v1/*: any*/)
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
                      },
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
            "storageKey": "profiles(first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
          },
          {
            "alias": null,
            "args": (v6/*: any*/),
            "filters": [
              "orderBy"
            ],
            "handle": "connection",
            "key": "MembershipsPage_profiles",
            "kind": "LinkedHandle",
            "name": "profiles"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "d3ee4e9e305c7a2f57179cf540f4cd6c",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "viewer",
            "profiles"
          ]
        }
      ]
    },
    "name": "MembershipsPageQuery",
    "operationKind": "query",
    "text": "query MembershipsPageQuery {\n  viewer {\n    profiles(first: 1000, orderBy: {direction: ASC, field: ORGANIZATION_NAME}) {\n      edges {\n        node {\n          id\n          membership {\n            ...MembershipCardFragment\n            id\n          }\n          organization {\n            name\n            ...MembershipCard_organizationFragment\n            id\n          }\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    id\n  }\n}\n\nfragment MembershipCardFragment on Membership {\n  lastSession {\n    id\n    expiresAt\n  }\n}\n\nfragment MembershipCard_organizationFragment on Organization {\n  id\n  name\n  logoUrl\n}\n"
  }
};
})();

(node as any).hash = "c87fe091011b4d5e5bffe0309eb4211f";

export default node;
