/**
 * @generated SignedSource<<d797330ab78ac421e3c64cebdb2ec975>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworksPageCardFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
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
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:framework:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:framework:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:framework:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:framework:delete\")"
    }
  ],
  "type": "Framework",
  "abstractKey": null
};

(node as any).hash = "dfb64192cedec377c8df8a6d8251e815";

export default node;
