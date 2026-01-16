/**
 * @generated SignedSource<<f777a3e905bf0a68614d31011688f76d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VersionActionsFragment$data = {
  readonly id: string;
  readonly signed: boolean;
  readonly " $fragmentType": "VersionActionsFragment";
};
export type VersionActionsFragment$key = {
  readonly " $data"?: VersionActionsFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VersionActionsFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VersionActionsFragment",
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
      "name": "signed",
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "9d3f7cc67712ad429dc0b1072fe2fe97";

export default node;
