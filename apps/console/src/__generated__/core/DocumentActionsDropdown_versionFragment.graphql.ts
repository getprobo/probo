/**
 * @generated SignedSource<<ab58ae3e6f4142422681638a391a5dc4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentActionsDropdown_versionFragment$data = {
  readonly canDeleteDraft: boolean;
  readonly id: string;
  readonly status: DocumentStatus;
  readonly version: number;
  readonly " $fragmentType": "DocumentActionsDropdown_versionFragment";
};
export type DocumentActionsDropdown_versionFragment$key = {
  readonly " $data"?: DocumentActionsDropdown_versionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_versionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentActionsDropdown_versionFragment",
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
      "name": "version",
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
      "alias": "canDeleteDraft",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document-version:delete-draft"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document-version:delete-draft\")"
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "00619a8709984275b2e18c2c44dd4e36";

export default node;
