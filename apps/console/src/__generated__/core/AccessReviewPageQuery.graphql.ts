/**
 * @generated SignedSource<<aa0ebb6a0911f903c5dc0483a244b175>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AccessReviewPageQuery$variables = {
  organizationId: string;
};
export type AccessReviewPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly accessReview: {
      readonly canCreateCampaign: boolean;
      readonly canCreateSource: boolean;
      readonly canUpdateAccessReview: boolean;
      readonly id: string;
      readonly identitySource: {
        readonly id: string;
      } | null | undefined;
      readonly " $fragmentSpreads": FragmentRefs<"AccessReviewPageFragment" | "AccessReviewPageSourcesFragment">;
    } | null | undefined;
    readonly id: string;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type AccessReviewPageQuery = {
  response: AccessReviewPageQuery$data;
  variables: AccessReviewPageQuery$variables;
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
  "concreteType": "AccessSource",
  "kind": "LinkedField",
  "name": "identitySource",
  "plural": false,
  "selections": [
    (v3/*: any*/)
  ],
  "storageKey": null
},
v5 = {
  "alias": "canCreateSource",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-source:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-source:create\")"
},
v6 = {
  "alias": "canCreateCampaign",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-review-campaign:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-review-campaign:create\")"
},
v7 = {
  "alias": "canUpdateAccessReview",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:access-review:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:access-review:update\")"
},
v8 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
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
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v12 = {
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
v13 = {
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
v14 = [
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "AccessReviewPageQuery",
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
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "AccessReview",
                "kind": "LinkedField",
                "name": "accessReview",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "AccessReviewPageSourcesFragment"
                  },
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "AccessReviewPageFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AccessReviewPageQuery",
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
                "concreteType": "AccessReview",
                "kind": "LinkedField",
                "name": "accessReview",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  {
                    "alias": null,
                    "args": (v8/*: any*/),
                    "concreteType": "AccessSourceConnection",
                    "kind": "LinkedField",
                    "name": "accessSources",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "AccessSourceEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "AccessSource",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v9/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "connectorId",
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
                                  {
                                    "alias": null,
                                    "args": null,
                                    "kind": "ScalarField",
                                    "name": "provider",
                                    "storageKey": null
                                  },
                                  (v3/*: any*/)
                                ],
                                "storageKey": null
                              },
                              (v10/*: any*/),
                              {
                                "alias": "canDelete",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:access-source:delete"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:access-source:delete\")"
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v11/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v12/*: any*/),
                      (v13/*: any*/)
                    ],
                    "storageKey": "accessSources(first:50,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  },
                  {
                    "alias": null,
                    "args": (v8/*: any*/),
                    "filters": (v14/*: any*/),
                    "handle": "connection",
                    "key": "AccessReviewPage_accessSources",
                    "kind": "LinkedHandle",
                    "name": "accessSources"
                  },
                  {
                    "alias": null,
                    "args": (v8/*: any*/),
                    "concreteType": "AccessReviewCampaignConnection",
                    "kind": "LinkedField",
                    "name": "campaigns",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "AccessReviewCampaignEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "AccessReviewCampaign",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v9/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "status",
                                "storageKey": null
                              },
                              (v10/*: any*/),
                              {
                                "alias": "canDelete",
                                "args": [
                                  {
                                    "kind": "Literal",
                                    "name": "action",
                                    "value": "core:access-review-campaign:delete"
                                  }
                                ],
                                "kind": "ScalarField",
                                "name": "permission",
                                "storageKey": "permission(action:\"core:access-review-campaign:delete\")"
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v11/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v12/*: any*/),
                      (v13/*: any*/)
                    ],
                    "storageKey": "campaigns(first:50,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  },
                  {
                    "alias": null,
                    "args": (v8/*: any*/),
                    "filters": (v14/*: any*/),
                    "handle": "connection",
                    "key": "AccessReviewPage_campaigns",
                    "kind": "LinkedHandle",
                    "name": "campaigns"
                  }
                ],
                "storageKey": null
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
    "cacheID": "b8bf4fa845016ad52c25746279a79e0f",
    "id": null,
    "metadata": {},
    "name": "AccessReviewPageQuery",
    "operationKind": "query",
    "text": "query AccessReviewPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      accessReview {\n        id\n        identitySource {\n          id\n        }\n        canCreateSource: permission(action: \"core:access-source:create\")\n        canCreateCampaign: permission(action: \"core:access-review-campaign:create\")\n        canUpdateAccessReview: permission(action: \"core:access-review:update\")\n        ...AccessReviewPageSourcesFragment\n        ...AccessReviewPageFragment\n      }\n    }\n    id\n  }\n}\n\nfragment AccessReviewCampaignRowFragment on AccessReviewCampaign {\n  id\n  name\n  status\n  createdAt\n  canDelete: permission(action: \"core:access-review-campaign:delete\")\n}\n\nfragment AccessReviewPageFragment on AccessReview {\n  campaigns(first: 50, orderBy: {direction: DESC, field: CREATED_AT}) {\n    edges {\n      node {\n        id\n        ...AccessReviewCampaignRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment AccessReviewPageSourcesFragment on AccessReview {\n  accessSources(first: 50, orderBy: {direction: DESC, field: CREATED_AT}) {\n    edges {\n      node {\n        id\n        name\n        connectorId\n        connector {\n          provider\n          id\n        }\n        ...AccessSourceRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment AccessSourceRowFragment on AccessSource {\n  id\n  name\n  connectorId\n  connector {\n    provider\n    id\n  }\n  createdAt\n  canDelete: permission(action: \"core:access-source:delete\")\n}\n"
  }
};
})();

(node as any).hash = "35582045d055bf703c8eac32c266d8b9";

export default node;
