/**
 * @generated SignedSource<<3cf0d5be883754321b4278ac59d77a54>>
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
    readonly canDelete: boolean;
    readonly canUpdate: boolean;
    readonly createdAt: string;
    readonly fileName: string;
    readonly fileUrl: string;
    readonly id: string;
    readonly validFrom: string | null | undefined;
    readonly validUntil: string | null | undefined;
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
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:vendor-data-privacy-agreement:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:vendor-data-privacy-agreement:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:vendor-data-privacy-agreement:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:vendor-data-privacy-agreement:delete\")"
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};

(node as any).hash = "c07606a63a9dc76f5cfad938c9d54268";

export default node;
