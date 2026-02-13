/**
 * @generated SignedSource<<f6caf8d14eaaea74987cd5f91fedaed3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipCardFragment$data = {
  readonly lastSession: {
    readonly expiresAt: string;
    readonly id: string;
  } | null | undefined;
  readonly " $fragmentType": "MembershipCardFragment";
};
export type MembershipCardFragment$key = {
  readonly " $data"?: MembershipCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MembershipCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembershipCardFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Session",
      "kind": "LinkedField",
      "name": "lastSession",
      "plural": false,
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
          "name": "expiresAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Membership",
  "abstractKey": null
};

(node as any).hash = "3da5f340850d352daaa553ca7b78aa37";

export default node;
