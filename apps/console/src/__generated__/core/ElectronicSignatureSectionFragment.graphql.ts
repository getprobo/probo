/**
 * @generated SignedSource<<8ef79d0009e7493acec0b0dd749911a2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ElectronicSignatureEventType = "CERTIFICATE_GENERATED" | "CONSENT_GIVEN" | "DOCUMENT_VIEWED" | "FULL_NAME_TYPED" | "PROCESSING_ERROR" | "SEAL_COMPUTED" | "SIGNATURE_ACCEPTED" | "SIGNATURE_COMPLETED" | "TIMESTAMP_REQUESTED";
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
import { FragmentRefs } from "relay-runtime";
export type ElectronicSignatureSectionFragment$data = {
  readonly certificateFileUrl: string | null | undefined;
  readonly events: ReadonlyArray<{
    readonly actorEmail: string;
    readonly eventType: ElectronicSignatureEventType;
    readonly id: string;
    readonly occurredAt: string;
  }>;
  readonly signedAt: string | null | undefined;
  readonly status: ElectronicSignatureStatus;
  readonly " $fragmentType": "ElectronicSignatureSectionFragment";
};
export type ElectronicSignatureSectionFragment$key = {
  readonly " $data"?: ElectronicSignatureSectionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ElectronicSignatureSectionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "ElectronicSignatureSectionFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "signedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "certificateFileUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "ElectronicSignatureEvent",
      "kind": "LinkedField",
      "name": "events",
      "plural": true,
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
          "name": "eventType",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "actorEmail",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "occurredAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "ElectronicSignature",
  "abstractKey": null
};

(node as any).hash = "70f3c5dd8a387d419c5ebdf752908162";

export default node;
