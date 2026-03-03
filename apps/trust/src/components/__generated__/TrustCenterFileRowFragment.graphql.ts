/**
 * @generated SignedSource<<bd09e92db1b1b5bf9c45145f4058e88f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterFileRowFragment$data = {
  readonly access: {
    readonly id: string;
    readonly status: DocumentAccessStatus;
  } | null | undefined;
  readonly id: string;
  readonly isUserAuthorized: boolean;
  readonly name: string;
  readonly " $fragmentType": "TrustCenterFileRowFragment";
};
export type TrustCenterFileRowFragment$key = {
  readonly " $data"?: TrustCenterFileRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFileRowFragment">;
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
  "name": "TrustCenterFileRowFragment",
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
      "name": "isUserAuthorized",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "DocumentAccess",
      "kind": "LinkedField",
      "name": "access",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "status",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "TrustCenterFile",
  "abstractKey": null
};
})();

(node as any).hash = "98e5e399ab6d45c22cd85d663cfb3748";

export default node;
