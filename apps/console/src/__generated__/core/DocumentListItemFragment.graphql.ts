/**
 * @generated SignedSource<<b6234102a57c0c233979c0bc51c2b49c>>
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
import { FragmentRefs } from "relay-runtime";
export type DocumentListItemFragment$data = {
  readonly canDelete: boolean;
  readonly classification: DocumentClassification;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly lastVersion: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly signatures: {
          readonly totalCount: number;
        };
        readonly signedSignatures: {
          readonly totalCount: number;
        };
        readonly status: DocumentStatus;
        readonly version: number;
      };
    }>;
  };
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly title: string;
  readonly updatedAt: string;
  readonly " $fragmentType": "DocumentListItemFragment";
};
export type DocumentListItemFragment$key = {
  readonly " $data"?: DocumentListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentListItemFragment">;
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
  "kind": "Literal",
  "name": "first",
  "value": 0
},
v2 = [
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "totalCount",
    "storageKey": null
  }
];
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentListItemFragment",
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
      "kind": "ScalarField",
      "name": "documentType",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "classification",
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
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document:delete\")"
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
      "alias": "lastVersion",
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 1
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
                (v0/*: any*/),
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
                  "name": "version",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "filter",
                      "value": {
                        "activeContract": true
                      }
                    },
                    (v1/*: any*/)
                  ],
                  "concreteType": "DocumentVersionSignatureConnection",
                  "kind": "LinkedField",
                  "name": "signatures",
                  "plural": false,
                  "selections": (v2/*: any*/),
                  "storageKey": "signatures(filter:{\"activeContract\":true},first:0)"
                },
                {
                  "alias": "signedSignatures",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "filter",
                      "value": {
                        "activeContract": true,
                        "states": [
                          "SIGNED"
                        ]
                      }
                    },
                    (v1/*: any*/)
                  ],
                  "concreteType": "DocumentVersionSignatureConnection",
                  "kind": "LinkedField",
                  "name": "signatures",
                  "plural": false,
                  "selections": (v2/*: any*/),
                  "storageKey": "signatures(filter:{\"activeContract\":true,\"states\":[\"SIGNED\"]},first:0)"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "versions(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
    }
  ],
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "6e72b37fa90b7a272ebba38fe80b9f93";

export default node;
