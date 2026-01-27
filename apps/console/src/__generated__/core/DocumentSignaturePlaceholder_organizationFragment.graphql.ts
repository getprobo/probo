/**
 * @generated SignedSource<<d2c766afc66acdd05219397c4bfd1b44>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturePlaceholder_organizationFragment$data = {
  readonly canRequestSignature: boolean;
  readonly " $fragmentType": "DocumentSignaturePlaceholder_organizationFragment";
};
export type DocumentSignaturePlaceholder_organizationFragment$key = {
  readonly " $data"?: DocumentSignaturePlaceholder_organizationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_organizationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturePlaceholder_organizationFragment",
  "selections": [
    {
      "alias": "canRequestSignature",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document-version:request-signature"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document-version:request-signature\")"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "4bb0c71b4b90f335c9938b0283639f36";

export default node;
