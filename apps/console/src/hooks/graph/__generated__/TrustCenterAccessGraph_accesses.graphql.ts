/**
 * @generated SignedSource<<3717625e10534aaff2d7967a1824405b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterAccessGraph_accesses$data = {
  readonly accesses: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly cursor: any;
      readonly node: {
        readonly active: boolean;
        readonly createdAt: any;
        readonly documentAccesses: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly active: boolean;
              readonly createdAt: any;
              readonly document: {
                readonly documentType: DocumentType;
                readonly id: string;
                readonly title: string;
              } | null | undefined;
              readonly id: string;
              readonly report: {
                readonly audit: {
                  readonly framework: {
                    readonly name: string;
                  };
                  readonly id: string;
                } | null | undefined;
                readonly filename: string;
                readonly id: string;
              } | null | undefined;
              readonly updatedAt: any;
            };
          }>;
        };
        readonly email: string;
        readonly hasAcceptedNonDisclosureAgreement: boolean;
        readonly id: string;
        readonly name: string;
      };
    }>;
    readonly pageInfo: {
      readonly endCursor: any | null | undefined;
      readonly hasNextPage: boolean;
      readonly hasPreviousPage: boolean;
      readonly startCursor: any | null | undefined;
    };
  };
  readonly id: string;
  readonly " $fragmentType": "TrustCenterAccessGraph_accesses";
};
export type TrustCenterAccessGraph_accesses$key = {
  readonly " $data"?: TrustCenterAccessGraph_accesses$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterAccessGraph_accesses">;
};

import TrustCenterAccessGraphPaginationQuery_graphql from './TrustCenterAccessGraphPaginationQuery.graphql';

const node: ReaderFragment = (function(){
var v0 = [
  "accesses"
],
v1 = {
  "kind": "Literal",
  "name": "orderBy",
  "value": {
    "direction": "DESC",
    "field": "CREATED_AT"
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
  "name": "active",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
};
return {
  "argumentDefinitions": [
    {
      "kind": "RootArgument",
      "name": "count"
    },
    {
      "kind": "RootArgument",
      "name": "cursor"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": "count",
        "cursor": "cursor",
        "direction": "forward",
        "path": (v0/*: any*/)
      }
    ],
    "refetch": {
      "connection": {
        "forward": {
          "count": "count",
          "cursor": "cursor"
        },
        "backward": null,
        "path": (v0/*: any*/)
      },
      "fragmentPathInResult": [
        "node"
      ],
      "operation": TrustCenterAccessGraphPaginationQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "TrustCenterAccessGraph_accesses",
  "selections": [
    {
      "alias": "accesses",
      "args": [
        (v1/*: any*/)
      ],
      "concreteType": "TrustCenterAccessConnection",
      "kind": "LinkedField",
      "name": "__TrustCenterAccessGraph_accesses_connection",
      "plural": false,
      "selections": [
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
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "endCursor",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenterAccessEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "cursor",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "TrustCenterAccess",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "email",
                  "storageKey": null
                },
                (v3/*: any*/),
                (v4/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "hasAcceptedNonDisclosureAgreement",
                  "storageKey": null
                },
                (v5/*: any*/),
                {
                  "alias": null,
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "first",
                      "value": 100
                    },
                    (v1/*: any*/)
                  ],
                  "concreteType": "TrustCenterDocumentAccessConnection",
                  "kind": "LinkedField",
                  "name": "documentAccesses",
                  "plural": false,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "TrustCenterDocumentAccessEdge",
                      "kind": "LinkedField",
                      "name": "edges",
                      "plural": true,
                      "selections": [
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "TrustCenterDocumentAccess",
                          "kind": "LinkedField",
                          "name": "node",
                          "plural": false,
                          "selections": [
                            (v2/*: any*/),
                            (v4/*: any*/),
                            (v5/*: any*/),
                            {
                              "alias": null,
                              "args": null,
                              "kind": "ScalarField",
                              "name": "updatedAt",
                              "storageKey": null
                            },
                            {
                              "alias": null,
                              "args": null,
                              "concreteType": "Document",
                              "kind": "LinkedField",
                              "name": "document",
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
                                {
                                  "alias": null,
                                  "args": null,
                                  "kind": "ScalarField",
                                  "name": "documentType",
                                  "storageKey": null
                                }
                              ],
                              "storageKey": null
                            },
                            {
                              "alias": null,
                              "args": null,
                              "concreteType": "Report",
                              "kind": "LinkedField",
                              "name": "report",
                              "plural": false,
                              "selections": [
                                (v2/*: any*/),
                                {
                                  "alias": null,
                                  "args": null,
                                  "kind": "ScalarField",
                                  "name": "filename",
                                  "storageKey": null
                                },
                                {
                                  "alias": null,
                                  "args": null,
                                  "concreteType": "Audit",
                                  "kind": "LinkedField",
                                  "name": "audit",
                                  "plural": false,
                                  "selections": [
                                    (v2/*: any*/),
                                    {
                                      "alias": null,
                                      "args": null,
                                      "concreteType": "Framework",
                                      "kind": "LinkedField",
                                      "name": "framework",
                                      "plural": false,
                                      "selections": [
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
                          "storageKey": null
                        }
                      ],
                      "storageKey": null
                    }
                  ],
                  "storageKey": "documentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "__typename",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        },
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
      "storageKey": "__TrustCenterAccessGraph_accesses_connection(orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
    },
    (v2/*: any*/)
  ],
  "type": "TrustCenter",
  "abstractKey": null
};
})();

(node as any).hash = "9e29aa4a382ca2d4bb9d1b014d8d81fd";

export default node;
