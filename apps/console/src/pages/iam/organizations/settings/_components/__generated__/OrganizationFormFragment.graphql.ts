/**
 * @generated SignedSource<<8deb33fbc8305caa3d3de954cf555c79>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationFormFragment$data = {
  readonly description: string | null | undefined;
  readonly email: string | null | undefined;
  readonly headquarterAddress: string | null | undefined;
  readonly horizontalLogoUrl: string | null | undefined;
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly websiteUrl: string | null | undefined;
  readonly " $fragmentType": "OrganizationFormFragment";
};
export type OrganizationFormFragment$key = {
  readonly " $data"?: OrganizationFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationFormFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationFormFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "name",
        "storageKey": null
      },
      "action": "THROW"
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
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "0894093a550e5ba033a33b255b1bf2c7";

export default node;
