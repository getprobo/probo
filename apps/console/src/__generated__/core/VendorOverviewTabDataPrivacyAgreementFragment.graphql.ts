/**
 * @generated SignedSource<<03e78d5a57f41471ed7099febe61e666>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorOverviewTabDataPrivacyAgreementFragment$data = {
  readonly dataPrivacyAgreement: {
    readonly createdAt: any;
    readonly fileName: string;
    readonly fileUrl: string;
    readonly id: string;
    readonly validFrom: any | null | undefined;
    readonly validUntil: any | null | undefined;
  } | null | undefined;
  readonly " $fragmentType": "VendorOverviewTabDataPrivacyAgreementFragment";
};
export type VendorOverviewTabDataPrivacyAgreementFragment$key = {
  readonly " $data"?: VendorOverviewTabDataPrivacyAgreementFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorOverviewTabDataPrivacyAgreementFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorOverviewTabDataPrivacyAgreementFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "VendorDataPrivacyAgreement",
      "kind": "LinkedField",
      "name": "dataPrivacyAgreement",
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

(node as any).hash = "9a3e19e369a57ce4c1dd08760d244b74";

export default node;
