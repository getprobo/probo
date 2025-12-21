/**
 * @generated SignedSource<<ccb8e318151fff4495fb2eaacc671612>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SessionDropdownFragment$data = {
  readonly email: any;
  readonly profileFor: {
    readonly firstName: string | null | undefined;
    readonly lastName: string | null | undefined;
  } | null | undefined;
  readonly " $fragmentType": "SessionDropdownFragment";
};
export type SessionDropdownFragment$key = {
  readonly " $data"?: SessionDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SessionDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "organizationId"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "SessionDropdownFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "email",
      "storageKey": null
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Variable",
          "name": "organizationId",
          "variableName": "organizationId"
        }
      ],
      "concreteType": "IdentityProfile",
      "kind": "LinkedField",
      "name": "profileFor",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "firstName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "lastName",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Identity",
  "abstractKey": null
};

(node as any).hash = "de1d0f069e5648898cab6a56a394d86f";

export default node;
