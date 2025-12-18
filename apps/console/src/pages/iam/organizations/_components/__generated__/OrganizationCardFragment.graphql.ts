/**
 * @generated SignedSource<<06aa55abfa21c09c85bb308f813565d8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationCardFragment$data = {
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly " $fragmentType": "OrganizationCardFragment";
};
export type OrganizationCardFragment$key = {
  readonly " $data"?: OrganizationCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationCardFragment",
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
      "name": "logoUrl",
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "a8b9f4a515650db79f41507fb52f7b96";

export default node;
