// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { IDataObject, IExecuteFunctions, INodeProperties } from 'n8n-workflow';

const basisOptions = [
	{ name: 'Substantially Similar Law — Section 129(2)(a)', value: 'SUBSTANTIALLY_SIMILAR_LAW' },
	{
		name: 'Adequate Equivalent Protection — Section 129(2)(b)',
		value: 'ADEQUATE_EQUIVALENT_PROTECTION',
	},
	{ name: 'Data Subject Consent — Section 129(3)(a)', value: 'DATA_SUBJECT_CONSENT' },
	{ name: 'Data Subject Contract — Section 129(3)(b)', value: 'DATA_SUBJECT_CONTRACT' },
	{ name: 'Third-Party Contract — Section 129(3)(c)', value: 'THIRD_PARTY_CONTRACT' },
	{ name: 'Legal Proceedings or Rights — Section 129(3)(d)', value: 'LEGAL_PROCEEDINGS' },
	{ name: 'Adverse Action — Section 129(3)(e)', value: 'ADVERSE_ACTION' },
	{ name: 'Reasonable Precautions — Section 129(3)(f)', value: 'REASONABLE_PRECAUTIONS' },
	{ name: 'Vital Interests — Section 129(3)(g)', value: 'VITAL_INTERESTS' },
];

export function malaysiaTransferProperties(operation: 'create' | 'update'): INodeProperties[] {
	const show = { resource: ['tia'], operation: [operation] };
	const malaysiaShow = { ...show, includeMalaysiaPDPA: [true] };

	return [
		{
			displayName: 'Include Malaysia PDPA Transfer Record',
			name: 'includeMalaysiaPDPA',
			type: 'boolean',
			displayOptions: { show },
			default: false,
			description: 'Whether to record the section 129 basis, recipient, safeguards, and approval',
		},
		{
			displayName: 'Malaysia Transfer Basis',
			name: 'malaysiaBasis',
			type: 'options',
			displayOptions: { show: malaysiaShow },
			options: basisOptions,
			default: 'SUBSTANTIALLY_SIMILAR_LAW',
			required: true,
		},
		{
			displayName: 'Destination Country Code',
			name: 'malaysiaDestinationCountry',
			type: 'string',
			displayOptions: { show: malaysiaShow },
			default: '',
			description: 'Foreign ISO country code; MY and GLOBAL are not accepted',
			required: true,
		},
		{
			displayName: 'Recipient Third-Party ID',
			name: 'malaysiaRecipientThirdPartyId',
			type: 'string',
			displayOptions: { show: malaysiaShow },
			default: '',
			required: true,
		},
		{
			displayName: 'Receiver Registration Number',
			name: 'malaysiaReceiverRegistrationNumber',
			type: 'string',
			displayOptions: { show: malaysiaShow },
			default: '',
		},
		{
			displayName: 'Receiver DPO or Contact',
			name: 'malaysiaReceiverContact',
			type: 'string',
			displayOptions: { show: malaysiaShow },
			default: '',
			required: true,
		},
		{
			displayName: 'Transfer Purpose',
			name: 'malaysiaTransferPurpose',
			type: 'string',
			typeOptions: { rows: 3 },
			displayOptions: { show: malaysiaShow },
			default: '',
			required: true,
		},
		{
			displayName: 'Personal Data Categories',
			name: 'malaysiaPersonalDataCategories',
			type: 'string',
			typeOptions: { rows: 3 },
			displayOptions: { show: malaysiaShow },
			default: '',
			required: true,
		},
		{
			displayName: 'Safeguards and Evidence',
			name: 'malaysiaSafeguards',
			type: 'string',
			typeOptions: { rows: 3 },
			displayOptions: { show: malaysiaShow },
			default: '',
			required: true,
		},
		{
			displayName: 'Approval Status',
			name: 'malaysiaApprovalStatus',
			type: 'options',
			displayOptions: { show: malaysiaShow },
			options: [
				{ name: 'Pending', value: 'PENDING' },
				{ name: 'Approved', value: 'APPROVED' },
				{ name: 'Rejected', value: 'REJECTED' },
			],
			default: 'PENDING',
			required: true,
		},
		{
			displayName: 'Approval Notes',
			name: 'malaysiaApprovalNotes',
			type: 'string',
			typeOptions: { rows: 3 },
			displayOptions: { show: malaysiaShow },
			default: '',
			description: 'Required when the transfer is rejected',
		},
		{
			displayName: 'Review Evidence',
			name: 'malaysiaReviewEvidence',
			type: 'string',
			typeOptions: { rows: 3 },
			displayOptions: { show: malaysiaShow },
			default: '',
			description: 'Required when the transfer is approved',
		},
	];
}

export function getMalaysiaTransferInput(
	execute: IExecuteFunctions,
	itemIndex: number,
): IDataObject | undefined {
	if (!(execute.getNodeParameter('includeMalaysiaPDPA', itemIndex, false) as boolean)) {
		return undefined;
	}

	return {
		basis: execute.getNodeParameter('malaysiaBasis', itemIndex) as string,
		destinationCountry: execute.getNodeParameter('malaysiaDestinationCountry', itemIndex) as string,
		recipientThirdPartyId: execute.getNodeParameter(
			'malaysiaRecipientThirdPartyId',
			itemIndex,
		) as string,
		receiverRegistrationNumber: execute.getNodeParameter(
			'malaysiaReceiverRegistrationNumber',
			itemIndex,
			'',
		) as string,
		receiverContact: execute.getNodeParameter('malaysiaReceiverContact', itemIndex) as string,
		transferPurpose: execute.getNodeParameter('malaysiaTransferPurpose', itemIndex) as string,
		personalDataCategories: execute.getNodeParameter(
			'malaysiaPersonalDataCategories',
			itemIndex,
		) as string,
		safeguards: execute.getNodeParameter('malaysiaSafeguards', itemIndex) as string,
		approvalStatus: execute.getNodeParameter('malaysiaApprovalStatus', itemIndex) as string,
		approvalNotes: execute.getNodeParameter('malaysiaApprovalNotes', itemIndex, '') as string,
		reviewEvidence: execute.getNodeParameter('malaysiaReviewEvidence', itemIndex, '') as string,
	};
}
