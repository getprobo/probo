/**
 * @generated SignedSource<<dfd688860df6b21ffe3a3f14bb2e2bf1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SCIMConfigurationFragment$data = {
  readonly canCreateSCIMConfiguration: boolean;
  readonly canDeleteSCIMConfiguration: boolean;
  readonly scimConfiguration: {
    readonly bridge: {
      readonly id: string;
    } | null | undefined;
    readonly endpointUrl: string;
    readonly id: string;
  } | null | undefined;
  readonly " $fragmentType": "SCIMConfigurationFragment";
};
export type SCIMConfigurationFragment$key = {
  readonly " $data"?: SCIMConfigurationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SCIMConfigurationFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SCIMConfigurationFragment",
  "selections": [
    {
      "alias": "canCreateSCIMConfiguration",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:scim-configuration:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:scim-configuration:create\")"
    },
    {
      "alias": "canDeleteSCIMConfiguration",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:scim-configuration:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:scim-configuration:delete\")"
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "SCIMConfiguration",
      "kind": "LinkedField",
      "name": "scimConfiguration",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "endpointUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "SCIMBridge",
          "kind": "LinkedField",
          "name": "bridge",
          "plural": false,
          "selections": [
            (v0/*: any*/)
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};
})();

(node as any).hash = "5b128ae2e680a0ffe0dd931e4e86f52e";

export default node;
