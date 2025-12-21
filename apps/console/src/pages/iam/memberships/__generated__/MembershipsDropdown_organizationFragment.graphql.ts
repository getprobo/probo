/**
 * @generated SignedSource<<84862118e94833cec2ce241e7329ba04>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipsDropdown_organizationFragment$data = {
  readonly name: string;
  readonly " $fragmentType": "MembershipsDropdown_organizationFragment";
};
export type MembershipsDropdown_organizationFragment$key = {
  readonly " $data"?: MembershipsDropdown_organizationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_organizationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipsDropdown_organizationFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "89ef8f037f10c51776df8e5add87cc29";

export default node;
