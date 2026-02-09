/**
 * @generated SignedSource<<e59dc83c1f91e928764947766661f044>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SCIMSettingsPageQuery$variables = {
  organizationId: string;
};
export type SCIMSettingsPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly id: string;
    readonly scimConfiguration: {
      readonly bridge: {
        readonly id: string;
      } | null | undefined;
      readonly id: string;
      readonly " $fragmentSpreads": FragmentRefs<"SCIMEventListFragment">;
    } | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"ConnectorListFragment" | "SCIMConfigurationFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type SCIMSettingsPageQuery = {
  response: SCIMSettingsPageQuery$data;
  variables: SCIMSettingsPageQuery$variables;
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v5 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 20
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SCIMSettingsPageQuery",
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
                (v3/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "SCIMConfiguration",
                  "kind": "LinkedField",
                  "name": "scimConfiguration",
                  "plural": false,
                  "selections": [
                    (v3/*: any*/),
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "SCIMBridge",
                      "kind": "LinkedField",
                      "name": "bridge",
                      "plural": false,
                      "selections": [
                        (v3/*: any*/)
                      ],
                      "storageKey": null
                    },
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "SCIMEventListFragment"
                    }
                  ],
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "SCIMConfigurationFragment"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "ConnectorListFragment"
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
    "name": "SCIMSettingsPageQuery",
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
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "SCIMConfiguration",
                "kind": "LinkedField",
                "name": "scimConfiguration",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SCIMBridge",
                    "kind": "LinkedField",
                    "name": "bridge",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "excludedUserNames",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Connector",
                        "kind": "LinkedField",
                        "name": "connector",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          (v4/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "concreteType": "SCIMEventConnection",
                    "kind": "LinkedField",
                    "name": "events",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "SCIMEventEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "SCIMEvent",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "method",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "path",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "statusCode",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "errorMessage",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "ipAddress",
                                "storageKey": null
                              },
                              (v4/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Membership",
                                "kind": "LinkedField",
                                "name": "membership",
                                "plural": false,
                                "selections": [
                                  (v3/*: any*/),
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
                                      (v3/*: any*/)
                                    ],
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
                      }
                    ],
                    "storageKey": "events(first:20)"
                  },
                  {
                    "alias": null,
                    "args": (v5/*: any*/),
                    "filters": null,
                    "handle": "connection",
                    "key": "SCIMEventListFragment_events",
                    "kind": "LinkedHandle",
                    "name": "events"
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "endpointUrl",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              },
              {
                "alias": "canCreateSCIMConfiguration",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "iam:scim-configuration:create"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"iam:scim-configuration:create\")"
              },
              {
                "alias": "canDeleteSCIMConfiguration",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "iam:scim-configuration:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"iam:scim-configuration:delete\")"
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
    "cacheID": "1074234a9a3734270b5da7420dea2ec6",
    "id": null,
    "metadata": {},
    "name": "SCIMSettingsPageQuery",
    "operationKind": "query",
    "text": "query SCIMSettingsPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      scimConfiguration {\n        id\n        bridge {\n          id\n        }\n        ...SCIMEventListFragment\n      }\n      ...SCIMConfigurationFragment\n      ...ConnectorListFragment\n    }\n    id\n  }\n}\n\nfragment ConnectorListFragment on Organization {\n  scimConfiguration {\n    ...GoogleWorkspaceConnectorFragment\n    id\n  }\n}\n\nfragment GoogleWorkspaceConnectorFragment on SCIMConfiguration {\n  id\n  bridge {\n    id\n    excludedUserNames\n    connector {\n      id\n      createdAt\n    }\n  }\n}\n\nfragment SCIMConfigurationFragment on Organization {\n  canCreateSCIMConfiguration: permission(action: \"iam:scim-configuration:create\")\n  canDeleteSCIMConfiguration: permission(action: \"iam:scim-configuration:delete\")\n  scimConfiguration {\n    id\n    endpointUrl\n    bridge {\n      id\n    }\n  }\n}\n\nfragment SCIMEventListFragment on SCIMConfiguration {\n  events(first: 20) {\n    edges {\n      node {\n        id\n        ...SCIMEventListItemFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment SCIMEventListItemFragment on SCIMEvent {\n  id\n  method\n  path\n  statusCode\n  errorMessage\n  ipAddress\n  createdAt\n  membership {\n    id\n    profile {\n      fullName\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9db69676a14d2a6c4062fee58ce728e1";

export default node;
