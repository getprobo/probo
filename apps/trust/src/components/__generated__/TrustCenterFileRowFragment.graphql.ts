/**
 * @generated SignedSource<<0261adc72d21035b906b6156ea9d0b40>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustCenterFileRowFragment$data = {
  readonly hasUserRequestedAccess: boolean;
  readonly id: string;
  readonly isUserAuthorized: boolean;
  readonly name: string;
  readonly " $fragmentType": "TrustCenterFileRowFragment";
};
export type TrustCenterFileRowFragment$key = {
  readonly " $data"?: TrustCenterFileRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFileRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TrustCenterFileRowFragment",
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
      "name": "isUserAuthorized",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasUserRequestedAccess",
      "storageKey": null
    }
  ],
  "type": "TrustCenterFile",
  "abstractKey": null
};

(node as any).hash = "4ff11c14e65566f0dfe3a77ea73903e8";

export default node;
