/**
 * @generated SignedSource<<e765fe663fce16d50260a876b9ab0752>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageReferenceListItemFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly logoUrl: string;
  readonly name: string;
  readonly rank: number;
  readonly websiteUrl: string;
  readonly " $fragmentType": "CompliancePageReferenceListItemFragment";
};
export type CompliancePageReferenceListItemFragment$key = {
  readonly " $data"?: CompliancePageReferenceListItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageReferenceListItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageReferenceListItemFragment",
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
      "name": "logoUrl",
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
      "name": "websiteUrl",
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
          "value": "core:trust-center-reference:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-reference:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-reference:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-reference:delete\")"
    }
  ],
  "type": "TrustCenterReference",
  "abstractKey": null
};

(node as any).hash = "c7b132a9fea8a7c8cbf1a29af14dbb76";

export default node;
