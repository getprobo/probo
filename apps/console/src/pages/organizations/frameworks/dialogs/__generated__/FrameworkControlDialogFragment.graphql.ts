/**
 * @generated SignedSource<<3a62e5f931a990d83caca02ab629b0c0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ControlStatus = "EXCLUDED" | "INCLUDED";
import { FragmentRefs } from "relay-runtime";
export type FrameworkControlDialogFragment$data = {
  readonly bestPractice: boolean;
  readonly description: string | null | undefined;
  readonly exclusionJustification: string | null | undefined;
  readonly id: string;
  readonly name: string;
  readonly sectionTitle: string;
  readonly status: ControlStatus;
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
      "name": "status",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "exclusionJustification",
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

(node as any).hash = "481f75ce8c525c1fa12517ef173ad868";

export default node;
