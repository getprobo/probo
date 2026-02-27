/**
 * @generated SignedSource<<eaf60f297321410652305948a79c4240>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageBadgeListItemFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly iconUrl: string;
  readonly id: string;
  readonly name: string;
  readonly rank: number;
  readonly " $fragmentType": "CompliancePageBadgeListItemFragment";
};
export type CompliancePageBadgeListItemFragment$key = {
  readonly " $data"?: CompliancePageBadgeListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageBadgeListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageBadgeListItemFragment",
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
      "name": "iconUrl",
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
      "name": "rank",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:compliance-badge:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:compliance-badge:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:compliance-badge:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:compliance-badge:delete\")"
    }
  ],
  "type": "ComplianceBadge",
  "abstractKey": null
};

(node as any).hash = "bac36a0964b4938196dd7831a6b8406a";

export default node;
