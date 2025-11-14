/**
 * @generated SignedSource<<aa77cda8d89dabe0fcfdcb82f2352efd>>
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
export type DocumentsPageRowFragment$data = {
  readonly classification: DocumentClassification;
  readonly description: string | null | undefined;
  readonly documentType: DocumentType;
  readonly id: string;
  readonly owner: {
    readonly fullName: string;
    readonly id: string;
  };
  readonly title: string;
  readonly updatedAt: any;
  readonly versions: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly signatures?: {
          readonly edges: ReadonlyArray<{
            readonly node: {
              readonly id: string;
              readonly state: DocumentVersionSignatureState;
            };
          }>;
        };
        readonly status: DocumentStatus;
        readonly version: number;
      };
    }>;
  };
  readonly " $fragmentType": "DocumentsPageRowFragment";
};
export type DocumentsPageRowFragment$key = {
  readonly " $data"?: DocumentsPageRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentsPageRowFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [
    {
      "defaultValue": false,
      "kind": "LocalArgument",
      "name": "includeSignatures"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentsPageRowFragment",
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
      "name": "description",
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
                  "condition": "includeSignatures",
                  "kind": "Condition",
                  "passingValue": true,
                  "selections": [
                    {
                      "alias": null,
                      "args": [
                        {
                          "kind": "Literal",
                          "name": "first",
                          "value": 1000
                        }
                      ],
                      "concreteType": "DocumentVersionSignatureConnection",
                      "kind": "LinkedField",
                      "name": "signatures",
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
                                }
                              ],
                              "storageKey": null
                            }
                          ],
                          "storageKey": null
                        }
                      ],
                      "storageKey": "signatures(first:1000)"
                    }
                  ]
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
  "type": "Document",
  "abstractKey": null
};
})();

(node as any).hash = "ae1f77bd2c1aca4b5d47af89d0407379";

export default node;
