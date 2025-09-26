/**
 * @generated SignedSource<<4c861a832ad2f79bd9d6ddcc364f6898>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorComplianceTabFragment_report$data = {
  readonly fileSize: any;
  readonly fileUrl: string;
  readonly id: string;
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
      "name": "id",
      "storageKey": null
    },
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

(node as any).hash = "92ad78c9e4b4add9b1f622485d81a6ff";

export default node;
