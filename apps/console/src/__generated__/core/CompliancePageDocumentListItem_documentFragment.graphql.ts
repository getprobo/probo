/**
 * @generated SignedSource<<112e917e3fb881859a56b794f07ce7e7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDocumentListItem_documentFragment$data = {
  readonly documentType: DocumentType;
  readonly id: string;
  readonly lastVersion: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly status: DocumentStatus;
      };
    }>;
  };
  readonly title: string;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly " $fragmentType": "CompliancePageDocumentListItem_documentFragment";
};
export type CompliancePageDocumentListItem_documentFragment$key = {
  readonly " $data"?: CompliancePageDocumentListItem_documentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListItem_documentFragment">;
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
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDocumentListItem_documentFragment",
  "selections": [
    (v0/*: any*/),
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
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "trustCenterVisibility",
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

(node as any).hash = "ab2b10ffbd76c59d1a3facab0f794402";

export default node;
