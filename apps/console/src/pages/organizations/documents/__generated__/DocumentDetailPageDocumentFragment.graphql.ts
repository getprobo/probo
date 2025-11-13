/**
 * @generated SignedSource<<d5c96a7bfed3f3152aff6729ab9f5ee2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type DocumentDetailPageDocumentFragment$data = {
  readonly classification: DocumentClassification;
  readonly controlsInfo?: {
    readonly totalCount: number;
  };
  readonly documentType: DocumentType;
  readonly id: string;
  readonly organization: {
    readonly id: string;
  };
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly requestedVersions?: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly classification: DocumentClassification;
        readonly content: string;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly publishedAt: any | null | undefined;
        readonly signatures?: {
          readonly __id: string;
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: DocumentVersionSignatureState;
            };
          }>;
        };
        readonly status: DocumentStatus;
        readonly updatedAt: any;
        readonly version: number;
      };
    }>;
  };
  readonly title: string;
  readonly versions?: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly classification: DocumentClassification;
        readonly content: string;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly publishedAt: any | null | undefined;
        readonly signatures?: {
          readonly __id: string;
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: DocumentVersionSignatureState;
            };
          }>;
        };
        readonly status: DocumentStatus;
        readonly updatedAt: any;
        readonly version: number;
      };
    }>;
  };
  readonly " $fragmentType": "DocumentDetailPageDocumentFragment";
};
export type DocumentDetailPageDocumentFragment$key = {
  readonly " $data"?: DocumentDetailPageDocumentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentDetailPageDocumentFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "count": null,
  "cursor": null,
  "direction": "forward",
  "path": null
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
  "name": "classification",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v1/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fullName",
      "storageKey": null
    }
  ],
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
v7 = {
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
v8 = [
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
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "content",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "status",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "publishedAt",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "version",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
            "storageKey": null
          },
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "condition": "includeSignatures",
            "kind": "Condition",
            "passingValue": true,
            "selections": [
              {
                "alias": "signatures",
                "args": null,
                "concreteType": "DocumentVersionSignatureConnection",
                "kind": "LinkedField",
                "name": "__DocumentDetailPage_signatures_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "DocumentVersionSignatureEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "DocumentVersionSignature",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v1/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "state",
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
                  (v6/*: any*/),
                  (v7/*: any*/)
                ],
                "storageKey": null
              }
            ]
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      },
      (v5/*: any*/)
    ],
    "storageKey": null
  },
  (v6/*: any*/),
  (v7/*: any*/)
];
return {
  "argumentDefinitions": [
    {
      "defaultValue": false,
      "kind": "LocalArgument",
      "name": "includeControls"
    },
    {
      "defaultValue": false,
      "kind": "LocalArgument",
      "name": "includeSignatures"
    },
    {
      "defaultValue": false,
      "kind": "LocalArgument",
      "name": "useRequestedVersions"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      (v0/*: any*/),
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "versions"
        ]
      },
      (v0/*: any*/),
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "requestedVersions"
        ]
      }
    ]
  },
  "name": "DocumentDetailPageDocumentFragment",
  "selections": [
    (v1/*: any*/),
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
        (v1/*: any*/)
      ],
      "storageKey": null
    },
    (v3/*: any*/),
    {
      "condition": "includeControls",
      "kind": "Condition",
      "passingValue": true,
      "selections": [
        {
          "alias": "controlsInfo",
          "args": [
            {
              "kind": "Literal",
              "name": "first",
              "value": 0
            }
          ],
          "concreteType": "ControlConnection",
          "kind": "LinkedField",
          "name": "controls",
          "plural": false,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "totalCount",
              "storageKey": null
            }
          ],
          "storageKey": "controls(first:0)"
        }
      ]
    },
    {
      "condition": "useRequestedVersions",
      "kind": "Condition",
      "passingValue": false,
      "selections": [
        {
          "alias": "versions",
          "args": null,
          "concreteType": "DocumentVersionConnection",
          "kind": "LinkedField",
          "name": "__DocumentDetailPage_versions_connection",
          "plural": false,
          "selections": (v8/*: any*/),
          "storageKey": null
        }
      ]
    },
    {
      "condition": "useRequestedVersions",
      "kind": "Condition",
      "passingValue": true,
      "selections": [
        {
          "alias": "requestedVersions",
          "args": null,
          "concreteType": "DocumentVersionConnection",
          "kind": "LinkedField",
          "name": "__DocumentDetailPage_requestedVersions_connection",
          "plural": false,
          "selections": (v8/*: any*/),
          "storageKey": null
        }
      ]
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "6eb338aa9c19eb32c369fe11a7878740";

export default node;
