/**
 * @generated SignedSource<<f13ba91513aa9a472ab62c4c6a93b6bb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkBadgeFragment$data = {
  readonly darkLogoURL: string | null | undefined;
  readonly id: string;
  readonly lightLogoURL: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "FrameworkBadgeFragment";
};
export type FrameworkBadgeFragment$key = {
  readonly " $data"?: FrameworkBadgeFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworkBadgeFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworkBadgeFragment",
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
      "name": "lightLogoURL",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "darkLogoURL",
      "storageKey": null
    }
  ],
  "type": "Framework",
  "abstractKey": null
};

(node as any).hash = "cdfaf9a4a0e2cb667a5ef8d3e7db1f99";

export default node;
