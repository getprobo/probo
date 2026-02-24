/**
 * @generated SignedSource<<86e7c2dc76fe7f187d4d72908c54875d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type AccessEntryAuthMethod = "API_KEY" | "PASSWORD" | "SERVICE_ACCOUNT" | "SSO" | "UNKNOWN";
export type AccessEntryDecision = "APPROVED" | "DEFER" | "ESCALATE" | "MODIFY" | "PENDING" | "REVOKE";
export type AccessEntryFlag = "EXCESSIVE" | "INACTIVE" | "NEW" | "NONE" | "ORPHANED" | "ROLE_MISMATCH";
export type AccessEntryIncrementalTag = "NEW" | "REMOVED" | "UNCHANGED";
export type MfaStatus = "DISABLED" | "ENABLED" | "UNKNOWN";
import { FragmentRefs } from "relay-runtime";
export type AccessEntryRowFragment$data = {
  readonly authMethod: AccessEntryAuthMethod;
  readonly canDecide: boolean;
  readonly decision: AccessEntryDecision;
  readonly decisionNote: string | null | undefined;
  readonly email: string;
  readonly flag: AccessEntryFlag;
  readonly fullName: string;
  readonly id: string;
  readonly incrementalTag: AccessEntryIncrementalTag;
  readonly mfaStatus: MfaStatus;
  readonly role: string;
  readonly " $fragmentType": "AccessEntryRowFragment";
};
export type AccessEntryRowFragment$key = {
  readonly " $data"?: AccessEntryRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"AccessEntryRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "AccessEntryRowFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "email",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fullName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "role",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "flag",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "decision",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "decisionNote",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "incrementalTag",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "mfaStatus",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "authMethod",
      "storageKey": null
    },
    {
      "alias": "canDecide",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:access-entry:decide"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:access-entry:decide\")"
    }
  ],
  "type": "AccessEntry",
  "abstractKey": null
};

(node as any).hash = "1c4d1d5476b1f9ab833821127966718e";

export default node;
