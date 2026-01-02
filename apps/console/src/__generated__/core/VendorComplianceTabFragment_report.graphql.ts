/**
 * @generated SignedSource<<dc7aecc55aa17f43d9a601e60eea165a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorComplianceTabFragment_report$data = {
  readonly file: {
    readonly fileName: string;
    readonly mimeType: string;
    readonly size: number;
  } | null | undefined;
  readonly id: string;
  readonly reportDate: string;
  readonly reportName: string;
  readonly validUntil: string | null | undefined;
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
      "concreteType": "File",
      "kind": "LinkedField",
      "name": "file",
      "plural": false,
      "selections": [
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
          "name": "mimeType",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "size",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "VendorComplianceReport",
  "abstractKey": null
};

(node as any).hash = "18f52de61577c8b1c61784d2404c24ae";

export default node;
