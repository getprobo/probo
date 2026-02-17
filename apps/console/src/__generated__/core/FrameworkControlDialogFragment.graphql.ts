/**
 * @generated SignedSource<<43181b3f35ce7df5f54b19bf82eeaad0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworkControlDialogFragment$data = {
  readonly bestPractice: boolean;
  readonly description: string | null | undefined;
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
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "bestPractice",
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};

(node as any).hash = "3ca9ebd60cc24f1080a594bf4ca0b928";

export default node;
