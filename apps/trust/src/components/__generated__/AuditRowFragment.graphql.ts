/**
 * @generated SignedSource<<9540bd74ce3b0dd5366c237b56070d9f>>
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
  readonly name: string | null | undefined;
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
  "name": "name",
  "storageKey": null
},
v1 = {
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
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "Report",
      "kind": "LinkedField",
      "name": "report",
      "plural": false,
      "selections": [
        (v1/*: any*/),
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
        (v1/*: any*/),
        (v0/*: any*/),
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

(node as any).hash = "ebf47a97014ba4dad0da94f2bb01666f";

export default node;
