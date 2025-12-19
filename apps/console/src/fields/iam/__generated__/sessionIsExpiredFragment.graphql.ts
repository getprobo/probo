/**
 * @generated SignedSource<<b6a869ebbc0ded2da70f59b7856b906f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type sessionIsExpiredFragment$data = {
  readonly expiresAt: any;
  readonly " $fragmentType": "sessionIsExpiredFragment";
};
export type sessionIsExpiredFragment$key = {
  readonly " $data"?: sessionIsExpiredFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"sessionIsExpiredFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "sessionIsExpiredFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "expiresAt",
      "storageKey": null
    }
  ],
  "type": "Session",
  "abstractKey": null
};

(node as any).hash = "ef7bc8bbfa6d2573b0df09b5b13b00b2";

export default node;
