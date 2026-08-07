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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'AI System ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the AI system to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the AI system',
	},
	{
		displayName: 'Status',
		name: 'status',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['update'],
			},
		},
		options: [
			{ name: '(Unchanged)', value: '' },
			{ name: 'Active', value: 'ACTIVE' },
			{ name: 'In Development', value: 'IN_DEVELOPMENT' },
			{ name: 'Decommissioned', value: 'DECOMMISSIONED' },
		],
		default: '',
		description: 'The status of the AI system',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['aiSystem'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Autonomy Level',
				name: 'autonomyLevel',
				type: 'string',
				default: '',
				description: 'The autonomy level. Empty string clears the value.',
			},
			{
				displayName: 'Company Roles',
				name: 'companyRoles',
				type: 'string',
				default: '',
				description: 'Comma-separated company roles (PROVIDER, DEPLOYER, USER, DEVELOPER)',
			},
			{
				displayName: 'Data Sources And Type',
				name: 'dataSourcesAndType',
				type: 'string',
				default: '',
				description: 'Data sources and types. Empty string clears the value.',
			},
			{
				displayName: 'Deployment Date',
				name: 'deploymentDate',
				type: 'string',
				default: '',
				description: 'Deployment date (ISO 8601 datetime). Empty string clears the value.',
			},
			{
				displayName: 'Human Oversight Mechanism',
				name: 'humanOversightMechanism',
				type: 'string',
				default: '',
				description: 'Human oversight mechanism. Empty string clears the value.',
			},
			{
				displayName: 'Intended Use Cases',
				name: 'intendedUseCases',
				type: 'string',
				default: '',
				description: 'Intended use cases. Empty string clears the value.',
			},
			{
				displayName: 'Key Stakeholders',
				name: 'keyStakeholders',
				type: 'string',
				default: '',
				description: 'Key stakeholders. Empty string clears the value.',
			},
			{
				displayName: 'Last Review Date',
				name: 'lastReviewDate',
				type: 'string',
				default: '',
				description: 'Last review date (ISO 8601 datetime). Empty string clears the value.',
			},
			{
				displayName: 'Next Review Date',
				name: 'nextReviewDate',
				type: 'string',
				default: '',
				description: 'Next review date (ISO 8601 datetime). Empty string clears the value.',
			},
			{
				displayName: 'Notes',
				name: 'notes',
				type: 'string',
				default: '',
				description: 'Additional notes. Empty string clears the value.',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'Owner profile ID. Empty string clears the value.',
			},
			{
				displayName: 'Purpose',
				name: 'purpose',
				type: 'string',
				default: '',
				description: 'The purpose. Empty string clears the value.',
			},
			{
				displayName: 'Risk Classification',
				name: 'riskClassification',
				type: 'options',
				options: [
					{ name: '(Clear)', value: '__CLEAR__' },
					{ name: '(Unchanged)', value: '' },
					{ name: 'GPAI', value: 'GPAI' },
					{ name: 'High Risk', value: 'HIGH_RISK' },
					{ name: 'Limited', value: 'LIMITED' },
					{ name: 'Minimal', value: 'MINIMAL' },
				],
				default: '',
				description: 'The risk classification of the AI system',
			},
			{
				displayName: 'Source',
				name: 'source',
				type: 'string',
				default: '',
				description: 'The source. Empty string clears the value.',
			},
			{
				displayName: 'Version',
				name: 'version',
				type: 'string',
				default: '',
				description: 'The version. Empty string clears the value.',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const status = this.getNodeParameter('status', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		version?: string;
		companyRoles?: string;
		ownerId?: string;
		source?: string;
		purpose?: string;
		intendedUseCases?: string;
		autonomyLevel?: string;
		humanOversightMechanism?: string;
		riskClassification?: string;
		keyStakeholders?: string;
		dataSourcesAndType?: string;
		deploymentDate?: string;
		lastReviewDate?: string;
		nextReviewDate?: string;
		notes?: string;
	};

	const query = `
		mutation UpdateAiSystem($input: UpdateAiSystemInput!) {
			updateAiSystem(input: $input) {
				aiSystem {
					id
					name
					version
					companyRoles
					status
					source
					purpose
					intendedUseCases
					autonomyLevel
					humanOversightMechanism
					riskClassification
					keyStakeholders
					dataSourcesAndType
					deploymentDate
					lastReviewDate
					nextReviewDate
					notes
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id };

	if (name) {
		input.name = name;
	}

	if (status) {
		input.status = status;
	}

	if (additionalFields.version !== undefined) {
		input.version = additionalFields.version === '' ? null : additionalFields.version;
	}

	if (additionalFields.companyRoles) {
		input.companyRoles = additionalFields.companyRoles
			.split(',')
			.map((role) => role.trim())
			.filter(Boolean);
	}

	if (additionalFields.ownerId !== undefined) {
		input.ownerId = additionalFields.ownerId === '' ? null : additionalFields.ownerId;
	}

	if (additionalFields.source !== undefined) {
		input.source = additionalFields.source === '' ? null : additionalFields.source;
	}

	if (additionalFields.purpose !== undefined) {
		input.purpose = additionalFields.purpose === '' ? null : additionalFields.purpose;
	}

	if (additionalFields.intendedUseCases !== undefined) {
		input.intendedUseCases =
			additionalFields.intendedUseCases === '' ? null : additionalFields.intendedUseCases;
	}

	if (additionalFields.autonomyLevel !== undefined) {
		input.autonomyLevel =
			additionalFields.autonomyLevel === '' ? null : additionalFields.autonomyLevel;
	}

	if (additionalFields.humanOversightMechanism !== undefined) {
		input.humanOversightMechanism =
			additionalFields.humanOversightMechanism === ''
				? null
				: additionalFields.humanOversightMechanism;
	}

	if (additionalFields.riskClassification) {
		input.riskClassification =
			additionalFields.riskClassification === '__CLEAR__'
				? null
				: additionalFields.riskClassification;
	}

	if (additionalFields.keyStakeholders !== undefined) {
		input.keyStakeholders =
			additionalFields.keyStakeholders === '' ? null : additionalFields.keyStakeholders;
	}

	if (additionalFields.dataSourcesAndType !== undefined) {
		input.dataSourcesAndType =
			additionalFields.dataSourcesAndType === '' ? null : additionalFields.dataSourcesAndType;
	}

	if (additionalFields.deploymentDate !== undefined) {
		input.deploymentDate =
			additionalFields.deploymentDate === '' ? null : additionalFields.deploymentDate;
	}

	if (additionalFields.lastReviewDate !== undefined) {
		input.lastReviewDate =
			additionalFields.lastReviewDate === '' ? null : additionalFields.lastReviewDate;
	}

	if (additionalFields.nextReviewDate !== undefined) {
		input.nextReviewDate =
			additionalFields.nextReviewDate === '' ? null : additionalFields.nextReviewDate;
	}

	if (additionalFields.notes !== undefined) {
		input.notes = additionalFields.notes === '' ? null : additionalFields.notes;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
