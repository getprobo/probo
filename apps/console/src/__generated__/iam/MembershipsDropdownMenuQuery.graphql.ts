/**
 * @generated SignedSource<<d7c7a74300c4b145c88eca24f58a9e55>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdownMenuQuery$variables = Record<PropertyKey, never>;
export type MembershipsDropdownMenuQuery$data = {
  readonly viewer: {
    readonly profiles: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly membership: {
            readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdownMenuItemFragment">;
          };
          readonly organization: {
            readonly name: string;
            readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdownMenuItem_organizationFragment">;
          };
        };
      }>;
    };
  };
};
export type MembershipsDropdownMenuQuery = {
  response: MembershipsDropdownMenuQuery$data;
  variables: MembershipsDropdownMenuQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "kind": "Literal",
    "name": "filter",
    "value": {
      "state": "ACTIVE"
    }
  },
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "ORGANIZATION_NAME"
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
    "name": "MembershipsDropdownMenuQuery",
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
                "concreteType": "ProfileConnection",
                "kind": "LinkedField",
                "name": "profiles",
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
                          "kind": "RequiredField",
                          "field": {
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
                                  "concreteType": "Organization",
                                  "kind": "LinkedField",
                                  "name": "organization",
                                  "plural": false,
                                  "selections": [
                                    (v2/*: any*/),
                                    {
                                      "args": null,
                                      "kind": "FragmentSpread",
                                      "name": "MembershipsDropdownMenuItem_organizationFragment"
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
                                  "concreteType": "Membership",
                                  "kind": "LinkedField",
                                  "name": "membership",
                                  "plural": false,
                                  "selections": [
                                    {
                                      "args": null,
                                      "kind": "FragmentSpread",
                                      "name": "MembershipsDropdownMenuItemFragment"
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
                      "storageKey": null
                    },
                    "action": "THROW"
                  }
                ],
                "storageKey": "profiles(filter:{\"state\":\"ACTIVE\"},first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
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
    "name": "MembershipsDropdownMenuQuery",
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
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Membership",
                        "kind": "LinkedField",
                        "name": "membership",
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
            "storageKey": "profiles(filter:{\"state\":\"ACTIVE\"},first:1000,orderBy:{\"direction\":\"ASC\",\"field\":\"ORGANIZATION_NAME\"})"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "b62c27cc3a20f286290bf579d2f528d3",
    "id": null,
    "metadata": {},
    "name": "MembershipsDropdownMenuQuery",
    "operationKind": "query",
    "text": "query MembershipsDropdownMenuQuery {\n  viewer {\n    profiles(first: 1000, orderBy: {direction: ASC, field: ORGANIZATION_NAME}, filter: {state: ACTIVE}) {\n      edges {\n        node {\n          id\n          organization {\n            name\n            ...MembershipsDropdownMenuItem_organizationFragment\n            id\n          }\n          membership {\n            ...MembershipsDropdownMenuItemFragment\n            id\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment MembershipsDropdownMenuItemFragment on Membership {\n  id\n  lastSession {\n    id\n    expiresAt\n  }\n}\n\nfragment MembershipsDropdownMenuItem_organizationFragment on Organization {\n  id\n  name\n  logoUrl\n}\n"
  }
};
})();

(node as any).hash = "b51ef962fc8bf18c380b7373135373a3";

export default node;
