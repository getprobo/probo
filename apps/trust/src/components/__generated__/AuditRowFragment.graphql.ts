/**
 * @generated SignedSource<<3265a0991323eee4011c668b68d1b365>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AuditRowFragment$data = {
  readonly framework: {
    readonly darkLogoURL: string | null | undefined;
    readonly id: string;
    readonly lightLogoURL: string | null | undefined;
    readonly name: string;
  };
  readonly report: {
    readonly filename: string;
    readonly hasUserRequestedAccess: boolean;
    readonly id: string;
    readonly isUserAuthorized: boolean;
  } | null | undefined;
  readonly " $fragmentType": "AuditRowFragment";
};
export type AuditRowFragment$key = {
  readonly " $data"?: AuditRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"AuditRowFragment">;
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
  "name": "AuditRowFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Report",
      "kind": "LinkedField",
      "name": "report",
      "plural": false,
      "selections": [
        (v0/*: any*/),
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
  "type": "Audit",
  "abstractKey": null
};
})();

(node as any).hash = "417e65b500df4b4cff874dab2bea90ee";

export default node;
