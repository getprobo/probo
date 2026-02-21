/**
 * @generated SignedSource<<63cc59b828971b5b1c0983e2a439f95d>>
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
          readonly organization: {
            readonly name: string;
            readonly " $fragmentSpreads": FragmentRefs<"MembershipCard_organizationFragment">;
          };
          readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
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
  "name": "filter",
  "value": {
    "state": "ACTIVE"
  }
},
v1 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "ASC",
    "field": "ORGANIZATION_NAME"
  }
},
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
v7 = [
  (v0/*: any*/),
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  (v1/*: any*/)
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
                  (v0/*: any*/),
                  (v1/*: any*/)
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
                                  (v3/*: any*/),
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
                  (v6/*: any*/)
                ],
                "storageKey": "__MembershipsPage_profiles_connection(filter:{\"state\":\"ACTIVE\"},orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
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
            "args": (v7/*: any*/),
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
                      (v2/*: any*/),
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
                          (v2/*: any*/)
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
                          (v3/*: any*/),
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
                      },
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v5/*: any*/)
                ],
                "storageKey": null
              },
              (v6/*: any*/)
            ],
            "storageKey": "profiles(filter:{\"state\":\"ACTIVE\"},first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
          },
          {
            "alias": null,
            "args": (v7/*: any*/),
            "filters": [
              "orderBy",
              "filter"
            ],
            "handle": "connection",
            "key": "MembershipsPage_profiles",
            "kind": "LinkedHandle",
            "name": "profiles"
          },
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "70c954537c603303bcd55418f6bd73cf",
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
    "text": "query MembershipsPageQuery {\n  viewer {\n    profiles(first: 1000, orderBy: {direction: ASC, field: ORGANIZATION_NAME}, filter: {state: ACTIVE}) {\n      edges {\n        node {\n          id\n          ...MembershipCardFragment\n          organization {\n            name\n            ...MembershipCard_organizationFragment\n            id\n          }\n          __typename\n        }\n        cursor\n      }\n      pageInfo {\n        endCursor\n        hasNextPage\n      }\n    }\n    id\n  }\n}\n\nfragment MembershipCardFragment on Profile {\n  state\n  membership {\n    lastSession {\n      id\n      expiresAt\n    }\n    id\n  }\n}\n\nfragment MembershipCard_organizationFragment on Organization {\n  id\n  name\n  logoUrl\n}\n"
  }
};
})();

(node as any).hash = "c82587c425f688858d4625426153dbdc";

export default node;
