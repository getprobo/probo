/**
 * @generated SignedSource<<a87fd05a4f80d2d7094647468d695967>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ViewerMembershipDropdownFragment$data = {
  readonly viewerMembership: {
    readonly identity: {
      readonly canListAPIKeys: boolean;
      readonly email: string;
    };
    readonly profile: {
      readonly fullName: string;
    };
  };
  readonly " $fragmentType": "ViewerMembershipDropdownFragment";
};
export type ViewerMembershipDropdownFragment$key = {
  readonly " $data"?: ViewerMembershipDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ViewerMembershipDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "ViewerMembershipDropdownFragment",
  "selections": [
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
                },
                {
                  "alias": "canListAPIKeys",
                  "args": [
                    {
                      "kind": "Literal",
                      "name": "action",
                      "value": "iam:personal-api-key:list"
                    }
                  ],
                  "kind": "ScalarField",
                  "name": "permission",
                  "storageKey": "permission(action:\"iam:personal-api-key:list\")"
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

(node as any).hash = "82c179a1ec2b90ce5f12c0959f97150f";

export default node;
