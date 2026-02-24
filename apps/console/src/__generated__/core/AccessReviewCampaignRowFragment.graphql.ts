/**
 * @generated SignedSource<<186c245cc9321c653836b1527de5d665>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type AccessReviewCampaignStatus = "CANCELLED" | "COMPLETED" | "DRAFT" | "FAILED" | "IN_PROGRESS" | "PENDING_ACTIONS";
import { FragmentRefs } from "relay-runtime";
export type AccessReviewCampaignRowFragment$data = {
  readonly canDelete: boolean;
  readonly createdAt: string;
  readonly id: string;
  readonly name: string;
  readonly status: AccessReviewCampaignStatus;
  readonly " $fragmentType": "AccessReviewCampaignRowFragment";
};
export type AccessReviewCampaignRowFragment$key = {
  readonly " $data"?: AccessReviewCampaignRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"AccessReviewCampaignRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "AccessReviewCampaignRowFragment",
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
      "name": "status",
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
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:access-review-campaign:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:access-review-campaign:delete\")"
    }
  ],
  "type": "AccessReviewCampaign",
  "abstractKey": null
};

(node as any).hash = "464ac2cdb0e4f9f1d133c3e3577ab63a";

export default node;
