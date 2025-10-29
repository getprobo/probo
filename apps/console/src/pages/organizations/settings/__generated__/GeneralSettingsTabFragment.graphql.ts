/**
 * @generated SignedSource<<02beba57812b8dd7c5c61fb10291de5a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type GeneralSettingsTabFragment$data = {
  readonly createdAt: any;
  readonly description: string | null | undefined;
  readonly email: string | null | undefined;
  readonly headquarterAddress: string | null | undefined;
  readonly horizontalLogoUrl: string | null | undefined;
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly updatedAt: any;
  readonly websiteUrl: string | null | undefined;
  readonly " $fragmentType": "GeneralSettingsTabFragment";
};
export type GeneralSettingsTabFragment$key = {
  readonly " $data"?: GeneralSettingsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"GeneralSettingsTabFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "GeneralSettingsTabFragment",
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
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "horizontalLogoUrl",
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
      "name": "websiteUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "email",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "headquarterAddress",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "updatedAt",
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "f6552148ce1c0061c4f5cbef78d46c0e";

export default node;
