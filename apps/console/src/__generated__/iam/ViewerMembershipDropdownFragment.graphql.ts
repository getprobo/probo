/**
 * @generated SignedSource<<d0e62302bc1d00dffb0c61f56777472b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ViewerMembershipDropdownFragment$data = {
  readonly viewer: {
    readonly fullName: string;
    readonly identity: {
      readonly canListAPIKeys: boolean;
      readonly email: string;
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
        "concreteType": "Profile",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "fullName",
            "storageKey": null
          },
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

(node as any).hash = "fd5df361935e3e2bf4eaac1b2b4c5528";

export default node;
