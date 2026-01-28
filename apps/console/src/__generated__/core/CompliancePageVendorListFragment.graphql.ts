/**
 * @generated SignedSource<<b8031a4cc01354ec19a8dbce53ae3162>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageVendorListFragment$data = {
  readonly vendors: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"CompliancePageVendorListItem_vendorFragment">;
      };
    }>;
  };
  readonly " $fragmentType": "CompliancePageVendorListFragment";
};
export type CompliancePageVendorListFragment$key = {
  readonly " $data"?: CompliancePageVendorListFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageVendorListFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageVendorListFragment",
  "selections": [
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        }
      ],
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
                  "name": "CompliancePageVendorListItem_vendorFragment"
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "vendors(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "dc9e2540ae0f8b56dc322ef6cbbfe805";

export default node;
