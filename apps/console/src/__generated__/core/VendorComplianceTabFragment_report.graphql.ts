/**
 * @generated SignedSource<<af8474eed98f58b7c1f877660e4f9545>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorComplianceTabFragment_report$data = {
  readonly canDelete: boolean;
  readonly file: {
    readonly downloadUrl: string;
    readonly fileName: string;
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
          "name": "size",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "downloadUrl",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor-compliance-report:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor-compliance-report:delete\")"
    }
  ],
  "type": "VendorComplianceReport",
  "abstractKey": null
};

(node as any).hash = "cbf9fe2b6540de802f43eb7f840a24a7";

export default node;
