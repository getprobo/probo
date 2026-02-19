/**
 * @generated SignedSource<<79146bf2cc4dc6f1cc6ea9ba41b26641>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ReportRowFragment$data = {
  readonly file: {
    readonly filename: string;
    readonly hasUserRequestedAccess: boolean;
    readonly isUserAuthorized: boolean;
  } | null | undefined;
  readonly framework: {
    readonly darkLogoURL: string | null | undefined;
    readonly id: string;
    readonly lightLogoURL: string | null | undefined;
    readonly name: string;
  };
  readonly frameworkType: string | null | undefined;
  readonly id: string;
  readonly " $fragmentType": "ReportRowFragment";
};
export type ReportRowFragment$key = {
  readonly " $data"?: ReportRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ReportRowFragment">;
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
  "name": "ReportRowFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "frameworkType",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "ReportFile",
      "kind": "LinkedField",
      "name": "file",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "filename",
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
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": [
        (v0/*: any*/),
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
          "name": "lightLogoURL",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "darkLogoURL",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Report",
  "abstractKey": null
};
})();

(node as any).hash = "aa421c44d432b45d8c3da39e6cc87afc";

export default node;
