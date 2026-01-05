/**
 * @generated SignedSource<<73d5a0059348fd0c5950ae3a873cc383>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SessionDropdownFragment$data = {
  readonly canDelete: boolean;
  readonly viewerMembership: {
    readonly identity: {
      readonly email: string;
    };
    readonly profile: {
      readonly fullName: string;
    };
  };
  readonly " $fragmentType": "SessionDropdownFragment";
};
export type SessionDropdownFragment$key = {
  readonly " $data"?: SessionDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SessionDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SessionDropdownFragment",
  "selections": [
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:organization:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:organization:delete\")"
    },
    {
      "kind": "RequiredField",
      "field": {
        "alias": null,
        "args": null,
        "concreteType": "Membership",
        "kind": "LinkedField",
        "name": "viewerMembership",
        "plural": false,
        "selections": [
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "concreteType": "Identity",
              "kind": "LinkedField",
              "name": "identity",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "email",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            "action": "THROW"
          },
          {
            "kind": "RequiredField",
            "field": {
              "alias": null,
              "args": null,
              "concreteType": "MembershipProfile",
              "kind": "LinkedField",
              "name": "profile",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "fullName",
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            "action": "THROW"
          }
        ],
        "storageKey": null
      },
      "action": "THROW"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "51810eafbf165e7b5eeec42cb268ae54";

export default node;
