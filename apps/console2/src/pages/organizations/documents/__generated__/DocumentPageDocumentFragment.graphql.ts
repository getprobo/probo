/**
 * @generated SignedSource<<c182ef1b9491681644676e99e0b07ee7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentVersionSignatureState = "REQUESTED" | "SIGNED";
import { FragmentRefs } from "relay-runtime";
export type DocumentPageDocumentFragment$data = {
  readonly id: string;
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly title: string;
  readonly versions: {
    readonly __id: string;
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly content: string;
        readonly id: string;
        readonly publishedAt: any | null | undefined;
        readonly signatures: {
          readonly __id: string;
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly signedBy: {
                readonly id: string;
              };
              readonly state: DocumentVersionSignatureState;
              readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesDialog_signature">;
            };
          }>;
        };
        readonly status: DocumentStatus;
        readonly updatedAt: any;
        readonly version: number;
        readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesDialog_version" | "DocumentVersionHistoryDialogFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "DocumentPageDocumentFragment";
};
export type DocumentPageDocumentFragment$key = {
  readonly " $data"?: DocumentPageDocumentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentPageDocumentFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v3 = {
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
v4 = {
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
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "connection": [
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": null
      },
      {
        "count": null,
        "cursor": null,
        "direction": "forward",
        "path": [
          "versions"
        ]
      }
    ]
  },
  "name": "DocumentPageDocumentFragment",
  "selections": [
    (v0/*: any*/),
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
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "owner",
      "plural": false,
      "selections": [
        (v0/*: any*/),
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
    {
      "alias": "versions",
      "args": null,
      "concreteType": "DocumentVersionConnection",
      "kind": "LinkedField",
      "name": "__DocumentPage_versions_connection",
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
                (v0/*: any*/),
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
                {
                  "alias": "signatures",
                  "args": null,
                  "concreteType": "DocumentVersionSignatureConnection",
                  "kind": "LinkedField",
                  "name": "__DocumentPage_signatures_connection",
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
                            (v0/*: any*/),
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
                              "concreteType": "People",
                              "kind": "LinkedField",
                              "name": "signedBy",
                              "plural": false,
                              "selections": [
                                (v0/*: any*/)
                              ],
                              "storageKey": null
                            },
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "DocumentSignaturesDialog_signature"
                            },
                            (v1/*: any*/)
                          ],
                          "storageKey": null
                        },
                        (v2/*: any*/)
                      ],
                      "storageKey": null
                    },
                    (v3/*: any*/),
                    (v4/*: any*/)
                  ],
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "DocumentVersionHistoryDialogFragment"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "DocumentSignaturesDialog_version"
                },
                (v1/*: any*/)
              ],
              "storageKey": null
            },
            (v2/*: any*/)
          ],
          "storageKey": null
        },
        (v3/*: any*/),
        (v4/*: any*/)
      ],
      "storageKey": null
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "b42e98e8e26bc94a2d412bd73cd601f2";

export default node;
