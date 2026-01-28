/**
 * @generated SignedSource<<5acd0b499b718393c592fa3d05cf3157>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageVendorListItem_vendorFragment$data = {
  readonly canUpdate: boolean;
  readonly category: VendorCategory;
  readonly id: string;
  readonly name: string;
  readonly showOnTrustCenter: boolean;
  readonly " $fragmentType": "CompliancePageVendorListItem_vendorFragment";
};
export type CompliancePageVendorListItem_vendorFragment$key = {
  readonly " $data"?: CompliancePageVendorListItem_vendorFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageVendorListItem_vendorFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageVendorListItem_vendorFragment",
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
      "name": "category",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "showOnTrustCenter",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor:update\")"
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};

(node as any).hash = "147628fced6667323cce1bf7e81fabfb";

export default node;
