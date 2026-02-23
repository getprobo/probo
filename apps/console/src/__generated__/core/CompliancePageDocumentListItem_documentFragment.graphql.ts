/**
 * @generated SignedSource<<85da952604018b5f879924bf3f8d5a27>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDocumentListItem_documentFragment$data = {
  readonly documentType: DocumentType;
  readonly id: string;
  readonly latestPublishedVersion: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly title: string;
      };
    }>;
  };
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly " $fragmentType": "CompliancePageDocumentListItem_documentFragment";
};
export type CompliancePageDocumentListItem_documentFragment$key = {
  readonly " $data"?: CompliancePageDocumentListItem_documentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListItem_documentFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDocumentListItem_documentFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
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
      "name": "trustCenterVisibility",
      "storageKey": null
    },
    {
      "alias": "latestPublishedVersion",
      "args": [
        {
          "kind": "Literal",
          "name": "filter",
          "value": {
            "status": "PUBLISHED"
          }
        },
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
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "title",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "versions(filter:{\"status\":\"PUBLISHED\"},first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "4c6d4e935f4a54b6d2ef486f60aab5f6";

export default node;
