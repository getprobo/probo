/**
 * @generated SignedSource<<18ee21b1b07c34808fb34640fd0eee77>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterGraphQuery$variables = {
  organizationId: string;
};
export type TrustCenterGraphQuery$data = {
  readonly organization: {
    readonly audits?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterAuditsCardFragment">;
        };
      }>;
    };
    readonly canCreateTrustCenterFile?: boolean;
    readonly customDomain?: {
      readonly domain: string;
      readonly id: string;
    } | null | undefined;
    readonly documents?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterDocumentsCardFragment">;
        };
      }>;
    };
    readonly id?: string;
    readonly name?: string;
    readonly slackConnections?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly channel: string | null | undefined;
          readonly channelId: string | null | undefined;
          readonly createdAt: any;
          readonly id: string;
          readonly updatedAt: any;
        };
      }>;
    };
    readonly trustCenter?: {
      readonly active: boolean;
      readonly canCreateReference: boolean;
      readonly canDeleteNDA: boolean;
      readonly canGetNDA: boolean;
      readonly canUpdate: boolean;
      readonly canUploadNDA: boolean;
      readonly createdAt: any;
      readonly id: string;
      readonly ndaFileName: string | null | undefined;
      readonly ndaFileUrl: string | null | undefined;
      readonly updatedAt: any;
    } | null | undefined;
    readonly trustCenterFiles?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFilesCardFragment">;
        };
      }>;
    };
    readonly vendors?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterVendorsCardFragment">;
        };
      }>;
    };
  };
};
export type TrustCenterGraphQuery = {
  response: TrustCenterGraphQuery$data;
  variables: TrustCenterGraphQuery$variables;
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
  "alias": "canCreateTrustCenterFile",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center-file:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center-file:create\")"
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "CustomDomain",
  "kind": "LinkedField",
  "name": "customDomain",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "domain",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "TrustCenter",
  "kind": "LinkedField",
  "name": "trustCenter",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "active",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "ndaFileName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "ndaFileUrl",
      "storageKey": null
    },
    (v6/*: any*/),
    (v7/*: any*/),
    {
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
    {
      "alias": "canGetNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:get-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:get-nda\")"
    },
    {
      "alias": "canUploadNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:upload-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:upload-nda\")"
    },
    {
      "alias": "canDeleteNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:delete-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:delete-nda\")"
    },
    {
      "alias": "canCreateReference",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-reference:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-reference:create\")"
    }
  ],
  "storageKey": null
},
v9 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
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
v14 = {
  "alias": null,
  "args": (v9/*: any*/),
  "concreteType": "SlackConnectionConnection",
  "kind": "LinkedField",
  "name": "slackConnections",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "SlackConnectionEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "SlackConnection",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "channel",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "channelId",
              "storageKey": null
            },
            (v6/*: any*/),
            (v7/*: any*/)
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "slackConnections(first:100)"
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "trustCenterVisibility",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "category",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterGraphQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Document",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "TrustCenterDocumentsCardFragment"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "AuditConnection",
                "kind": "LinkedField",
                "name": "audits",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AuditEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Audit",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "TrustCenterAuditsCardFragment"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "audits(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "VendorConnection",
                "kind": "LinkedField",
                "name": "vendors",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Vendor",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "TrustCenterVendorsCardFragment"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "vendors(first:100)"
              },
              {
                "alias": "trustCenterFiles",
                "args": null,
                "concreteType": "TrustCenterFileConnection",
                "kind": "LinkedField",
                "name": "__TrustCenterPage_trustCenterFiles_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterFileEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterFile",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "TrustCenterFilesCardFragment"
                          },
                          (v10/*: any*/)
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
                "storageKey": null
              },
              (v14/*: any*/)
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
    "name": "TrustCenterGraphQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v10/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "DocumentConnection",
                "kind": "LinkedField",
                "name": "documents",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Document",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "title",
                            "storageKey": null
                          },
                          (v6/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "documentType",
                            "storageKey": null
                          },
                          (v15/*: any*/),
                          {
                            "alias": null,
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "first",
                                "value": 1
                              }
                            ],
                            "concreteType": "DocumentVersionConnection",
                            "kind": "LinkedField",
                            "name": "versions",
                            "plural": false,
                            "selections": [
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "DocumentVersionEdge",
                                "kind": "LinkedField",
                                "name": "edges",
                                "plural": true,
                                "selections": [
                                  {
                                    "alias": null,
                                    "args": null,
                                    "concreteType": "DocumentVersion",
                                    "kind": "LinkedField",
                                    "name": "node",
                                    "plural": false,
                                    "selections": [
                                      (v2/*: any*/),
                                      {
                                        "alias": null,
                                        "args": null,
                                        "kind": "ScalarField",
                                        "name": "status",
                                        "storageKey": null
                                      }
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": "versions(first:1)"
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "documents(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "AuditConnection",
                "kind": "LinkedField",
                "name": "audits",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "AuditEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Audit",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Framework",
                            "kind": "LinkedField",
                            "name": "framework",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "validFrom",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "validUntil",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "state",
                            "storageKey": null
                          },
                          (v15/*: any*/),
                          (v6/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "audits(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "VendorConnection",
                "kind": "LinkedField",
                "name": "vendors",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Vendor",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v16/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "description",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "showOnTrustCenter",
                            "storageKey": null
                          },
                          (v6/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "vendors(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "concreteType": "TrustCenterFileConnection",
                "kind": "LinkedField",
                "name": "trustCenterFiles",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterFileEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterFile",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v16/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "fileUrl",
                            "storageKey": null
                          },
                          (v15/*: any*/),
                          (v6/*: any*/),
                          (v7/*: any*/),
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:trust-center-file:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:trust-center-file:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:trust-center-file:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:trust-center-file:delete\")"
                          },
                          (v10/*: any*/)
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
                "storageKey": "trustCenterFiles(first:100)"
              },
              {
                "alias": null,
                "args": (v9/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "TrustCenterPage_trustCenterFiles",
                "kind": "LinkedHandle",
                "name": "trustCenterFiles"
              },
              (v14/*: any*/)
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
    "cacheID": "6af8bc39a3654fa715f0f235bcb5f488",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "organization",
            "trustCenterFiles"
          ]
        }
      ]
    },
    "name": "TrustCenterGraphQuery",
    "operationKind": "query",
    "text": "query TrustCenterGraphQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      name\n      canCreateTrustCenterFile: permission(action: \"core:trust-center-file:create\")\n      customDomain {\n        id\n        domain\n      }\n      trustCenter {\n        id\n        active\n        ndaFileName\n        ndaFileUrl\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:trust-center:update\")\n        canGetNDA: permission(action: \"core:trust-center:get-nda\")\n        canUploadNDA: permission(action: \"core:trust-center:upload-nda\")\n        canDeleteNDA: permission(action: \"core:trust-center:delete-nda\")\n        canCreateReference: permission(action: \"core:trust-center-reference:create\")\n      }\n      documents(first: 100) {\n        edges {\n          node {\n            id\n            ...TrustCenterDocumentsCardFragment\n          }\n        }\n      }\n      audits(first: 100) {\n        edges {\n          node {\n            id\n            ...TrustCenterAuditsCardFragment\n          }\n        }\n      }\n      vendors(first: 100) {\n        edges {\n          node {\n            id\n            ...TrustCenterVendorsCardFragment\n          }\n        }\n      }\n      trustCenterFiles(first: 100) {\n        edges {\n          node {\n            id\n            ...TrustCenterFilesCardFragment\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n      slackConnections(first: 100) {\n        edges {\n          node {\n            id\n            channel\n            channelId\n            createdAt\n            updatedAt\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment TrustCenterAuditsCardFragment on Audit {\n  id\n  name\n  framework {\n    name\n    id\n  }\n  validFrom\n  validUntil\n  state\n  trustCenterVisibility\n  createdAt\n}\n\nfragment TrustCenterDocumentsCardFragment on Document {\n  id\n  title\n  createdAt\n  documentType\n  trustCenterVisibility\n  versions(first: 1) {\n    edges {\n      node {\n        id\n        status\n      }\n    }\n  }\n}\n\nfragment TrustCenterFilesCardFragment on TrustCenterFile {\n  id\n  name\n  category\n  fileUrl\n  trustCenterVisibility\n  createdAt\n  updatedAt\n  canUpdate: permission(action: \"core:trust-center-file:update\")\n  canDelete: permission(action: \"core:trust-center-file:delete\")\n}\n\nfragment TrustCenterVendorsCardFragment on Vendor {\n  id\n  name\n  category\n  description\n  showOnTrustCenter\n  createdAt\n}\n"
  }
};
})();

(node as any).hash = "9669200b7f9e58c412b73b30fa506f51";

export default node;
