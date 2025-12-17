/**
 * @generated SignedSource<<1b79ba90084eae4139b4fd4a31e5061b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SettingsPageFragment$data = {
  readonly id: string;
  readonly name: string;
  readonly " $fragmentSpreads": FragmentRefs<"DomainSettingsTabFragment">;
  readonly " $fragmentType": "SettingsPageFragment";
};
export type SettingsPageFragment$key = {
  readonly " $data"?: SettingsPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SettingsPageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SettingsPageFragment",
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
      "args": null,
      "kind": "FragmentSpread",
      "name": "DomainSettingsTabFragment"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "c00c2edf8bd9f8255c6bd6943bd4d445";

export default node;
