/**
 * @generated SignedSource<<2f595b4efd5e7b13207d433830a64c98>>
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
    readonly id: string;
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
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Audit",
  "abstractKey": null
};
})();

(node as any).hash = "3aabed006b65dc913c81a63f0f6f4523";

export default node;
