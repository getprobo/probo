/**
 * @generated SignedSource<<10302f020070a7ab64ab93115a4b916b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturePlaceholder_personFragment$data = {
  readonly emailAddress: string;
  readonly fullName: string;
  readonly id: string;
  readonly " $fragmentType": "DocumentSignaturePlaceholder_personFragment";
};
export type DocumentSignaturePlaceholder_personFragment$key = {
  readonly " $data"?: DocumentSignaturePlaceholder_personFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_personFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturePlaceholder_personFragment",
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
      "name": "fullName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "emailAddress",
      "storageKey": null
    }
  ],
  "type": "Profile",
  "abstractKey": null
};

(node as any).hash = "cbbb481145654dbf42e18ca46ec040fb";

export default node;
