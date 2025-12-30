/**
 * @generated SignedSource<<66b5754d2d4ecfc54f6c698df247e241>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MemberListItem_permissionsFragment$data = {
  readonly canDeleteMembership: boolean;
  readonly canUpdateMembership: boolean;
  readonly " $fragmentType": "MemberListItem_permissionsFragment";
};
export type MemberListItem_permissionsFragment$key = {
  readonly " $data"?: MemberListItem_permissionsFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MemberListItem_permissionsFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
};
return {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "organizationId"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "MemberListItem_permissionsFragment",
  "selections": [
    {
      "alias": "canUpdateMembership",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:membership:update"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canDeleteMembership",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:membership:delete"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    }
  ],
  "type": "Identity",
  "abstractKey": null
};
})();

(node as any).hash = "d467db6a9f7f80521cf35c8c1a0c3416";

export default node;
