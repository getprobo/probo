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

import type { INodeProperties } from 'n8n-workflow';

export const incidentFields = `
	id
	title
	description
	occurredAt
	discoveredAt
	awarenessAt
	affectedDataSubjects
	affectedDataRecords
	personalDataTypes
	affectedSystem
	likelyConsequences
	containmentActions
	potentialPhysicalHarm
	potentialFinancialLoss
	potentialCreditOrPropertyDamage
	potentialIllegalUse
	sensitivePersonalData
	potentialIdentityFraud
	significantHarm
	significantScale
	notificationRecommendation
	notificationReasons
	notificationDecision
	decisionRationale
	decisionEvidence
	assessedByProfileId
	assessedAt
	ruleVersion
	ruleSource
	commissionerNotificationDueAt
	commissionerNotificationOverdue
	commissionerNotifiedAt
	commissionerNotificationReference
	commissionerConfirmationReceivedAt
	commissionerConfirmationReference
	phasedInformationDueAt
	delayedNotificationReason
	delayedNotificationEvidence
	dataSubjectsNotificationDueAt
	dataSubjectsNotificationOverdue
	dataSubjectsNotifiedAt
	dataSubjectsNotificationEvidence
	status
	createdByProfileId
	createdAt
	updatedAt
`;

export function incidentIdField(operation: string): INodeProperties {
	return {
		displayName: 'Incident ID',
		name: 'incidentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['malaysiaPDPABreach'],
				operation: [operation],
			},
		},
		default: '',
		description: 'The ID of the breach incident',
		required: true,
	};
}

export const decisionOptions = [
	{ name: 'Pending', value: 'PENDING' },
	{ name: 'Not Required', value: 'NOT_REQUIRED' },
	{ name: 'Commissioner Only', value: 'COMMISSIONER_ONLY' },
	{
		name: 'Commissioner and Data Subjects',
		value: 'COMMISSIONER_AND_DATA_SUBJECTS',
	},
];
