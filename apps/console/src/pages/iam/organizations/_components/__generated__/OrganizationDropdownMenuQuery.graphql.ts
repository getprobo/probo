/**
 * @generated SignedSource<<034b9a96de4b1efda9529e8165779342>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationDropdownMenuQuery$variables = Record<PropertyKey, never>;
export type OrganizationDropdownMenuQuery$data = {
  readonly viewer: {
    readonly memberships: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly organization: {
            readonly name: string;
          };
          readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdownMenuItemFragment">;
        };
      }>;
    };
  };
};
export type OrganizationDropdownMenuQuery = {
  response: OrganizationDropdownMenuQuery$data;
  variables: OrganizationDropdownMenuQuery$variables;
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
    "name": "OrganizationDropdownMenuQuery",
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
                              },
                              {
                                "args": null,
                                "kind": "FragmentSpread",
                                "name": "OrganizationDropdownMenuItemFragment"
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
    "name": "OrganizationDropdownMenuQuery",
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
            "storageKey": "memberships(first:1000,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
          },
          (v1/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "b880e24e59e2bb500bbb2b9026423c20",
    "id": null,
    "metadata": {},
    "name": "OrganizationDropdownMenuQuery",
    "operationKind": "query",
    "text": "query OrganizationDropdownMenuQuery {\n  viewer {\n    memberships(first: 1000, orderBy: {direction: DESC, field: CREATED_AT}) {\n      edges {\n        node {\n          id\n          organization {\n            name\n            id\n          }\n          ...OrganizationDropdownMenuItemFragment\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment OrganizationDropdownMenuItemFragment on Membership {\n  id\n  lastSession {\n    id\n    expiresAt\n  }\n  organization {\n    id\n    logoUrl\n    name\n  }\n}\n"
  }
};
})();

(node as any).hash = "27c82735b5dcef01422d87312c127519";

export default node;
