/**
 * @generated SignedSource<<d2d531924e3667658020b431850c193e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ElectronicSignatureEventType = "CERTIFICATE_GENERATED" | "CONSENT_GIVEN" | "DOCUMENT_VIEWED" | "FULL_NAME_TYPED" | "SEAL_COMPUTED" | "SIGNATURE_ACCEPTED" | "SIGNATURE_COMPLETED" | "TIMESTAMP_REQUESTED";
export type RecordSigningEventInput = {
  eventType: ElectronicSignatureEventType;
  signatureId: string;
};
export type NDAPageRecordSigningEventMutation$variables = {
  input: RecordSigningEventInput;
};
export type NDAPageRecordSigningEventMutation$data = {
  readonly recordSigningEvent: {
    readonly success: boolean;
  } | null | undefined;
};
export type NDAPageRecordSigningEventMutation = {
  response: NDAPageRecordSigningEventMutation$data;
  variables: NDAPageRecordSigningEventMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "RecordSigningEventPayload",
    "kind": "LinkedField",
    "name": "recordSigningEvent",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "success",
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "NDAPageRecordSigningEventMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NDAPageRecordSigningEventMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "72aa3c3ec642d5bbb317a6e2a65703de",
    "id": null,
    "metadata": {},
    "name": "NDAPageRecordSigningEventMutation",
    "operationKind": "mutation",
    "text": "mutation NDAPageRecordSigningEventMutation(\n  $input: RecordSigningEventInput!\n) {\n  recordSigningEvent(input: $input) {\n    success\n  }\n}\n"
  }
};
})();

(node as any).hash = "ec40eb438c6d6b20f31974e099535887";

export default node;
