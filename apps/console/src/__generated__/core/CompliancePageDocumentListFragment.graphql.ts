/**
 * @generated SignedSource<<2a0f1332c57638e29eeee8e4113858d3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDocumentListFragment$data = {
  readonly compliancePage: {
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListItem_compliancePageFragment">;
  };
  readonly documents: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly currentPublishedVersion: number | null | undefined;
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListItem_documentFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "CompliancePageDocumentListFragment";
};
export type CompliancePageDocumentListFragment$key = {
  readonly " $data"?: CompliancePageDocumentListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDocumentListFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDocumentListFragment",
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
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageDocumentListItem_compliancePageFragment"
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        }
      ],
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
                  "name": "currentPublishedVersion",
                  "storageKey": null
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "CompliancePageDocumentListItem_documentFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "documents(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "48e4f01fb5d80ae0856433598608d4cf";

export default node;
