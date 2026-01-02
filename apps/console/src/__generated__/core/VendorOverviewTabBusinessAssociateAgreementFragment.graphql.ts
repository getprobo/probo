/**
 * @generated SignedSource<<68b0a7b674474bde11fd886437848f4c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorOverviewTabBusinessAssociateAgreementFragment$data = {
  readonly businessAssociateAgreement: {
    readonly createdAt: string;
    readonly fileName: string;
    readonly fileUrl: string;
    readonly id: string;
    readonly validFrom: string | null | undefined;
    readonly validUntil: string | null | undefined;
  } | null | undefined;
  readonly " $fragmentType": "VendorOverviewTabBusinessAssociateAgreementFragment";
};
export type VendorOverviewTabBusinessAssociateAgreementFragment$key = {
  readonly " $data"?: VendorOverviewTabBusinessAssociateAgreementFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorOverviewTabBusinessAssociateAgreementFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorOverviewTabBusinessAssociateAgreementFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "VendorBusinessAssociateAgreement",
      "kind": "LinkedField",
      "name": "businessAssociateAgreement",
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
          "name": "fileName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fileUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "validFrom",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "validUntil",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};

(node as any).hash = "5f45b3392cbce2a58d3d68932ec68f88";

export default node;
