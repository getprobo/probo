/**
 * @generated SignedSource<<889efcfe178bb170887494345d05c48b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ProfileState = "ACTIVE" | "INACTIVE";
import { FragmentRefs } from "relay-runtime";
export type MembershipCardFragment$data = {
  readonly membership: {
    readonly lastSession: {
      readonly expiresAt: string;
      readonly id: string;
    } | null | undefined;
  };
  readonly state: ProfileState;
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
      "kind": "ScalarField",
      "name": "state",
      "storageKey": null
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Membership",
        "kind": "LinkedField",
        "name": "membership",
        "plural": false,
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
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Profile",
  "abstractKey": null
};

(node as any).hash = "2e4f6ddee75ca32542f959c44de58e25";

export default node;
