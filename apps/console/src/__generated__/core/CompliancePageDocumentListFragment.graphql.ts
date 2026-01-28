/**
 * @generated SignedSource<<520e633d06f7c68a9270600108ec81da>>
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
          "value": 1000
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
      "storageKey": "documents(first:1000)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "10da8f0838c700d4a2535ec5193844a8";

export default node;
