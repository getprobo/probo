/**
 * @generated SignedSource<<01d2d791b86ae99330d4ba2ecb01f1ed>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorComplianceTabFragment_report$data = {
  readonly fileSize: number;
  readonly fileUrl: string;
  readonly reportDate: any;
  readonly reportName: string;
  readonly validUntil: any | null | undefined;
  readonly " $fragmentType": "VendorComplianceTabFragment_report";
};
export type VendorComplianceTabFragment_report$key = {
  readonly " $data"?: VendorComplianceTabFragment_report$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorComplianceTabFragment_report">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorComplianceTabFragment_report",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "reportDate",
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
      "name": "reportName",
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
      "name": "fileSize",
      "storageKey": null
    }
  ],
  "type": "VendorComplianceReport",
  "abstractKey": null
};

(node as any).hash = "d39b70dfd0a316b158d7dc7253d0ff5e";

export default node;
