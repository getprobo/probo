/**
 * @generated SignedSource<<6052cc5cfc8407e61eaa1a842e976a90>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkControlDialogFragment$data = {
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly sectionTitle: string;
  readonly " $fragmentType": "FrameworkControlDialogFragment";
};
export type FrameworkControlDialogFragment$key = {
  readonly " $data"?: FrameworkControlDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworkControlDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworkControlDialogFragment",
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
      "name": "sectionTitle",
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};

(node as any).hash = "ff1a543d6a2225d9bacfe7db3b142f1a";

export default node;
