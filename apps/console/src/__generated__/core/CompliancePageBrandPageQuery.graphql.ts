/**
 * @generated SignedSource<<5e625eeec03803e3e72c6dd998bed003>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageBrandPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageBrandPageQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly compliancePage: {
      readonly canUpdate: boolean;
      readonly darkLogoFileUrl: string | null | undefined;
      readonly id: string;
      readonly logoFileUrl: string | null | undefined;
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageExternalUrlsSection_trustCenterFragment" | "CompliancePageFrameworkList_compliancePageFragment">;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type CompliancePageBrandPageQuery = {
  response: CompliancePageBrandPageQuery$data;
  variables: CompliancePageBrandPageQuery$variables;
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
  "name": "logoFileUrl",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "darkLogoFileUrl",
  "storageKey": null
},
v6 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center:update\")"
},
v7 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "ASC",
      "field": "RANK"
    }
  }
],
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "rank",
  "storageKey": null
},
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
  "name": "cursor",
  "storageKey": null
},
v11 = {
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
v12 = [
  "orderBy"
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageBrandPageQuery",
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
              {
                "kind": "RequiredField",
                "field": {
                  "alias": "compliancePage",
                  "args": null,
                  "concreteType": "TrustCenter",
                  "kind": "LinkedField",
                  "name": "trustCenter",
                  "plural": false,
                  "selections": [
                    (v3/*: any*/),
                    (v4/*: any*/),
                    (v5/*: any*/),
                    (v6/*: any*/),
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageFrameworkList_compliancePageFragment"
                    },
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "CompliancePageExternalUrlsSection_trustCenterFragment"
                    }
                  ],
                  "storageKey": null
                },
                "action": "THROW"
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
    "name": "CompliancePageBrandPageQuery",
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
              {
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "concreteType": "ComplianceFrameworkConnection",
                    "kind": "LinkedField",
                    "name": "complianceFrameworks",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceFrameworkEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "ComplianceFramework",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v8/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "visibility",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Framework",
                                "kind": "LinkedField",
                                "name": "framework",
                                "plural": false,
                                "selections": [
                                  (v3/*: any*/),
                                  (v9/*: any*/),
                                  {
                                    "alias": null,
                                    "args": null,
                                    "kind": "ScalarField",
                                    "name": "lightLogoURL",
                                    "storageKey": null
                                  },
                                  {
                                    "alias": null,
                                    "args": null,
                                    "kind": "ScalarField",
                                    "name": "darkLogoURL",
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v10/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v11/*: any*/)
                    ],
                    "storageKey": "complianceFrameworks(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "filters": (v12/*: any*/),
                    "handle": "connection",
                    "key": "CompliancePageFrameworkList_complianceFrameworks",
                    "kind": "LinkedHandle",
                    "name": "complianceFrameworks"
                  },
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "concreteType": "ComplianceExternalURLConnection",
                    "kind": "LinkedField",
                    "name": "externalUrls",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "ComplianceExternalURLEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "ComplianceExternalURL",
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
                                "name": "url",
                                "storageKey": null
                              },
                              (v8/*: any*/),
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v10/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v11/*: any*/),
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
                    "storageKey": "externalUrls(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"RANK\"})"
                  },
                  {
                    "alias": null,
                    "args": (v7/*: any*/),
                    "filters": (v12/*: any*/),
                    "handle": "connection",
                    "key": "CompliancePageExternalUrlsSection_externalUrls",
                    "kind": "LinkedHandle",
                    "name": "externalUrls"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "8a6a8094cbc0edebe011b284ac9638dc",
    "id": null,
    "metadata": {},
    "name": "CompliancePageBrandPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageBrandPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        id\n        logoFileUrl\n        darkLogoFileUrl\n        canUpdate: permission(action: \"core:trust-center:update\")\n        ...CompliancePageFrameworkList_compliancePageFragment\n        ...CompliancePageExternalUrlsSection_trustCenterFragment\n      }\n    }\n    id\n  }\n}\n\nfragment CompliancePageExternalUrlsSection_trustCenterFragment on TrustCenter {\n  id\n  canUpdate: permission(action: \"core:trust-center:update\")\n  externalUrls(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        name\n        url\n        rank\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment CompliancePageFrameworkList_compliancePageFragment on TrustCenter {\n  id\n  canUpdate: permission(action: \"core:trust-center:update\")\n  complianceFrameworks(first: 100, orderBy: {field: RANK, direction: ASC}) {\n    edges {\n      node {\n        id\n        rank\n        visibility\n        framework {\n          id\n          name\n          lightLogoURL\n          darkLogoURL\n        }\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "d0a4169ebdc6d099e49faa2839d6f33d";

export default node;
