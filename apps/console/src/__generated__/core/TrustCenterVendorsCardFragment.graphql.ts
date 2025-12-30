/**
 * @generated SignedSource<<b3d65d827fc901aa002b24df72f79ee2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterVendorsCardFragment$data = {
  readonly category: VendorCategory;
  readonly createdAt: any;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly name: string;
  readonly showOnTrustCenter: boolean;
  readonly " $fragmentType": "TrustCenterVendorsCardFragment";
};
export type TrustCenterVendorsCardFragment$key = {
  readonly " $data"?: TrustCenterVendorsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterVendorsCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TrustCenterVendorsCardFragment",
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
      "name": "name",
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
      "name": "description",
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};

(node as any).hash = "72b637589fd9299b5e982450fccbfd12";

export default node;
