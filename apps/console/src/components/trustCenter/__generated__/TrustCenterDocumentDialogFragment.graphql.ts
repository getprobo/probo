/**
 * @generated SignedSource<<3ff483dcc660e6b45d676db7fc05a30e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterDocumentDialogFragment$data = {
  readonly id: string;
  readonly showOnTrustCenter: boolean;
  readonly title: string;
  readonly " $fragmentType": "TrustCenterDocumentDialogFragment";
};
export type TrustCenterDocumentDialogFragment$key = {
  readonly " $data"?: TrustCenterDocumentDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterDocumentDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TrustCenterDocumentDialogFragment",
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
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "showOnTrustCenter",
      "storageKey": null
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "d886df25fdf197161a91bbc693c9fac8";

export default node;
