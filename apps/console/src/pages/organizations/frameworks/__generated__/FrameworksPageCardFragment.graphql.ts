/**
 * @generated SignedSource<<412cc494e90709fca575c68f6deef112>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworksPageCardFragment$data = {
  readonly darkLogoURL: string | null | undefined;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly lightLogoURL: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "FrameworksPageCardFragment";
};
export type FrameworksPageCardFragment$key = {
  readonly " $data"?: FrameworksPageCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworksPageCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworksPageCardFragment",
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
      "name": "description",
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

(node as any).hash = "fe86a14971741a77f587ae015ebf0a22";

export default node;
